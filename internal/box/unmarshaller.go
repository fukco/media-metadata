package box

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
)

const (
	anyVersion = math.MaxUint8
)

var ErrUnsupportedBoxVersion = errors.New("unsupported box version")

func readerHasSize(reader io.ReadSeeker, size uint64) bool {
	pre, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		return false
	}

	end, err := reader.Seek(0, io.SeekEnd)
	if err != nil {
		return false
	}

	if uint64(end-pre) < size {
		return false
	}

	_, err = reader.Seek(pre, io.SeekStart)
	return err == nil
}

type unmarshaller struct {
	reader io.ReadSeeker
	dst    Boxer
	bi     *BoxInfo
	size   uint64 //byte size
	index  uint64
}

func UnmarshalAny(r io.ReadSeeker, bi *BoxInfo, ctx *Context) (box Boxer, err error) {
	dst, err := bi.New()
	if err != nil {
		return nil, err
	}
	err = Unmarshal(r, bi, dst, ctx)
	return dst, err
}

func Unmarshal(r io.ReadSeeker, info *BoxInfo, dst Boxer, ctx *Context) (err error) {
	def := info.GetBoxDef()
	if def == nil {
		return ErrBoxDefinitionNotFound
	}

	v := reflect.ValueOf(dst).Elem()

	u := &unmarshaller{
		reader: r,
		dst:    dst,
		bi:     info,
		size:   info.Size - info.HeaderSize,
	}

	//sn, err := r.Seek(0, io.SeekCurrent)
	//if err != nil {
	//	return err
	//}

	if err := u.unmarshalStruct(v, def.fields, ctx); err != nil {
		//if errors.Is(err, ErrUnsupportedBoxVersion) {
		//	_, err := r.Seek(sn, io.SeekStart)
		//	if err != nil {
		//		return err
		//	}
		//}
		return err
	}

	return nil
}

func (u *unmarshaller) unmarshal(v reflect.Value, f *field, ctx *Context) error {
	var err error
	switch v.Type().Kind() {
	case reflect.Ptr:
		err = u.unmarshalPtr(v, f, ctx)
	case reflect.Struct:
		err = u.unmarshalStructInternal(v, f, ctx)
	case reflect.Array:
		err = u.unmarshalArray(v, f, ctx)
	case reflect.Slice:
		err = u.unmarshalSlice(v, f, ctx)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = u.unmarshalInt(v, f)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		err = u.unmarshalUint(v, f)
	//case reflect.Bool:
	//	err = u.unmarshalBool(v, f)
	case reflect.String:
		err = u.unmarshalString(v, f)
	default:
		return fmt.Errorf("unsupported type: %s", v.Type().Kind())
	}
	return err
}

func (u *unmarshaller) unmarshalPtr(v reflect.Value, fi *field, ctx *Context) error {
	v.Set(reflect.New(v.Type().Elem()))
	return u.unmarshal(v.Elem(), fi, ctx)
}

func (u *unmarshaller) unmarshalStructInternal(v reflect.Value, fi *field, ctx *Context) error {
	if fi.size != 0 && fi.size%8 == 0 {
		u2 := *u
		u2.size = uint64(fi.size / 8)
		u2.index = 0
		if err := u2.unmarshalStruct(v, fi.children, ctx); err != nil {
			return err
		}
		u.index += u2.index
		if u2.index != uint64(fi.size/8) {
			return errors.New("invalid alignment")
		}
		return nil
	}

	return u.unmarshalStruct(v, fi.children, ctx)
}

