package mp4

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/fukco/media-meta-parser/exif"
	"github.com/fukco/media-meta-parser/manufacturer"
	"github.com/fukco/media-meta-parser/manufacturer/sony"
	"github.com/fukco/media-meta-parser/media"
	"io"
)

/************************** moov **************************/
func BoxTypeMoov() BoxType { return StrToBoxType("moov") }

type Moov struct {
	ContainerBox
	Box
}

func init() {
	AppendBoxMap(BoxTypeMoov(), Moov{})
}

/************************** meta **************************/
func BoxTypeMeta() BoxType { return StrToBoxType("meta") }

type Meta struct {
	ContainerBox
	FullBox
}

func init() {
	AppendBoxMap(BoxTypeMeta(), Meta{})
}

/************************** xml  **************************/
func BoxTypeXml() BoxType { return StrToBoxType("xml ") }

type Xml struct {
	FullBox
}

func init() {
	AppendBoxMap(BoxTypeXml(), Xml{})
}

func (b Xml) GetMeta(r io.ReadSeeker, bi *BoxInfo, ctx *media.Context, meta *media.Meta) error {
	if ctx.Manufacturer == manufacturer.SONY {
		if _, err := r.Seek(int64(bi.Offset+bi.HeaderSize+4), io.SeekStart); err != nil {
			return err
		}
		xmlSize := bi.Size - bi.HeaderSize - 4
		buf := bytes.NewBuffer(make([]byte, 0, xmlSize))
		if _, err := io.CopyN(buf, r, int64(xmlSize)); err != nil {
			return err
		} else {
			v := &sony.NonRealTimeMeta{}
			err := xml.Unmarshal(buf.Bytes(), v)
			if err != nil {
				fmt.Println(err)
			}
			meta.Items = append(meta.Items, v)
		}
	}
	return nil
}

/************************** mdat **************************/
func BoxTypeMdat() BoxType { return StrToBoxType("mdat") }

type Mdat struct {
	Box
}

func init() {
	AppendBoxMap(BoxTypeMdat(), Mdat{})
}

func (b Mdat) GetMeta(r io.ReadSeeker, bi *BoxInfo, ctx *media.Context, meta *media.Meta) error {
	if _, err := r.Seek(int64(bi.Offset+bi.HeaderSize), io.SeekStart); err != nil {
		return err
	}
	if ctx.MajorBrand == string(media.SONYXAVC) {
		if rtmd, err := sony.ReadRTMD(r); err != nil {
			return err
		} else {
			meta.Items = append(meta.Items, rtmd)
			//return sony.RTMDMetadata(r, metadata)
		}
	}
	return nil
}

/************************** uuid **************************/
func BoxTypeUuid() BoxType { return StrToBoxType("uuid") }

type Uuid struct {
	Box
}

func init() {
	AppendBoxMap(BoxTypeUuid(), Uuid{})
}

func (b Uuid) GetMeta(r io.ReadSeeker, bi *BoxInfo, ctx *media.Context, meta *media.Meta) error {
	if bi.ExtendedType == [16]byte{0x85, 0xC0, 0xB6, 0x87, 0x82, 0x0F, 0x11, 0xE0, 0x81, 0x11, 0xF4, 0xCE, 0x46, 0x2B, 0x6A, 0x48} {
		ctx.Manufacturer = manufacturer.CANON
		buf := bytes.NewBuffer(make([]byte, 0, 4))
		if _, err := r.Seek(int64(bi.Offset+bi.HeaderSize+16), io.SeekStart); err != nil {
			return err
		}
		for {
			current, err := r.Seek(0, io.SeekCurrent)
			if err != nil {
				return err
			} else if current >= int64(bi.Offset+bi.Size) {
				break
			}
			buf.Reset()
			if _, err := io.CopyN(buf, r, 4); err != nil {
				return err
			}
			size := binary.BigEndian.Uint32(buf.Bytes())
			buf.Reset()
			if _, err := io.CopyN(buf, r, 4); err != nil {
				return err
			}
			boxType := buf.Bytes()
			if string(boxType) == "CNTH" {
				buf.Reset()
				if _, err := io.CopyN(buf, r, 4); err != nil {
					return err
				}
				cndaSize := binary.BigEndian.Uint32(buf.Bytes())
				if _, err := r.Seek(4, io.SeekCurrent); err != nil {
					return err
				}
				jpeg := bytes.NewBuffer(make([]byte, 0, cndaSize-8))
				if _, err := io.CopyN(jpeg, r, int64(cndaSize-8)); err != nil {
					return err
				}
				return exif.ProcessJPEG(jpeg.Bytes(), meta)
			}
			if _, err = r.Seek(current+int64(size), io.SeekStart); err != nil {
				return err
			}
		}
	}
	return nil
}
