package box

import (
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/fukco/media-meta-parser/common"
	"github.com/fukco/media-meta-parser/exif"
	"github.com/fukco/media-meta-parser/manufacturer"
	"github.com/fukco/media-meta-parser/manufacturer/sony/nrtmd"
	"github.com/fukco/media-meta-parser/manufacturer/sony/rtmd"
	"github.com/fukco/media-meta-parser/media"
	"io"
)

/************************** moov **************************/
func TypeMoov() Type { return StrToType("moov") }

type Moov struct {
	ContainerBox
	Box
}

func init() {
	AppendBoxMap(TypeMoov(), Moov{})
}

/************************** meta **************************/
func TypeMeta() Type { return StrToType("meta") }

type Meta struct {
	ContainerBox
	FullBox
}

func init() {
	AppendBoxMap(TypeMeta(), Meta{})
}

/************************** xml  **************************/
func TypeXml() Type { return StrToType("xml ") }

type Xml struct {
	FullBox
}

func init() {
	AppendBoxMap(TypeXml(), Xml{})
}

func (b Xml) GetMeta(r io.ReadSeeker, bi *Info, ctx *media.Context, meta *media.Meta) error {
	if ctx.Manufacturer == manufacturer.SONY {
		if _, err := r.Seek(int64(bi.Offset+bi.HeaderSize+4), io.SeekStart); err != nil {
			return err
		}
		xmlSize := bi.Size - bi.HeaderSize - 4
		byteSlice := make([]byte, xmlSize)
		if _, err := r.Read(byteSlice); err != nil {
			return err
		} else {
			v := &nrtmd.NonRealTimeMeta{}
			err := xml.Unmarshal(byteSlice, v)
			if err != nil {
				fmt.Println(err)
			}
			meta.Items = append(meta.Items, v)
		}
	}
	return nil
}

/************************** mdat **************************/
func TypeMdat() Type { return StrToType("mdat") }

type Mdat struct {
	Box
}

func init() {
	AppendBoxMap(TypeMdat(), Mdat{})
}

func (b Mdat) GetMeta(r io.ReadSeeker, bi *Info, ctx *media.Context, meta *media.Meta) error {
	if _, err := r.Seek(int64(bi.Offset+bi.HeaderSize), io.SeekStart); err != nil {
		return err
	}
	if ctx.MajorBrand == string(media.SONYXAVC) {
		if data, err := rtmd.ReadRTMD(r); err != nil {
			return err
		} else {
			meta.Items = append(meta.Items, data)
			//return sony.RTMDMetadata(r, metadata)
		}
	}
	return nil
}

/************************** uuid **************************/
func TypeUuid() Type { return StrToType("uuid") }

type Uuid struct {
	Box
}

func init() {
	AppendBoxMap(TypeUuid(), Uuid{})
}

func (b Uuid) GetMeta(r io.ReadSeeker, bi *Info, ctx *media.Context, meta *media.Meta) error {
	if bi.ExtendedType == [16]byte{0x85, 0xC0, 0xB6, 0x87, 0x82, 0x0F, 0x11, 0xE0, 0x81, 0x11, 0xF4, 0xCE, 0x46, 0x2B, 0x6A, 0x48} {
		ctx.Manufacturer = manufacturer.CANON
		buf := bytes.NewBuffer([]byte{})
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
			if _, err := io.CopyN(buf, r, 8); err != nil {
				return err
			}
			size := binary.BigEndian.Uint32(buf.Next(4))
			Type := string(buf.Next(4))
			if Type == "CNTH" {
				if _, err := io.CopyN(buf, r, 4); err != nil {
					return err
				}
				cndaSize := binary.BigEndian.Uint32(buf.Next(4))
				if _, err := r.Seek(4, io.SeekCurrent); err != nil {
					return err
				}
				if _, err := io.CopyN(buf, r, int64(cndaSize-8)); err != nil {
					return err
				}
				data := make([]byte, cndaSize-8)
				copy(data, buf.Next(int(cndaSize-8)))
				return exif.ProcessJPEG(data, meta)
			}
			if _, err = r.Seek(current+int64(size), io.SeekStart); err != nil {
				return err
			}
		}
	} else if bi.ExtendedType == [16]byte{0x50, 0x52, 0x4F, 0x46, 0x21, 0xD2, 0x4F, 0xCE, 0xBB, 0x88, 0x69, 0x5C, 0xFA, 0xC9, 0xC7, 0x40} {
		//UUID-PROF
		profile := &Profile{
			AudioProfile:      AudioProfile{},
			FileGlobalProfile: FileGlobalProfile{},
			VideoProfile:      VideoProfile{},
		}
		buf := bytes.NewBuffer([]byte{})
		if _, err := r.Seek(int64(bi.Offset+bi.HeaderSize+20), io.SeekStart); err != nil {
			return err
		}
		if _, err := io.CopyN(buf, r, 4); err != nil {
			return err
		}
		featureEntries := int(binary.BigEndian.Uint32(buf.Next(4)))
		for i := 0; i < featureEntries; i++ {
			current, err := r.Seek(0, io.SeekCurrent)
			if err != nil {
				return err
			} else if current >= int64(bi.Offset+bi.Size) {
				break
			}
			if _, err := io.CopyN(buf, r, 8); err != nil {
				return err
			}
			size := binary.BigEndian.Uint32(buf.Next(4))
			featureCode := string(buf.Next(4))
			if featureCode == "VPRF" {
				// VideoProf
				for i := 0; i < int(size-8)/4; i++ {
					buf.Reset()
					if _, err := io.CopyN(buf, r, 4); err != nil {
						return err
					}
					if i == 5 {
						videoAvgBitrate := binary.BigEndian.Uint32(buf.Next(4))
						profile.VideoProfile.VideoAvgBitrate = common.ConvertBitrate(videoAvgBitrate)
					} else if i == 10 {
						profile.VideoProfile.PixelAspectRatio = fmt.Sprintf("%d:%d", binary.BigEndian.Uint16(buf.Next(2)), binary.BigEndian.Uint16(buf.Next(2)))
					}
				}
			}
			if _, err = r.Seek(current+int64(size), io.SeekStart); err != nil {
				return err
			}
		}
		meta.Items = append(meta.Items, profile)
	}
	return nil
}
