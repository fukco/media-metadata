package box

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
)

// Info has common information of box
type Info struct {
	// Offset specifies an offset of the box in a file.
	Offset uint64

	// Size specifies size(bytes) of box.
	Size uint64

	// HeaderSize specifies size(bytes) of common fields which are defined as "Box" class member at ISO/IEC 14496-12.
	HeaderSize uint64

	// Type specifies box type which is represented by 4 characters.
	Type Type

	// ExtendedType specifies box extended type which is represented by 16 characters.
	ExtendedType [16]byte
}

const (
	defaultHeaderSize = 8
)

// ReadBoxInfo reads common fields which are defined as "Box" class member at ISO/IEC 14496-12.
func ReadBoxInfo(r io.ReadSeeker) (*Info, error) {
	offset, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	bi := &Info{
		Offset: uint64(offset),
	}

	// read 8 bytes
	buf := bytes.NewBuffer([]byte{})
	if _, err := io.CopyN(buf, r, defaultHeaderSize); err != nil {
		return nil, err
	}
	bi.HeaderSize = defaultHeaderSize

	// pick size and type
	data := buf.Next(defaultHeaderSize)
	bi.Size = uint64(binary.BigEndian.Uint32(data[:4]))
	bi.Type = Type{data[4], data[5], data[6], data[7]}

	if bi.Size == 1 {
		// read more 8 bytes
		if _, err := io.CopyN(buf, r, 8); err != nil {
			return nil, err
		}
		bi.HeaderSize += 8
		bi.Size = binary.BigEndian.Uint64(buf.Next(8))
	} else if bi.Size == 0 {
		// box extends to end of file
		offsetEOF, err := r.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		bi.Size = uint64(offsetEOF) - bi.Offset
		if _, err := bi.SeekToPayload(r); err != nil {
			return nil, err
		}
	}

	if bi.Type == StrToType("uuid") {
		if _, err := io.CopyN(buf, r, 16); err != nil {
			return nil, err
		}
		copy(bi.ExtendedType[:], buf.Next(16))
	}
	return bi, nil
}

func isFullBox(bi *Info) bool {
	if _, ok := Map[bi.Type]; ok {
		t := reflect.TypeOf(Map[bi.Type])
		for i := 0; i < t.NumField(); i++ {
			if reflect.TypeOf(FullBox{}) == t.Field(i).Type {
				return true
			}
		}
	}
	return false
}

func (bi *Info) SeekToStart(s io.Seeker) (int64, error) {
	return s.Seek(int64(bi.Offset), io.SeekStart)
}

func (bi *Info) SeekToPayload(s io.Seeker) (int64, error) {
	return s.Seek(int64(bi.Offset+bi.HeaderSize), io.SeekStart)
}

func (bi *Info) SeekToEnd(s io.Seeker) (int64, error) {
	return s.Seek(int64(bi.Offset+bi.Size), io.SeekStart)
}
