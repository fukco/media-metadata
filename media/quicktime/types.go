package quicktime

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/fukco/media-meta-parser/exif"
	"github.com/fukco/media-meta-parser/manufacturer"
	"github.com/fukco/media-meta-parser/manufacturer/nikon"
	"github.com/fukco/media-meta-parser/manufacturer/panasonic"
	"github.com/fukco/media-meta-parser/manufacturer/sony/rtmd"
	"github.com/fukco/media-meta-parser/media"
	"github.com/fukco/media-meta-parser/metadata"
	"io"
)

/************************** moov **************************/
func TypeMoov() media.BoxType { return media.StrToType("moov") }

type Moov struct {
	media.ContainerBox
	media.Box
}

func init() {
	media.AppendQuicktimeMap(TypeMoov(), Moov{})
}

/************************** meta **************************/
func TypeMeta() media.BoxType { return media.StrToType("meta") }

type Meta struct {
	media.ContainerBox
}

func init() {
	media.AppendQuicktimeMap(TypeMeta(), Meta{})
}

/************************** keys **************************/
func TypeKeys() media.BoxType { return media.StrToType("keys") }

type KeyEntry struct {
	Size      uint32
	Namespace string
	Value     string
}

type Keys struct {
	media.FullBox
	EntryCount uint32
	KeyEntries []KeyEntry
}

func init() {
	media.AppendQuicktimeMap(TypeKeys(), Keys{})
}

func (a Keys) GetMeta(m *media.Media, bi *media.BoxInfo, meta *media.Meta) error {
	var r io.ReadSeeker = m.File
	if _, err := r.Seek(int64(bi.Offset+bi.HeaderSize+4), io.SeekStart); err != nil {
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
func TypeIlst() media.BoxType { return media.StrToType("ilst") }

type Ilst struct {
	media.FullBox
	EntryCount uint32
	KeyEntries []KeyEntry
}

func init() {
	media.AppendQuicktimeMap(TypeIlst(), Ilst{})
}

func (a Ilst) GetMeta(m *media.Media, bi *media.BoxInfo, meta *media.Meta) error {
	var r io.ReadSeeker = m.File
	if _, err := bi.SeekToPayload(r); err != nil {
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
		if uint64(currentSeek)+uint64(itemSize) >= bi.Offset+bi.Size {
			break
		}
		_, _ = r.Seek(currentSeek+int64(itemSize), io.SeekStart)
	}
	if len(metadataItems.MetadataItems) > 0 {
		meta.Items = append(meta.Items, metadataItems)
	}

	if m.Ftyp.MajorBrand == string(media.SONYXAVC) {
		if result, err := rtmd.ReadRTMD(r); err != nil {
			return err
		} else {
			meta.Items = append(meta.Items, result)
		}
	}
	return nil
}

/************************** udta **************************/
func TypeUdta() media.BoxType { return media.StrToType("udta") }

type Udta struct {
	media.ContainerBox
}

func init() {
	media.AppendQuicktimeMap(TypeUdta(), Udta{})
}

/************************** PANA **************************/
func TypePANA() media.BoxType { return media.StrToType("PANA") }

type PANA struct {
}

func init() {
	media.AppendQuicktimeMap(TypePANA(), PANA{})
}

func (a PANA) GetMeta(m *media.Media, bi *media.BoxInfo, meta *media.Meta) error {
	var r io.ReadSeeker = m.File
	if _, err := r.Seek(int64(bi.Offset+bi.HeaderSize+0x4080), io.SeekStart); err != nil {
		return err
	}
	byteSlice := make([]byte, 12)
	if _, err := r.Read(byteSlice); err != nil {
		return err
	}
	if bytes.Compare(byteSlice[:4], []byte{0xFF, 0xD8, 0xFF, 0xE1}) == 0 &&
		bytes.Compare(byteSlice[6:], []byte{0x45, 0x78, 0x69, 0x66, 0, 0}) == 0 {
		data := make([]byte, bi.Size-bi.HeaderSize-0x4080-12)
		if _, err := r.Read(data); err != nil {
			return err
		} else {
			if err := exif.Process(data, false, meta, manufacturer.PANASONIC); err != nil {
				return err
			}
		}
	}
	return nil
}

/************************** MVTG **************************/
func TypeMVTG() media.BoxType { return media.StrToType("MVTG") }

type MVTG struct {
}

func init() {
	media.AppendQuicktimeMap(TypeMVTG(), MVTG{})
}

func (a MVTG) GetMeta(m *media.Media, bi *media.BoxInfo, meta *media.Meta) error {
	var r io.ReadSeeker = m.File
	m.Manufacturer = manufacturer.FUJIFILM
	if _, err := r.Seek(int64(bi.Offset+bi.HeaderSize+16), io.SeekStart); err != nil {
		return err
	}
	data := make([]byte, bi.Size-bi.HeaderSize)
	if _, err := r.Read(data); err != nil {
		return err
	}
	if err := exif.Process(data, true, meta, manufacturer.FUJIFILM); err != nil {
		return err
	}
	return nil
}

/************************** NCDT **************************/
func TypeNCDT() media.BoxType { return media.StrToType("NCDT") }

type NCDT struct {
}

func init() {
	media.AppendQuicktimeMap(TypeNCDT(), NCDT{})
}

func (a NCDT) GetMeta(m *media.Media, bi *media.BoxInfo, meta *media.Meta) error {
	var r io.ReadSeeker = m.File
	buf := bytes.NewBuffer([]byte{})
	current, err := bi.SeekToPayload(r)
	if err != nil {
		return err
	}
	for {
		if current >= int64(bi.Offset+bi.Size) {
			break
		}
		if _, err := io.CopyN(buf, r, 8); err == nil {
			data := buf.Next(8)
			size := binary.BigEndian.Uint32(data[:4])
			if bytes.Compare(data[4:], []byte("NCTG")) == 0 {
				content := make([]byte, size)
				_, err := r.Read(content)
				if err != nil {
					return err
				}
				return nikon.ProcessNCTG(meta, content)
			} else {
				if _, err := r.Seek(int64(size-8), io.SeekCurrent); err != nil {
					return err
				}
			}
			current += int64(size)
		} else {
			return err
		}
	}
	return nil
}
