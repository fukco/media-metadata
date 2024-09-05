package meta

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"github.com/fukco/media-metadata/internal/box"
	"github.com/fukco/media-metadata/internal/common"
	"github.com/fukco/media-metadata/internal/exif"
	"github.com/fukco/media-metadata/internal/manufacturer"
	"github.com/fukco/media-metadata/internal/manufacturer/nikon"
	"github.com/fukco/media-metadata/internal/manufacturer/panasonic"
	"github.com/fukco/media-metadata/internal/manufacturer/sony/nrtmd"
	"github.com/fukco/media-metadata/internal/manufacturer/sony/rtmd"
	"io"
)

type keyItemPair struct {
	keys *box.Keys
	ilst *box.Ilst
}

func Read(r io.ReadSeeker) (*Metadata, error) {
	fileStructure, err := box.ReadFileStructure(r)
	if err != nil {
		return nil, err
	}
	metadata := &Metadata{
		Manufacturer: fileStructure.Mfr,
		Mp4Meta:      &Mp4Meta{},
		MakerMeta:    &MakerMeta{},
	}
	err = iterator(r, metadata, fileStructure, fileStructure.BoxDetails)
	if err != nil {
		return nil, err
	}
	pair := &keyItemPair{}
	searchKeysAndItems(pair, fileStructure.BoxDetails)
	err = handleKeyAndItems(pair, metadata, fileStructure)
	if err != nil {
		return nil, err
	}
	return metadata, nil
}

func searchKeysAndItems(pair *keyItemPair, boxDetails []*box.BoxDetail) {
	for _, detail := range boxDetails {
		if pair.keys != nil && pair.ilst != nil {
			return
		}
		if detail.Type == box.MetadataItemKeysAtom {
			pair.keys = detail.Boxer.(*box.Keys)
		} else if detail.Type == box.MetadataItemListAtom {
			pair.ilst = detail.Boxer.(*box.Ilst)
		}
		if len(detail.Children) > 0 {
			searchKeysAndItems(pair, detail.Children)
		}
	}
}

