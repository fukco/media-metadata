package atom

import (
	"bytes"
	"encoding/binary"
	"github.com/fukco/media-meta-parser/media"
	"io"
	"reflect"
)

// Info has common information of atom
type Info struct {
	// Offset specifies an offset of the atom in a file.
	Offset uint64

	// Size specifies size(bytes) of atom.
	Size uint64

	// HeaderSize specifies size(bytes) of common fields which are defined as "Atom" class member at ISO/IEC 14496-12.
	HeaderSize uint64

	// Type specifies atom type which is represented by 4 characters.
	Type media.BoxType

	// ExtendedType specifies box extended type which is represented by 16 characters.
	ExtendedType [16]byte
}

const (
	defaultHeaderSize = 8
)

// ReadAtomInfo reads common fields which are defined as "Atom" class member at ISO/IEC 14496-12.
func ReadAtomInfo(r io.ReadSeeker) (*Info, error) {
	offset, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	ai := &Info{
		Offset: uint64(offset),
	}

	// read 8 bytes
	buf := bytes.NewBuffer([]byte{})
	if _, err := io.CopyN(buf, r, defaultHeaderSize); err != nil {
		return nil, err
	}
	ai.HeaderSize = defaultHeaderSize

	// pick size and type
	data := buf.Next(defaultHeaderSize)
	ai.Size = uint64(binary.BigEndian.Uint32(data))
	ai.Type = media.BoxType{data[4], data[5], data[6], data[7]}

	if ai.Size == 1 {
		// read more 8 bytes
		if _, err := io.CopyN(buf, r, 8); err != nil {
			return nil, err
		}
		ai.HeaderSize += 8
		ai.Size = binary.BigEndian.Uint64(buf.Next(8))
	} else if ai.Size == 0 {
		// box extends to end of file
		offsetEOF, err := r.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		ai.Size = uint64(offsetEOF) - ai.Offset
		if _, err := ai.SeekToPayload(r); err != nil {
			return nil, err
		}
	}

	if ai.Type == media.StrToType("uuid") {
		if _, err := io.CopyN(buf, r, 16); err != nil {
			return nil, err
		}
		copy(ai.ExtendedType[:], buf.Next(16))
	}
	return ai, nil
}

func IsFullAtom(bi *Info) bool {
	if _, ok := Map[bi.Type]; ok {
		t := reflect.TypeOf(Map[bi.Type])
		for i := 0; i < t.NumField(); i++ {
			if reflect.TypeOf(FullAtom{}) == t.Field(i).Type {
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
