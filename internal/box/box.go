package box

import (
	"fmt"
	"github.com/fukco/media-metadata/internal/manufacturer"
	"io"
	"reflect"
)

type Context struct {
	// QuickTimeKeysMetaEntryCount the expected number of items under the ilst box as observed from the keys box
	QuickTimeKeysMetaEntryCount int32

	Mfr manufacturer.Manufacturer
}

var BoxMap = make(map[BoxType]BoxDef, 64)
var UUIDBoxMap = make(map[UserType]BoxDef, 8)

type BoxDef struct {
	dataType     reflect.Type
	isContainer  bool
	boxOrFullBox BoxOrFullBox
	versions     []uint8
	fields       []*field
}

func AddBoxDef(payload Boxer, isContainer bool, boxOrFullBox BoxOrFullBox, versions ...uint8) {
	BoxMap[payload.BoxType()] = BoxDef{
		dataType:     reflect.TypeOf(payload).Elem(),
		isContainer:  isContainer,
		boxOrFullBox: boxOrFullBox,
		versions:     versions,
		fields:       buildFields(payload),
	}
}

func AddUUIDBoxDef(payload Boxer, userType UserType, isContainer bool, boxOrFullBox BoxOrFullBox, versions ...uint8) {
	UUIDBoxMap[userType] = BoxDef{
		dataType:     reflect.TypeOf(payload).Elem(),
		isContainer:  isContainer,
		boxOrFullBox: boxOrFullBox,
		versions:     versions,
		fields:       buildFields(payload),
	}
}

type BoxOrFullBox int8

const (
	IsBox BoxOrFullBox = iota
	IsFullBox
	Unknown
)

type CustomFielder interface {
	// GetFieldLength returns length of dynamic field
	GetFieldLength(name string, ctx *Context) uint
}

type CustomFieldBase struct {
}

func (c CustomFieldBase) GetFieldLength(name string, ctx *Context) uint {
	panic("GetFieldLength not implemented")
}

type Boxer interface {
	CustomFielder
	BoxType() BoxType
	UserType() UserType
}

type BoxBase struct {
	CustomFieldBase
}

func (b BoxBase) UserType() UserType {
	return [16]byte{}
}

func GetBoxVersion(payload Boxer, info BoxInfo) (uint8, error) {
	if payload.BoxType() != info.Type {
		return 0, fmt.Errorf("box type %d and boxinfo type %d not matched", payload.BoxType(), info.Type)
	}
	return info.Version, nil
}

func ReadBoxPayload(r io.ReadSeeker, bi *BoxInfo, ctx *Context) (Boxer, error) {
	sc, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	if uint64(sc) != bi.Offset+bi.HeaderSize {
		_, err = r.Seek(int64(bi.Offset+bi.HeaderSize), io.SeekStart)
		if err != nil {
			return nil, err
		}
	}
	box, err := UnmarshalAny(r, bi, ctx)
	if err != nil {
		return nil, err
	}

	_, err = r.Seek(int64(bi.Offset+bi.Size), io.SeekStart)
	if err != nil {
		return nil, err
	}

	if ctx != nil {
		if bi.Type == FileTypeBox {
			ftyp := box.(*Ftyp)
			ctx.Mfr = ftyp.GetManufacturer()
		} else if bi.Type == MetadataItemKeysAtom {
			keys := box.(*Keys)
			ctx.QuickTimeKeysMetaEntryCount = keys.EntryCount
		}
	}
	return box, nil
}