func (u *unmarshaller) unmarshalStruct(v reflect.Value, fs []*field, ctx *Context) error {
	if u.bi.FullBoxHeader != nil && !u.bi.IsSupportedVersion(u.bi.Version) {
		return ErrUnsupportedBoxVersion
	}

	for _, f := range fs {
		resolveFieldInstance(f, u.dst, v, ctx)
		if !isTargetField(u.bi, f) {
			continue
		}

		err := u.unmarshal(v.FieldByName(f.name), f, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u *unmarshaller) unmarshalArray(v reflect.Value, f *field, ctx *Context) error {
	for i := 0; i < v.Len(); i++ {
		err := u.unmarshal(v.Index(i), f, ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *unmarshaller) unmarshalSlice(v reflect.Value, f *field, ctx *Context) error {
	var slice reflect.Value
	elemType := v.Type().Elem()

	length := uint64(f.length)
	if f.length == LengthUnlimited {
		if f.size != 0 {
			left := (u.size - u.index) * 8
			if left%uint64(f.size) != 0 {
				return errors.New("invalid alignment")
			}
			length = left / uint64(f.size)
		} else {
			length = 0
		}
	}

	if elemType.Kind() == reflect.Uint8 && f.size == 8 {
		totalSize := length * uint64(f.size) / 8

		if !readerHasSize(u.reader, totalSize) {
			return fmt.Errorf("not enough bits")
		}

		buf := bytes.NewBuffer(make([]byte, 0, totalSize))
		if _, err := io.CopyN(buf, u.reader, int64(totalSize)); err != nil {
			return err
		}
		slice = reflect.ValueOf(buf.Bytes())
		u.index += totalSize

	} else {
		slice = reflect.MakeSlice(v.Type(), 0, 0)
		for i := 0; ; i++ {
			if f.length != LengthUnlimited && uint(i) >= f.length {
				break
			}
			if f.length == LengthUnlimited && u.index >= u.size {
				break
			}
			slice = reflect.Append(slice, reflect.Zero(elemType))
			if err := u.unmarshal(slice.Index(i), f, ctx); err != nil {
				return err
			}
			if u.index > u.size {
				return fmt.Errorf("failed to read array completely: fieldName=\"%s\"", f.name)
			}
		}
	}

	v.Set(slice)
	return nil
}

func (u *unmarshaller) unmarshalInt(v reflect.Value, fi *field) error {
	if err := checkFieldSize(fi); err != nil {
		return err
	}

	buf := make([]byte, fi.size/8)
	_, err := io.ReadFull(u.reader, buf)
	if err != nil {
		return err
	}

	u.index += uint64(fi.size / 8)

	signBit := false
	if len(buf) > 0 {
		signMask := byte(1 << 7)
		signBit = buf[0]&signMask != 0
		if signBit {
			buf[0] |= ^(signMask - 1)
		}
	}

	var val uint64
	if signBit {
		val = ^uint64(0)
	}
	for i := range buf {
		val <<= 8
		val |= uint64(buf[i])
	}
	v.SetInt(int64(val))
	return nil
}

func (u *unmarshaller) unmarshalUint(v reflect.Value, fi *field) error {
	if err := checkFieldSize(fi); err != nil {
		return err
	}

	buf := make([]byte, fi.size/8)
	_, err := io.ReadFull(u.reader, buf)
	if err != nil {
		return err
	}
	u.index += uint64(fi.size / 8)

	val := uint64(0)
	for i := range buf {
		val <<= 8
		val |= uint64(buf[i])
	}
	v.SetUint(val)

	return nil
}

func (u *unmarshaller) unmarshalString(v reflect.Value, f *field) error {
	data := make([]byte, 0, 16)
	for {
		if u.index >= u.size*8 {
			break
		}

		buf := make([]byte, 1)
		_, err := u.reader.Read(buf)
		if err != nil {
			return err
		}
		u.index += 1

		if buf[0] == 0 {
			break // null character
		}

		data = append(data, buf[0])
	}

	switch f.strType {
	case stringType_utfstring:
		v.SetString(string(data))
		return nil
	case stringType_utf8string:
		return u.unmarshalStringUtf8(v, data)
	default:
		return fmt.Errorf("unknown string type: %d", f.strType)
	}
}

func (u *unmarshaller) unmarshalStringUtf8(v reflect.Value, data []byte) error {
	if hex.EncodeToString(data) == "feff" || hex.EncodeToString(data) == "fffe" {
		//todo
		return nil
	}
	v.SetString(string(data))
	return nil
}
