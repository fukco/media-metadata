package mp4

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"
)

// BoxInfo has common information of box
type BoxInfo struct {
	// Offset specifies an offset of the box in a file.
	Offset uint64

	// Size specifies size(bytes) of box.
	Size uint64

	// HeaderSize specifies size(bytes) of common fields which are defined as "Box" class member at ISO/IEC 14496-12.
	HeaderSize uint64

	// Type specifies box type which is represented by 4 characters.
	Type BoxType

	// ExtendedType specifies box extended type which is represented by 16 characters.
	ExtendedType [16]byte
}

const (
	defaultHeaderSize = 8
)

// ReadBoxInfo reads common fields which are defined as "Box" class member at ISO/IEC 14496-12.
func ReadBoxInfo(r io.ReadSeeker) (*BoxInfo, error) {
	offset, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	bi := &BoxInfo{
		Offset: uint64(offset),
	}

	// read 8 bytes
	buf := bytes.NewBuffer(make([]byte, 0, defaultHeaderSize))
	if _, err := io.CopyN(buf, r, defaultHeaderSize); err != nil {
		return nil, err
	}
	bi.HeaderSize = defaultHeaderSize

	// pick size and type
	data := buf.Bytes()
	bi.Size = uint64(binary.BigEndian.Uint32(buf.Bytes()[:4]))
	bi.Type = BoxType{data[4], data[5], data[6], data[7]}

	if bi.Size == 1 {
		// read more 8 bytes
		buf.Reset()
		if _, err := io.CopyN(buf, r, 8); err != nil {
			return nil, err
		}
		bi.HeaderSize += 8
		bi.Size = binary.BigEndian.Uint64(buf.Bytes())
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

	if bi.Type == StrToBoxType("uuid") {
		buf.Reset()
		if _, err := io.CopyN(buf, r, 16); err != nil {
			return nil, err
		}
		copy(bi.ExtendedType[:], buf.Bytes())
	}
	return bi, nil
}

func isFullBox(bi *BoxInfo) bool {
	if _, ok := boxMap[bi.Type]; ok {
		t := reflect.TypeOf(boxMap[bi.Type])
		for i := 0; i < t.NumField(); i++ {
			if reflect.TypeOf(FullBox{}) == t.Field(i).Type {
				return true
			}
		}
	}
	return false
}

func (bi *BoxInfo) SeekToStart(s io.Seeker) (int64, error) {
	return s.Seek(int64(bi.Offset), io.SeekStart)
}

func (bi *BoxInfo) SeekToPayload(s io.Seeker) (int64, error) {
	return s.Seek(int64(bi.Offset+bi.HeaderSize), io.SeekStart)
}

func (bi *BoxInfo) SeekToEnd(s io.Seeker) (int64, error) {
	return s.Seek(int64(bi.Offset+bi.Size), io.SeekStart)
}
