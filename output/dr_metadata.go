package output

import (
	"fmt"
	"github.com/fukco/media-meta-parser/exif"
	"github.com/fukco/media-meta-parser/manufacturer/panasonic"
	"github.com/fukco/media-meta-parser/manufacturer/sony"
	"github.com/fukco/media-meta-parser/media"
	"github.com/fukco/media-meta-parser/quicktime"
	"strconv"
)

type DRMetadata struct {
	FileName      string `csv:"File Name"`
	ClipDirectory string `csv:"Clip Directory"`

	CameraType         string `csv:"Camera Type"`
	CameraManufacturer string `csv:"Camera Manufacturer"`
	CameraSerial       string `csv:"Camera Serial #"`
	CameraId           string `csv:"Camera ID"`
	CameraNotes        string `csv:"Camera Notes"`
	CameraFormat       string `csv:"Camera Format"`
	MediaType          string `csv:"Media Type"`
	TimeLapseInterval  string `csv:"Time-lapse Interval"`
	CameraFps          string `csv:"Camera FPS"`
	ShutterType        string `csv:"Shutter Type"`
	Shutter            string `csv:"Shutter"`
	ISO                string `csv:"ISO"`
	WhitePoint         string `csv:"White Point (Kelvin)"`
	WhiteBalanceTint   string `csv:"White Balance Tint"`
	CameraFirmware     string `csv:"Camera Firmware"`
	LensType           string `csv:"Lens Type"`
	LensNumber         string `csv:"Lens Number"`
	LensNotes          string `csv:"Lens Notes"`
	CameraApertureType string `csv:"Camera Aperture Type"`
	CameraAperture     string `csv:"Camera Aperture"`
	FocalPoint         string `csv:"Focal Point (mm)"`
	Distance           string `csv:"Distance"`
	Filter             string `csv:"Filter"`
	NdFilter           string `csv:"ND Filter"`
	CompressionRatio   string `csv:"Compression Ratio"`
	CodecBitrate       string `csv:"Codec Bitrate"`
	AspectRatioNotes   string `csv:"Aspect Ratio Notes"`
	GammaNotes         string `csv:"Gamma Notes"`
	ColorSpaceNotes    string `csv:"Color Space Notes"`
}

func drMetadataFromSonyXML(xml *sony.NonRealTimeMeta, drMetadata *DRMetadata) error {
	drMetadata.CameraType = xml.Device.ModelName
	drMetadata.CameraFps = xml.VideoFormat.VideoFrame.CaptureFps
	drMetadata.CameraManufacturer = xml.Device.Manufacturer
	drMetadata.CameraSerial = xml.Device.SerialNo
	drMetadata.AspectRatioNotes = xml.VideoFormat.VideoLayout.AspectRatio
	for _, group := range xml.AcquisitionRecord.Groups {
		if group.Name == sony.CameraUnitMetadataSet {
			for _, item := range group.Items {
				if item.Name == sony.CaptureGammaEquation {
					drMetadata.GammaNotes = item.Value
				} else if item.Name == sony.CaptureColorPrimaries {
					drMetadata.ColorSpaceNotes = item.Value
				}
			}
		}
	}
	return nil
}

func drMetadataFromSonyRTMD(rtmd *sony.RTMD, drMetadata *DRMetadata) error {
	drMetadata.Shutter = rtmd.CameraUnitMetadata.ShutterSpeedTime.String()
	drMetadata.ISO = strconv.Itoa(int(rtmd.CameraUnitMetadata.ISOSensitivity))
	drMetadata.CameraAperture = fmt.Sprintf("%.1f", rtmd.LensUnitMetadata.IrisFNumber)
	drMetadata.FocalPoint = fmt.Sprintf("%.0f mm", rtmd.LensUnitMetadata.LensZoom*1000)
	if rtmd.LensUnitMetadata.FocusRingPosition == 0xffff {
		drMetadata.Distance = "infinity"
	} else {
		drMetadata.Distance = fmt.Sprintf("%.2f m", rtmd.LensUnitMetadata.FocusPositionFromImagePlane)
	}
	return nil
}

func drMetadataFromPanasonicXML(xml *panasonic.ClipMain, drMetadata *DRMetadata) error {
	drMetadata.ISO = xml.UserArea.AcquisitionMetadata.CameraUnitMetadata.ISOSensitivity
	drMetadata.CameraFps = xml.ClipContent.EssenceList.Video.FrameRate
	drMetadata.GammaNotes = xml.UserArea.AcquisitionMetadata.CameraUnitMetadata.Gamma.CaptureGamma
	drMetadata.ColorSpaceNotes = xml.UserArea.AcquisitionMetadata.CameraUnitMetadata.Gamut.CaptureGamut
	return nil
}