func iterator(r io.ReadSeeker, metadata *Metadata, fileStructure *box.FileStructure, boxDetails []*box.BoxDetail) error {
	for _, detail := range boxDetails {
		err := handleBoxDetail(r, metadata, detail, fileStructure)
		if err != nil {
			return err
		}
		if len(detail.Children) > 0 {
			err := iterator(r, metadata, fileStructure, detail.Children)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func handleBoxDetail(r io.ReadSeeker, metadata *Metadata, boxDetail *box.BoxDetail, fileStructure *box.FileStructure) error {
	switch boxDetail.Type {
	case box.MediaBox:
		err := handleMediaBox(r, metadata, boxDetail)
		if err != nil {
			return err
		}
	case box.MetaBox:
		err := handleMeta(metadata, boxDetail, fileStructure)
		if err != nil {
			return err
		}
	case box.UUIDExtensionBox:
		err := handleUUIDExtensionBox(metadata, boxDetail)
		if err != nil {
			return err
		}
	case box.CanonCNDA:
		err := handleCanonCNDA(metadata, boxDetail, fileStructure)
		if err != nil {
			return err
		}
	case box.PanasonicPANABox:
		err := handlePanasonicPANABox(r, metadata, boxDetail, fileStructure)
		if err != nil {
			return err
		}
	case box.FujiMVTGBox:
		err := handleFujiMVTGBox(r, metadata, boxDetail)
		if err != nil {
			return err
		}
	case box.NikonNCTGBox:
		err := handleNikonNCTGBox(metadata, boxDetail)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleKeyAndItems(pair *keyItemPair, metadata *Metadata, fileStructure *box.FileStructure) error {
	if pair.keys == nil || pair.ilst == nil {
		return nil
	}
	keySlice := make([]string, 0, fileStructure.QuickTimeKeysMetaEntryCount)
	for _, entry := range pair.keys.Entries {
		keySlice = append(keySlice, string(entry.Value))
	}
	metaItemKeyValues := make(map[string]any, fileStructure.QuickTimeKeysMetaEntryCount)
	valueMap := make(map[int]any, fileStructure.QuickTimeKeysMetaEntryCount)
	for _, item := range pair.ilst.Items {
		value, _ := item.Value.Value()
		valueMap[int(item.Type)] = value
	}
	for i := 0; i < int(fileStructure.QuickTimeKeysMetaEntryCount); i++ {
		if keySlice[i] == "com.panasonic.Semi-Pro.metadata.xml" {
			handlePanaMetaItemXML(metadata, (valueMap[i+1]).(string))
		} else {
			metaItemKeyValues[keySlice[i]] = valueMap[i+1]
		}
	}
	if len(metaItemKeyValues) > 0 {
		metadata.MetaItemKeyValues = metaItemKeyValues
	}
	return nil
}

func handleMediaBox(r io.ReadSeeker, metadata *Metadata, boxDetail *box.BoxDetail) error {
	handlerType := ""
	for _, child := range boxDetail.Children {
		if child.Type == box.HandlerReferenceBox {
			hdlr := child.Boxer.(*box.Hdlr)
			handlerType = string(hdlr.HandlerType[:])
			break
		}
	}
	if handlerType != "meta" {
		return nil
	}
	sampleSize, offset := uint32(0), uint32(0)
	for _, c1 := range boxDetail.Children {
		if c1.Type == box.MediaInformationBox {
			for _, c2 := range c1.Children {
				if c2.Type == box.SampleTableBox {
					for _, child := range c2.Children {
						stsz, ok := child.Boxer.(*box.Stsz)
						if ok {
							sampleSize = stsz.Size
						}
						stco, ok := child.Boxer.(*box.Stco)
						if ok {
							offset = stco.Offsets[0]
						}
					}
				}
			}
		}
	}
	if sampleSize > 0 && offset > 0 {
		RTMD, err := rtmd.ReadRTMD(r, sampleSize, offset)
		if err != nil {
			return err
		}
		if metadata.MakerMeta.Sony == nil {
			metadata.MakerMeta.Sony = &Sony{}
		}
		metadata.MakerMeta.Sony.RTMD = RTMD
	}
	return nil
}

func handleMeta(metadata *Metadata, boxDetail *box.BoxDetail, fileStructure *box.FileStructure) error {
	if fileStructure.Mfr != manufacturer.SONY {
		return nil
	}
	handlerType := ""
	for _, child := range boxDetail.Children {
		if child.Type == box.HandlerReferenceBox {
			hdlr := child.Boxer.(*box.Hdlr)
			handlerType = string(hdlr.HandlerType[:])
			break
		}
	}
	if handlerType != "nrtm" {
		return nil
	}
	for _, child := range boxDetail.Children {
		if child.Type == box.XMLBox {
			xmlBox := child.Boxer.(*box.XML)
			v := &nrtmd.NonRealTimeMeta{}
			err := xml.Unmarshal([]byte(xmlBox.Xml), v)
			if err != nil {
				return err
			}
			if metadata.MakerMeta.Sony == nil {
				metadata.MakerMeta.Sony = &Sony{}
			}
			metadata.MakerMeta.Sony.NonRealTimeMeta = v
		}
	}
	return nil
}

func handleUUIDExtensionBox(metadata *Metadata, boxDetail *box.BoxDetail) error {
	if boxDetail.Boxer.UserType() == box.TypeUUIDProf() {
		uuidProf := boxDetail.Boxer.(*box.UUIDProf)
		for _, item := range uuidProf.ProfileItems {
			if item.Type == uint32(box.VideoProfile) {
				videoAvgBitrateIndex, pixelAspectRatioIndex := 5, 10
				videoAvgBitrate := binary.BigEndian.Uint32(item.Data[videoAvgBitrateIndex*4 : videoAvgBitrateIndex*4+4])
				metadata.Mp4Meta.VideoProfile = &VideoProfile{
					VideoAvgBitrate: common.ConvertBitrate(videoAvgBitrate),
					PixelAspectRatio: fmt.Sprintf("%d:%d", binary.BigEndian.Uint16(item.Data[pixelAspectRatioIndex*4:pixelAspectRatioIndex*4+2]),
						binary.BigEndian.Uint16(item.Data[pixelAspectRatioIndex*4+2:pixelAspectRatioIndex*4+4])),
				}
				break
			}
		}
	}
	return nil
}

func handleCanonCNDA(metadata *Metadata, boxDetail *box.BoxDetail, fileStructure *box.FileStructure) error {
	cnda := boxDetail.Boxer.(*box.CNDA)
	exifMeta, err := exif.ProcessJPEG(cnda.Data, fileStructure.Mfr)
	if err != nil {
		return err
	}
	metadata.ExifMeta = exifMeta
	return nil
}

func handlePanasonicPANABox(r io.ReadSeeker, metadata *Metadata, boxDetail *box.BoxDetail, fileStructure *box.FileStructure) error {
	bi := boxDetail.BoxInfo
	_, err := r.Seek(int64(bi.Offset+bi.HeaderSize+0x4080), io.SeekStart)
	if err != nil {
		return err
	}
	data := make([]byte, bi.Size-bi.HeaderSize-0x4080)
	if _, err := r.Read(data); err != nil {
		return err
	}
	exifMeta, err := exif.ProcessJPEG(data, fileStructure.Mfr)
	if err != nil {
		return err
	}
	metadata.ExifMeta = exifMeta
	return nil
}

func handleFujiMVTGBox(r io.ReadSeeker, metadata *Metadata, boxDetail *box.BoxDetail) error {
	bi := boxDetail.BoxInfo
	_, err := r.Seek(int64(bi.Offset+bi.HeaderSize+16), io.SeekStart)
	if err != nil {
		return err
	}
	data := make([]byte, bi.Size-bi.HeaderSize-16)
	if _, err := r.Read(data); err != nil {
		return err
	}
	exifMeta, err := exif.Process(data, true, manufacturer.FUJIFILM)
	if err != nil {
		return err
	}
	metadata.ExifMeta = exifMeta
	return nil
}

func handleNikonNCTGBox(metadata *Metadata, boxDetail *box.BoxDetail) error {
	nctgBox := boxDetail.Boxer.(*box.NCTG)
	nctg, err := nikon.ProcessNCTG(nctgBox)
	if err != nil {
		return err
	}
	metadata.MakerMeta.Nikon = &Nikon{
		nctg,
	}
	return nil
}

func handlePanaMetaItemXML(metadata *Metadata, value string) {
	v := &panasonic.ClipMain{}
	if err := xml.Unmarshal([]byte(value), v); err != nil {
		fmt.Println(err)
	}
	metadata.MakerMeta.Panasonic = &Panasonic{v}
}
