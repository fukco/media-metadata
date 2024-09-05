package box

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/fukco/media-metadata/internal/common"
	"github.com/fukco/media-metadata/internal/exif"
	"github.com/fukco/media-metadata/internal/manufacturer"
	"time"
)

var ErrBoxDefinitionNotFound = errors.New("box definition not found")

/************************** ftyp **************************/
type Ftyp struct {
	BoxBase
	MajorBrand       [4]byte               `mp4:"size=8"`
	MinorVersion     uint32                `mp4:"size=32"`
	CompatibleBrands []CompatibleBrandElem `mp4:"size=32"` // reach to end of the box
}

type CompatibleBrandElem struct {
	CompatibleBrand [4]byte `mp4:"size=8"`
}

func (f *Ftyp) BoxType() BoxType {
	return FileTypeBox
}

func init() {
	AddBoxDef(&Ftyp{}, false, IsBox)
}

func (f *Ftyp) GetManufacturer() manufacturer.Manufacturer {
	if string(f.MajorBrand[:]) == string(SONYXAVC) {
		return manufacturer.SONY
	} else if string(f.MajorBrand[:]) == string(NIKO) {
		return manufacturer.NIKON
	}
	for _, brand := range f.CompatibleBrands {
		if string(brand.CompatibleBrand[:]) == string(PANABRAND) {
			return manufacturer.PANASONIC
		} else if string(brand.CompatibleBrand[:]) == string(NIKO) {
			return manufacturer.NIKON
		} else if string(brand.CompatibleBrand[:]) == string(CAEP) {
			return manufacturer.CANON
		}
	}
	return manufacturer.Unknown
}

/************************** moov **************************/
type Moov struct {
	BoxBase
}

func (m *Moov) BoxType() BoxType {
	return MovieBox
}

func init() {
	AddBoxDef(&Moov{}, true, IsBox)
}

/************************** mvhd **************************/
type Mvhd struct {
	BoxBase

	CreationTimeV0     uint32    `mp4:"size=32,ver=0"`
	ModificationTimeV0 uint32    `mp4:"size=32,ver=0"`
	CreationTimeV1     uint64    `mp4:"size=64,ver=1"`
	ModificationTimeV1 uint64    `mp4:"size=64,ver=1"`
	Timescale          uint32    `mp4:"size=32"`
	DurationV0         uint32    `mp4:"size=32,ver=0"`
	DurationV1         uint64    `mp4:"size=64,ver=1"`
	Rate               int32     `mp4:"size=32"` // fixed-point 16.16 - template=0x00010000
	Volume             int16     `mp4:"size=16"` // template=0x0100
	Reserved           int16     `mp4:"size=16,const=0"`
	Reserved2          [2]uint32 `mp4:"size=32,const=0"`
	Matrix             [9]int32  `mp4:"size=32,hex"` // template={ 0x00010000,0,0,0,0x00010000,0,0,0,0x40000000 }
	PreDefined         [6]int32  `mp4:"size=32"`
	NextTrackID        uint32    `mp4:"size=32"`
}

func (m *Mvhd) BoxType() BoxType {
	return MovieHeaderBox
}

func init() {
	AddBoxDef(&Mvhd{}, false, IsFullBox, 0, 1)
}

func (m *Mvhd) GetCreationTime(version uint8) *time.Time {
	// UTC timezone don't know timezone
	switch version {
	case 0:
		return common.Mp4Epoch(int64(m.CreationTimeV0))
	case 1:
		return common.Mp4Epoch(int64(m.CreationTimeV1))
	default:
		return nil
	}
}

func (m *Mvhd) GetModificationTime(version uint8) *time.Time {
	// UTC timezone don't know timezone
	switch version {
	case 0:
		return common.Mp4Epoch(int64(m.ModificationTimeV0))
	case 1:
		return common.Mp4Epoch(int64(m.ModificationTimeV1))
	default:
		return nil
	}
}