func drMetadataFromExif(exifMeta *exif.ExifMeta, drMetadata *DRMetadata) error {
	if ifd0Tags, ok := exifMeta.Tags[string(exif.Group_IFD0)]; ok {
		for i := range ifd0Tags {
			tag := ifd0Tags[i]
			if !tag.Undefined {
				if tag.ID == 0x010f {
					drMetadata.CameraManufacturer = tag.Value
				} else if tag.ID == 0x0110 {
					drMetadata.CameraType = tag.Value
				}
			}
		}
	}
	if exifTags, ok := exifMeta.Tags[string(exif.Group_Exif)]; ok {
		values := make(map[uint16]string, len(exifTags))
		for i := range exifTags {
			tag := exifTags[i]
			if !tag.Undefined {
				values[tag.ID] = tag.Value
			}
		}
		drMetadata.Shutter = values[0x829a]
		drMetadata.CameraAperture = values[0x829d]
		if value, ok := values[0x8831]; ok {
			drMetadata.ISO = value
		} else if values[0x8830] == "Standard Output Sensitivity" {
			drMetadata.ISO = values[0x8827]
		}
		drMetadata.FocalPoint = values[0x920a]
		drMetadata.CameraSerial = values[0xa431]
		drMetadata.LensNotes = values[0xa432]
		drMetadata.LensNumber = values[0xa435]
	}
	if makerTags, ok := exifMeta.Tags[fmt.Sprintf("%s: %s", exif.MakerIFD, exif.Panasonic)]; ok {
		for i := range makerTags {
			tag := makerTags[i]
			if tag.ID == 0x02 {
				drMetadata.CameraFirmware = tag.Value
			} else if tag.ID == 0x25 {
				drMetadata.CameraSerial = tag.Value
			} else if tag.ID == 0x44 {
				drMetadata.WhitePoint = tag.Value
			} else if tag.ID == 0x51 {
				drMetadata.LensType = tag.Value
			} else if tag.ID == 0x52 {
				drMetadata.LensNumber = tag.Value
			} else if tag.ID == 0x9d {
				drMetadata.NdFilter = tag.Value
			} else if tag.ID == 0x9f {
				drMetadata.ShutterType = tag.Value
			}
		}
	}
	if makerTags, ok := exifMeta.Tags[fmt.Sprintf("%s: %s", exif.MakerIFD, exif.FUJIFILM)]; ok {
		for i := range makerTags {
			tag := makerTags[i]
			if tag.ID == 0x10 {
				drMetadata.CameraSerial = tag.Value
			} else if tag.ID == 0x3803 {
				drMetadata.GammaNotes = tag.Value
			} else if tag.ID == 0x3820 {
				drMetadata.CameraFps = tag.Value
			}
		}
	}
	if makerTags, ok := exifMeta.Tags[fmt.Sprintf("%s: %s", exif.MakerIFD, exif.Canon)]; ok {
		for i := range makerTags {
			tag := makerTags[i]
			if tag.ID == 0x7 {
				drMetadata.CameraFirmware = tag.Value
			} else if tag.ID == 0x95 && drMetadata.LensType == "" {
				drMetadata.LensType = tag.Value
			}
		}
	}
	if makerSubTags, ok := exifMeta.Tags[fmt.Sprintf("%s: %s/%s", exif.MakerIFD, exif.Canon, exif.Group_Canon_Shot_Info)]; ok {
		values := make(map[uint16]string, len(makerSubTags))
		for i := range makerSubTags {
			tag := makerSubTags[i]
			values[tag.ID] = tag.Value
		}
		drMetadata.NdFilter = values[28]
		autoISO, _ := strconv.ParseInt(values[1], 10, 64)
		baseISO, _ := strconv.ParseInt(values[2], 10, 64)
		drMetadata.ISO = fmt.Sprintf("%d", baseISO/autoISO*100)
	}
	if makerSubTags, ok := exifMeta.Tags[fmt.Sprintf("%s: %s/%s", exif.MakerIFD, exif.Canon, exif.Group_Canon_Processing_Info)]; ok {
		for i := range makerSubTags {
			tag := makerSubTags[i]
			if tag.ID == 9 {
				drMetadata.WhitePoint = tag.Value
			}
		}
	}
	return nil
}

func drMetadataFromMetadataItems(metadataItems *quicktime.MetadataItems, drMetadata *DRMetadata) error {
	itemsMap := metadataItems.MetadataItems
	if value, ok := itemsMap["com.apple.proapps.image.{TIFF}.Software"]; ok {
		drMetadata.CameraFirmware = value
	}
	if value, ok := itemsMap["com.atomos.hdr.gamma"]; ok {
		drMetadata.GammaNotes = value
	}
	if value, ok := itemsMap["com.apple.proapps.image.{TIFF}.Make"]; ok {
		drMetadata.CameraManufacturer = value
	}
	if value, ok := itemsMap["com.atomos.hdr.camera"]; ok {
		drMetadata.CameraFormat = value
	}
	if value, ok := itemsMap["com.atomos.hdr.monitormode"]; ok {
		drMetadata.CameraNotes = fmt.Sprintf("MonitorMode: %s", value)
	}
	if value, ok := itemsMap["com.atomos.hdr.gamut"]; ok {
		drMetadata.ColorSpaceNotes = value
	}
	if value, ok := itemsMap["com.apple.proapps.image.{TIFF}.Model"]; ok {
		drMetadata.CameraType = value
	}
	return nil
}

func DRMetadataFromMeta(meta *media.Meta, csv *DRMetadata) error {
	for i := range meta.Items {
		item := meta.Items[i]
		switch item.(type) {
		case *sony.NonRealTimeMeta:
			if err := drMetadataFromSonyXML(item.(*sony.NonRealTimeMeta), csv); err != nil {
				return err
			}
		case *sony.RTMD:
			if err := drMetadataFromSonyRTMD(item.(*sony.RTMD), csv); err != nil {
				return err
			}
		case *panasonic.ClipMain:
			if err := drMetadataFromPanasonicXML(item.(*panasonic.ClipMain), csv); err != nil {
				return err
			}
		case *exif.ExifMeta:
			if err := drMetadataFromExif(item.(*exif.ExifMeta), csv); err != nil {
				return err
			}
		case *quicktime.MetadataItems:
			if err := drMetadataFromMetadataItems(item.(*quicktime.MetadataItems), csv); err != nil {
				return err
			}
		}
	}
	return nil
}
