package atom

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/fukco/media-meta-parser/exif"
	"github.com/fukco/media-meta-parser/manufacturer"
	"github.com/fukco/media-meta-parser/manufacturer/panasonic"
	"github.com/fukco/media-meta-parser/manufacturer/sony/rtmd"
	"github.com/fukco/media-meta-parser/media"
	"github.com/fukco/media-meta-parser/metadata"
	"io"
)

/************************** moov **************************/
func TypeMoov() Type { return strToType("moov") }

type Moov struct {
	ContainerAtom
	Atom
}

func init() {
	appendAtomMap(TypeMoov(), Moov{})
}

/************************** meta **************************/
func TypeMeta() Type { return strToType("meta") }

type Meta struct {
	ContainerAtom
}

func init() {
	appendAtomMap(TypeMeta(), Meta{})
}

/************************** keys **************************/
func TypeKeys() Type { return strToType("keys") }

type KeyEntry struct {
	Size      uint32
	Namespace string
	Value     string
}

type Keys struct {
	FullAtom
	EntryCount uint32
	KeyEntries []KeyEntry
}

func init() {
	appendAtomMap(TypeKeys(), Keys{})
}

func (a Keys) GetMeta(r io.ReadSeeker, ai *Info, ctx *media.Context, meta *media.Meta) error {
	if _, err := r.Seek(int64(ai.Offset+ai.HeaderSize+4), io.SeekStart); err != nil {
		return err
	}
	buf := bytes.NewBuffer([]byte{})
	if _, err := io.CopyN(buf, r, 4); err != nil {
		return err
	}
	a.EntryCount = binary.BigEndian.Uint32(buf.Next(4))
	keyEntries := make([]*KeyEntry, 0, a.EntryCount)
	for i := 0; uint32(i) < a.EntryCount; i++ {
		buf.Reset()
		if _, err := io.CopyN(buf, r, 8); err != nil {
			return err
		}
		keySize := binary.BigEndian.Uint32(buf.Next(4))
		keyNamespace := string(buf.Next(4))
		if _, err := io.CopyN(buf, r, int64(keySize-8)); err != nil {
			return err
		}
		keyValue := buf.String()
		keyEntry := &KeyEntry{
			keySize,
			keyNamespace,
			keyValue,
		}
		keyEntries = append(keyEntries, keyEntry)
	}
	meta.Temp["metaKeys"] = keyEntries
	return nil
}

/************************** ilst **************************/
func TypeIlst() Type { return strToType("ilst") }

type Ilst struct {
	FullAtom
	EntryCount uint32
	KeyEntries []KeyEntry
}

func init() {
	appendAtomMap(TypeIlst(), Ilst{})
}

func (a Ilst) GetMeta(r io.ReadSeeker, ai *Info, ctx *media.Context, meta *media.Meta) error {
	if _, err := ai.SeekToPayload(r); err != nil {
		return err
	}
	buf := bytes.NewBuffer([]byte{})
	metadataItems := &metadata.Items{MetadataItems: make(map[string]string, 64)}
	for {
		currentSeek, _ := r.Seek(0, io.SeekCurrent)
		if _, err := io.CopyN(buf, r, 12); err != nil {
			return err
		}
		itemSize := binary.BigEndian.Uint32(buf.Next(4))
		base1Index := binary.BigEndian.Uint32(buf.Next(4))
		dataSize := binary.BigEndian.Uint32(buf.Next(4))
		keys := meta.Temp["metaKeys"].([]*KeyEntry)
		if _, err := io.CopyN(buf, r, int64(dataSize-4)); err != nil {
			return err
		}
		if keys[base1Index-1].Value == "com.panasonic.Semi-Pro.metadata.xml" {
			v := &panasonic.ClipMain{}
			data := make([]byte, dataSize-4)
			copy(data, buf.Next(int(dataSize-4)))
			if err := xml.Unmarshal(data[12:], v); err != nil {
				fmt.Println(err)
			}
			meta.Items = append(meta.Items, v)
		} else {
			metadataItems.MetadataItems[keys[base1Index-1].Value] = string(buf.Next(int(dataSize - 4))[12:])
		}
		if uint64(currentSeek)+uint64(itemSize) >= ai.Offset+ai.Size {
			break
		}
		_, _ = r.Seek(currentSeek+int64(itemSize), io.SeekStart)
	}
	if len(metadataItems.MetadataItems) > 0 {
		meta.Items = append(meta.Items, metadataItems)
	}

	if ctx.MajorBrand == string(media.SONYXAVC) {
		if result, err := rtmd.ReadRTMD(r); err != nil {
			return err
		} else {
			meta.Items = append(meta.Items, result)
		}
	}
	return nil
}

/************************** udta **************************/
func TypeUdta() Type { return strToType("udta") }

type Udta struct {
	ContainerAtom
}

func init() {
	appendAtomMap(TypeUdta(), Udta{})
}

/************************** PANA **************************/
func TypePANA() Type { return strToType("PANA") }

type PANA struct {
}

func init() {
	appendAtomMap(TypePANA(), PANA{})
}

func (a PANA) GetMeta(r io.ReadSeeker, ai *Info, ctx *media.Context, meta *media.Meta) error {
	if _, err := r.Seek(int64(ai.Offset+ai.HeaderSize+0x4080), io.SeekStart); err != nil {
		return err
	}
	byteSlice := make([]byte, 12)
	if _, err := r.Read(byteSlice); err != nil {
		return err
	}
	if bytes.Compare(byteSlice[:4], []byte{0xFF, 0xD8, 0xFF, 0xE1}) == 0 &&
		bytes.Compare(byteSlice[6:], []byte{0x45, 0x78, 0x69, 0x66, 0, 0}) == 0 {
		data := make([]byte, ai.Size-ai.HeaderSize-0x4080-12)
		if _, err := r.Read(data); err != nil {
			return err
		} else {
			if err := exif.Process(data, false, meta); err != nil {
				return err
			}
		}
	}
	return nil
}

/************************** MVTG **************************/
func TypeMVTG() Type { return strToType("MVTG") }

type MVTG struct {
}

func init() {
	appendAtomMap(TypeMVTG(), MVTG{})
}

func (a MVTG) GetMeta(r io.ReadSeeker, ai *Info, ctx *media.Context, meta *media.Meta) error {
	ctx.Manufacturer = manufacturer.FUJIFILM
	if _, err := r.Seek(int64(ai.Offset+ai.HeaderSize+16), io.SeekStart); err != nil {
		return err
	}
	data := make([]byte, ai.Size-ai.HeaderSize)
	if _, err := r.Read(data); err != nil {
		return err
	}
	if err := exif.Process(data, true, meta); err != nil {
		return err
	}
	return nil
}