func (m *Mvhd) GetDuration(version uint8) uint64 {
	switch version {
	case 0:
		return uint64(m.DurationV0)
	case 1:
		return m.DurationV1
	default:
		return 0
	}
}

/************************** meta **************************/
type Meta struct {
	BoxBase
}

func (m *Meta) BoxType() BoxType {
	return MetaBox
}

func init() {
	AddBoxDef(&Meta{}, true, Unknown)
}

/************************** trak **************************/
type Trak struct {
	BoxBase
}

func (m *Trak) BoxType() BoxType {
	return TrackBox
}

func init() {
	AddBoxDef(&Trak{}, true, IsBox)
}

/************************** mdia **************************/
type Mdia struct {
	BoxBase
}

func (m *Mdia) BoxType() BoxType {
	return MediaBox
}

func init() {
	AddBoxDef(&Mdia{}, true, IsBox)
}

/*************************** hdlr ****************************/
type Hdlr struct {
	BoxBase
	PreDefined  uint32    `mp4:"1,size=32"`
	HandlerType [4]byte   `mp4:"2,size=8"`
	Reserved    [3]uint32 `mp4:"3,size=32,const=0"`
	Name        string    `mp4:"4,string=utf8string"`
}

func (h *Hdlr) BoxType() BoxType {
	return HandlerReferenceBox
}

func init() {
	AddBoxDef(&Hdlr{}, false, IsFullBox, 0)
}

/************************** minf **************************/
type Minf struct {
	BoxBase
}

func (m *Minf) BoxType() BoxType {
	return MediaInformationBox
}

func init() {
	AddBoxDef(&Minf{}, true, IsBox)
}

/************************** stbl **************************/
type Stbl struct {
	BoxBase
}

func (m *Stbl) BoxType() BoxType {
	return SampleTableBox
}

func init() {
	AddBoxDef(&Stbl{}, true, IsBox)
}

/************************** stsc **************************/
type Stsc struct {
	BoxBase
	EntryCount uint32      `mp4:"size=32"`
	Entries    []StscEntry `mp4:"len=dynamic,size=96"`
}

type StscEntry struct {
	FirstChunk             uint32 `mp4:"size=32"`
	SamplesPerChunk        uint32 `mp4:"size=32"`
	SampleDescriptionIndex uint32 `mp4:"size=32"`
}

func (*Stsc) BoxType() BoxType {
	return SampleToChunkBox
}

// GetFieldLength returns length of dynamic field
func (stsc *Stsc) GetFieldLength(name string, ctx *Context) uint {
	switch name {
	case "Entries":
		return uint(stsc.EntryCount)
	}
	panic(fmt.Errorf("invalid name of dynamic-length field: boxType=stsc fieldName=%s", name))
}

func init() {
	AddBoxDef(&Stsc{}, false, IsFullBox)
}

/************************** stsz **************************/
type Stsz struct {
	BoxBase
	Size       uint32   `mp4:"size=32"`
	Count      uint32   `mp4:"size=32"`
	EntrySizes []uint32 `mp4:"size=32,len=dynamic"`
}

func (m *Stsz) GetFieldLength(name string, ctx *Context) uint {
	switch name {
	case "EntrySizes":
		if m.Size == 0 {
			return uint(m.Count)
		} else {
			return 0
		}
	default:
		return 0
	}
}

func (m *Stsz) BoxType() BoxType {
	return SampleSizeBox
}

func init() {
	AddBoxDef(&Stsz{}, false, IsFullBox)
}

/************************** stco **************************/
type Stco struct {
	BoxBase
	Count   uint32   `mp4:"size=32"`
	Offsets []uint32 `mp4:"size=32,len=dynamic"`
}

func (c *Stco) GetFieldLength(name string, ctx *Context) uint {
	switch name {
	case "Offsets":
		return uint(c.Count)
	default:
		return 0
	}
}

