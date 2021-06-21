package mp4

import (
	"bytes"
	"encoding/xml"
	"fmt"
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

func (b *Xml) getMeta(r io.ReadSeeker, bi *BoxInfo, ctx *media.Context, meta *media.Meta) error {
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
		//if ctx.MajorBrand == string(media.SONYXAVC) {
		//
		//}
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

func (b *Mdat) getMeta(r io.ReadSeeker, bi *BoxInfo, ctx *media.Context, meta *media.Meta) error {
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
