package quicktime

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/fukco/media-meta-parser/exif"
	"github.com/fukco/media-meta-parser/manufacturer"
	"github.com/fukco/media-meta-parser/manufacturer/panasonic"
	"github.com/fukco/media-meta-parser/manufacturer/sony"
	"github.com/fukco/media-meta-parser/media"
	"io"
)

/************************** moov **************************/
func AtomTypeMoov() AtomType { return StrToAtomType("moov") }

type Moov struct {
	ContainerAtom
	Atom
}

func init() {
	AppendAtomMap(AtomTypeMoov(), Moov{})
}

/************************** meta **************************/
func AtomTypeMeta() AtomType { return StrToAtomType("meta") }

type Meta struct {
	ContainerAtom
}

func init() {
	AppendAtomMap(AtomTypeMeta(), Meta{})
}

/************************** keys **************************/
func AtomTypeKeys() AtomType { return StrToAtomType("keys") }

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
	AppendAtomMap(AtomTypeKeys(), Keys{})
}

func (a Keys) GetMeta(r io.ReadSeeker, ai *AtomInfo, ctx *media.Context, meta *media.Meta) error {
	if _, err := r.Seek(int64(ai.Offset+ai.HeaderSize+4), io.SeekStart); err != nil {
		return err
	}
	entryCount := bytes.NewBuffer(make([]byte, 0, 4))
	if _, err := io.CopyN(entryCount, r, 4); err != nil {
		return err
	}
	a.EntryCount = binary.BigEndian.Uint32(entryCount.Bytes())
	keyEntries := make([]*KeyEntry, 0, a.EntryCount)
	for i := 0; uint32(i) < a.EntryCount; i++ {
		keySizeBytes := bytes.NewBuffer(make([]byte, 0, 4))
		if _, err := io.CopyN(keySizeBytes, r, 4); err != nil {
			return err
		}
		keySize := binary.BigEndian.Uint32(keySizeBytes.Bytes())
		namespaceBytes := bytes.NewBuffer(make([]byte, 0, 4))
		if _, err := io.CopyN(namespaceBytes, r, 4); err != nil {
			return err
		}
		keyNamespace := namespaceBytes.String()
		keyValueBytes := bytes.NewBuffer(make([]byte, 0, keySize-8))
		if _, err := io.CopyN(keyValueBytes, r, int64(keySize-8)); err != nil {
			return err
		}
		keyValue := keyValueBytes.String()
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
func AtomTypeIlst() AtomType { return StrToAtomType("ilst") }

type Ilst struct {
	FullAtom
	EntryCount uint32
	KeyEntries []KeyEntry
}

func init() {
	AppendAtomMap(AtomTypeIlst(), Ilst{})
}

func (a Ilst) GetMeta(r io.ReadSeeker, ai *AtomInfo, ctx *media.Context, meta *media.Meta) error {
	if _, err := ai.SeekToPayload(r); err != nil {
		return err
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	metadataItems := &MetadataItems{MetadataItems: make(map[string]string, 64)}
	for {
		currentSeek, _ := r.Seek(0, io.SeekCurrent)
		buf.Reset()
		if _, err := io.CopyN(buf, r, 4); err != nil {
			return err
		}
		itemSize := binary.BigEndian.Uint32(buf.Bytes())
		buf.Reset()
		if _, err := io.CopyN(buf, r, 4); err != nil {
			return err
		}
		base1Index := binary.BigEndian.Uint32(buf.Bytes())
		keys := meta.Temp["metaKeys"].([]*KeyEntry)
		buf.Reset()
		if _, err := io.CopyN(buf, r, 4); err != nil {
			return err
		}
		dataSize := binary.BigEndian.Uint32(buf.Bytes())
		buf.Reset()
		if _, err := io.CopyN(buf, r, int64(dataSize-4)); err != nil {
			return err
		}
		if keys[base1Index-1].Value == "com.panasonic.Semi-Pro.metadata.xml" {
			v := &panasonic.ClipMain{}
			if err := xml.Unmarshal(buf.Bytes()[12:], v); err != nil {
				fmt.Println(err)
			}
			meta.Items = append(meta.Items, v)
		} else {
			metadataItems.MetadataItems[keys[base1Index-1].Value] = string(buf.Bytes()[12:])
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
		if rtmd, err := sony.ReadRTMD(r); err != nil {
			return err
		} else {
			meta.Items = append(meta.Items, rtmd)
		}
	}
	return nil
}

/************************** udta **************************/
func AtomTypeUdta() AtomType { return StrToAtomType("udta") }

type Udta struct {
	ContainerAtom
}

func init() {
	AppendAtomMap(AtomTypeUdta(), Udta{})
}

/************************** PANA **************************/
func AtomTypePANA() AtomType { return StrToAtomType("PANA") }

type PANA struct {
}

func init() {
	AppendAtomMap(AtomTypePANA(), PANA{})
}

func (a PANA) GetMeta(r io.ReadSeeker, ai *AtomInfo, ctx *media.Context, meta *media.Meta) error {
	if _, err := r.Seek(int64(ai.Offset+ai.HeaderSize+0x4080), io.SeekStart); err != nil {
		return err
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	if _, err := io.CopyN(buf, r, 12); err != nil {
		return err
	}
	if bytes.Compare(buf.Bytes()[:4], []byte{0xFF, 0xD8, 0xFF, 0xE1}) == 0 &&
		bytes.Compare(buf.Bytes()[6:], []byte{0x45, 0x78, 0x69, 0x66, 0, 0}) == 0 {
		buf.Reset()
		if _, err := io.CopyN(buf, r, int64(ai.Size-ai.HeaderSize-0x4080-12)); err != nil {
			return err
		} else {
			if err := exif.Process(buf.Bytes(), false, meta); err != nil {
				return err
			}
		}
	}
	return nil
}

/************************** MVTG **************************/
func AtomTypeMVTG() AtomType { return StrToAtomType("MVTG") }

type MVTG struct {
}

func init() {
	AppendAtomMap(AtomTypeMVTG(), MVTG{})
}

func (a MVTG) GetMeta(r io.ReadSeeker, ai *AtomInfo, ctx *media.Context, meta *media.Meta) error {
	ctx.Manufacturer = manufacturer.FUJIFILM
	if _, err := r.Seek(int64(ai.Offset+ai.HeaderSize+16), io.SeekStart); err != nil {
		return err
	}
	buf := bytes.NewBuffer(make([]byte, 0, ai.Size-ai.HeaderSize))
	if _, err := io.CopyN(buf, r, int64(ai.Size-ai.HeaderSize)); err != nil {
		return err
	}
	if err := exif.Process(buf.Bytes(), true, meta); err != nil {
		return err
	}
	return nil
}