func (c *Stco) BoxType() BoxType {
	return ChunkOffsetBox
}

func init() {
	AddBoxDef(&Stco{}, false, IsFullBox)
}

/************************** keys **************************/
type Keys struct {
	BoxBase
	EntryCount int32 `mp4:"size=32"`
	Entries    []Key `mp4:"len=dynamic"`
}

type Key struct {
	CustomFieldBase
	Size      int32  `mp4:"size=32"`
	Namespace []byte `mp4:"size=8,len=4"`
	Value     []byte `mp4:"size=8,len=dynamic"`
}

func (k *Key) GetFieldLength(name string, ctx *Context) uint {
	switch name {
	case "Value":
		// sizeOf(KeySize)+sizeOf(KeyNamespace) = 8 bytes
		return uint(k.Size) - 8
	}
	panic(fmt.Errorf("invalid name of dynamic-length field: boxType=key fieldName=%s", name))
}

func (k *Keys) BoxType() BoxType {
	return MetadataItemKeysAtom
}

func (k *Keys) GetFieldLength(name string, ctx *Context) uint {
	switch name {
	case "Entries":
		return uint(k.EntryCount)
	}
	panic(fmt.Errorf("invalid name of dynamic-length field: boxType=keys fieldName=%s", name))
}

func init() {
	AddBoxDef(&Keys{}, false, IsFullBox)
}

/************************** ilst **************************/
type Ilst struct {
	BoxBase
	Items []MetadataItemAtom `mp4:"len=dynamic"`
}

type MetadataItemAtom struct {
	Size  int32            `mp4:"size=32"`
	Type  int32            `mp4:"size=32"` //index base-1
	Value MetadataDataAtom `mp4:""`
}

type MetadataDataAtom struct {
	CustomFieldBase
	Size     int32   `mp4:"size=32"`
	Type     [4]byte `mp4:"size=8,const=data"`
	DataType uint32  `mp4:"size=32"`
	DataLang uint32  `mp4:"size=32"`
	Data     []byte  `mp4:"size=8,len=dynamic"`
}

const (
	DataTypeReversed        = 0
	DataTypeStringUTF8      = 1
	DataTypeStringUTF16     = 2
	DataTypeUint32BigEndian = 77
)

func (m *MetadataDataAtom) GetFieldLength(name string, ctx *Context) uint {
	switch name {
	case "Data":
		return uint(m.Size - 16)
	default:
		panic(fmt.Errorf("invalid name of dynamic-length field: boxType=data fieldName=%s", name))
	}
}

func (m *MetadataDataAtom) Value() (any, error) {
	if m.DataType == DataTypeStringUTF8 {
		return string(m.Data), nil
	} else if m.DataType == DataTypeUint32BigEndian {
		return binary.BigEndian.Uint32(m.Data), nil
	}
	return nil, fmt.Errorf("not supported type: %d", m.DataType)
}

func (i *Ilst) BoxType() BoxType {
	return MetadataItemListAtom
}

func (i *Ilst) GetFieldLength(name string, ctx *Context) uint {
	switch name {
	case "Items":
		return uint(ctx.QuickTimeKeysMetaEntryCount)
	default:
		return 0
	}
}

func init() {
	AddBoxDef(&Ilst{}, false, IsBox)
}

/************************** xml  **************************/
type XML struct {
	BoxBase
	Xml string `mp4:"string=utfstring"`
}

func (x *XML) BoxType() BoxType {
	return XMLBox
}

func init() {
	AddBoxDef(&XML{}, false, IsFullBox)
}

/************************** udta **************************/
type Udta struct {
	BoxBase
}

func (u *Udta) BoxType() BoxType {
	return UserDataBox
}

func init() {
	AddBoxDef(&Udta{}, true, IsBox)
}

/************************** PANA **************************/
type PANA struct {
	BoxBase
}

func (p *PANA) BoxType() BoxType {
	return PanasonicPANABox
}

