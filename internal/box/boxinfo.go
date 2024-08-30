package box

import (
	"fmt"
	"io"
	"reflect"
)

// BoxInfo combine header and offset information
type BoxInfo struct {
	// Offset specifies an offset of the box in a file.
	Offset uint64

	// Size specifies size(bytes) of box.
	Size uint64

	// HeaderSize specifies size(bytes) of common fields which are defined as "Box" class member at ISO/IEC 14496-12.
	HeaderSize uint64

	// Type specifies box type which is represented by 4 characters.
	Type BoxType

	// UserType specifies box extended type which is represented by 16 characters.
	UserType UserType

	*FullBoxHeader
}

func ReadBoxInfo(r io.ReadSeeker) (*BoxInfo, error) {
	offset, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	bi := &BoxInfo{
		Offset: uint64(offset),
	}

	if header, fullBoxHeader, err := ReadBoxHeader(r); err != nil {
		return nil, err
	} else {
		bi.Size = uint64(header.Size)
		bi.HeaderSize = 8
		bi.Type = header.Type
		if header.UserType != [16]byte{} {
			bi.UserType = header.UserType
			bi.HeaderSize += 16
		}

		if fullBoxHeader != nil {
			bi.FullBoxHeader = fullBoxHeader
			bi.HeaderSize += 4
		}

		if header.Size == 1 {
			bi.Size = header.ExtendedSize
			bi.HeaderSize += 8
		} else if header.Size == 0 {
			offsetEOF, err := r.Seek(0, io.SeekEnd)
			if err != nil {
				return nil, err
			}
			bi.Size = uint64(offsetEOF) - bi.Offset
			if _, err := bi.SeekToPayload(r); err != nil {
				return nil, err
			}
		}
	}
	return bi, nil
}

func (bi *BoxInfo) IsContainerBox() bool {
	def := bi.GetBoxDef()
	if def != nil {
		return def.isContainer
	}
	return false
}

func (bi *BoxInfo) IsSupportedBox() bool {
	def := bi.GetBoxDef()
	if def != nil {
		return true
	}
	return false
}

func (bi *BoxInfo) IsSupportedVersion(ver uint8) bool {
	def := bi.GetBoxDef()
	if def == nil {
		return false
	}
	if len(def.versions) == 0 {
		return true
	}
	for _, sver := range def.versions {
		if ver == sver {
			return true
		}
	}
	return false
}

func (bi *BoxInfo) GetBoxDef() *BoxDef {
	if bi.Type == UUIDExtensionBox {
		return bi.UserType.getBoxDef()
	} else {
		return bi.Type.getBoxDef()
	}
}

func (bi *BoxInfo) New() (Boxer, error) {
	def := bi.GetBoxDef()
	if def == nil {
		return nil, ErrBoxDefinitionNotFound
	}

	box, ok := reflect.New(def.dataType).Interface().(Boxer)
	if !ok {
		return nil, fmt.Errorf("box type not implements IBox interface: %s", bi.Type.String())
	}

	return box, nil
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