func init() {
	AddBoxDef(&PANA{}, false, IsBox)
}

/************************** MVTG **************************/
type MVTG struct {
	BoxBase
}

func (m *MVTG) BoxType() BoxType {
	return FujiMVTGBox
}

func init() {
	AddBoxDef(&MVTG{}, false, IsBox)
}

/************************** NCDT **************************/
type NCDT struct {
	BoxBase
}

func (n *NCDT) BoxType() BoxType {
	return NikonNCDTBox
}

func init() {
	AddBoxDef(&NCDT{}, true, IsBox)
}

/************************** NCTG **************************/
type NCTG struct {
	BoxBase
	Tags []NCTGTag `mp4:""`
}

type NCTGTag struct {
	CustomFieldBase
	ID         uint32 `mp4:"size=32"`
	DataFormat uint16 `mp4:"size=16"`
	Count      uint16 `mp4:"size=16"`
	Data       []byte `mp4:"size=8,len=dynamic"`
}

func (n *NCTGTag) GetFieldLength(name string, ctx *Context) uint {
	switch name {
	case "Data":
		return uint(exif.TypeSize[exif.DataType(n.DataFormat)]) * uint(n.Count)
	default:
		return 0
	}
}

func (n *NCTG) BoxType() BoxType {
	return NikonNCTGBox
}

func init() {
	AddBoxDef(&NCTG{}, false, IsBox)
}

/************************** uuid **************************/
func TypeUUIDProf() UserType {
	return [16]byte{0x50, 0x52, 0x4F, 0x46, 0x21, 0xD2, 0x4F, 0xCE, 0xBB, 0x88, 0x69, 0x5C, 0xFA, 0xC9, 0xC7, 0x40}
}
func TypeUUIDCanon() UserType {
	return [16]byte{0x85, 0xC0, 0xB6, 0x87, 0x82, 0x0F, 0x11, 0xE0, 0x81, 0x11, 0xF4, 0xCE, 0x46, 0x2B, 0x6A, 0x48}
}

type UUIDProf struct {
	BoxBase
	Count        int32         `mp4:"size=32"`
	ProfileItems []ProfileItem `mp4:"len=dynamic"`
}

type ProfileItem struct {
	CustomFieldBase
	Size int32  `mp4:"size=32"`
	Type uint32 `mp4:"size=32"`
	Data []byte `mp4:"size=8,len=dynamic"`
}

func (u *UUIDProf) GetFieldLength(name string, ctx *Context) uint {
	switch name {
	case "ProfileItems":
		return uint(u.Count)
	default:
		return 0
	}
}

func (p *ProfileItem) GetFieldLength(name string, ctx *Context) uint {
	switch name {
	case "Data":
		return uint(p.Size - 8)
	default:
		return 0
	}
}

func (u *UUIDProf) BoxType() BoxType {
	return UUIDExtensionBox
}

func (u *UUIDProf) UserType() UserType {
	return TypeUUIDProf()
}

func init() {
	AddUUIDBoxDef(&UUIDProf{}, TypeUUIDProf(), false, IsFullBox)
}

type UUIDCanon struct {
	BoxBase
}

func (v *UUIDCanon) BoxType() BoxType {
	return UUIDExtensionBox
}

func (v *UUIDCanon) UserType() UserType {
	return TypeUUIDCanon()
}

func init() {
	AddUUIDBoxDef(&UUIDCanon{}, TypeUUIDCanon(), true, IsBox)
}

type CNTH struct {
	BoxBase
}

func (c *CNTH) BoxType() BoxType {
	return CanonCNTH
}

type CNDA struct {
	BoxBase
	Data []byte `mp4:"size=8"`
}

func (c *CNDA) BoxType() BoxType {
	return CanonCNDA
}

func init() {
	AddBoxDef(&CNTH{}, true, IsBox)
	AddBoxDef(&CNDA{}, false, IsBox)
}
