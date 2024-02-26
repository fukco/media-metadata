package resolve

import (
	"fmt"
	"github.com/fukco/media-meta-parser/exif"
	"github.com/fukco/media-meta-parser/manufacturer/nikon"
	"github.com/fukco/media-meta-parser/manufacturer/panasonic"
	"github.com/fukco/media-meta-parser/manufacturer/sony/nrtmd"
	"github.com/fukco/media-meta-parser/manufacturer/sony/rtmd"
	"github.com/fukco/media-meta-parser/media"
	"github.com/fukco/media-meta-parser/media/mp4"
	"github.com/fukco/media-meta-parser/metadata"
	"strconv"
	"strings"
)

type DRMetadata struct {
	CameraType         string
	CameraManufacturer string
	CameraSerial       string
	CameraId           string
	CameraNotes        string
	CameraFormat       string
	MediaType          string
	TimeLapseInterval  string
	CameraFps          string
	ShutterType        string
	Shutter            string
	ISO                string
	WhitePoint         string
	WhiteBalanceTint   string
	CameraFirmware     string
	LensType           string
	LensNumber         string
	LensNotes          string
	CameraApertureType string
	CameraAperture     string
	FocalPoint         string
	Distance           string
	Filter             string
	NDFilter           string
	CompressionRatio   string
	CodecBitrate       string
	SensorAreaCaptured string
	PARNotes           string
	AspectRatioNotes   string
	GammaNotes         string
	ColorSpaceNotes    string
}

func drMetadataFromSonyXML(xml *nrtmd.NonRealTimeMeta, drMetadata *DRMetadata) error {
	drMetadata.CameraType = xml.Device.ModelName
	drMetadata.CameraFps = xml.VideoFormat.VideoFrame.FormatFps
	drMetadata.CameraManufacturer = xml.Device.Manufacturer
	drMetadata.CameraSerial = xml.Device.SerialNo
	drMetadata.AspectRatioNotes = xml.VideoFormat.VideoLayout.AspectRatio
	for _, group := range xml.AcquisitionRecord.Groups {
		if group.Name == nrtmd.CameraUnitMetadataSet {
			for _, item := range group.Items {
				if item.Name == nrtmd.CaptureGammaEquation {
					drMetadata.GammaNotes = item.Value
				} else if item.Name == nrtmd.CaptureColorPrimaries {
					drMetadata.ColorSpaceNotes = item.Value
				}
			}
		}
	}
	return nil
}

func drMetadataFromSonyRTMD(rtmd *rtmd.RTMD, drMetadata *DRMetadata) error {
	if rtmd.CameraUnitMetadata.ShutterSpeedTime != nil {
		drMetadata.Shutter = rtmd.CameraUnitMetadata.ShutterSpeedTime.String()
	}
	if rtmd.CameraUnitMetadata.ISOSensitivity > 0 {
		if rtmd.CameraUnitMetadata.ISOSensitivity == rtmd.CameraUnitMetadata.ExposureIndexOfPhotoMeter {
			drMetadata.ISO = strconv.Itoa(int(rtmd.CameraUnitMetadata.ISOSensitivity))
		} else {
			drMetadata.ISO = fmt.Sprintf("%d EI:%d", rtmd.CameraUnitMetadata.ISOSensitivity, rtmd.CameraUnitMetadata.ExposureIndexOfPhotoMeter)
		}
	}
	if rtmd.LensUnitMetadata.IrisFNumber > 0 {
		drMetadata.CameraAperture = fmt.Sprintf("%.1f", rtmd.LensUnitMetadata.IrisFNumber)
	}
	if rtmd.LensUnitMetadata.LensZoomPtr != nil {
		drMetadata.FocalPoint = fmt.Sprintf("%.0f", *rtmd.LensUnitMetadata.LensZoomPtr*1000)
	}
	if rtmd.CameraUnitMetadata.ImagerDimensionWidth != 0 && rtmd.CameraUnitMetadata.ImagerDimensionHeight != 0 {
		drMetadata.SensorAreaCaptured = fmt.Sprintf("%dμm * %dμm", rtmd.CameraUnitMetadata.ImagerDimensionWidth, rtmd.CameraUnitMetadata.ImagerDimensionHeight)
	}
	if rtmd.LensUnitMetadata.FocusPositionFromImagePlane > 0 {
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
	if ifd0Tags, ok := exifMeta.Tags[string(exif.GroupIFD0)]; ok {
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
	if exifTags, ok := exifMeta.Tags[string(exif.GroupExif)]; ok {
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
		} else if value, ok := values[0x8827]; ok {
			drMetadata.ISO = value
		} else if value, ok := values[0x8832]; ok {
			drMetadata.ISO = value
		}
		drMetadata.FocalPoint = values[0x920a]
		drMetadata.CameraSerial = values[0xa431]
		drMetadata.LensNotes = values[0xa432]
		drMetadata.LensType = values[0xa434]
		drMetadata.LensNumber = values[0xa435]
	}
	if makerTags, ok := exifMeta.Tags[fmt.Sprintf("%s: %s", exif.MakerIFD, exif.Panasonic)]; ok {
		for i := range makerTags {
			tag := makerTags[i]
			switch tag.ID {
			case 0x02:
				drMetadata.CameraFirmware = tag.Value
			case 0x25:
				drMetadata.CameraSerial = tag.Value
			case 0x44:
				drMetadata.WhitePoint = tag.Value
			case 0x51:
				drMetadata.LensType = tag.Value
			case 0x52:
				drMetadata.LensNumber = tag.Value
			case 0x9d:
				drMetadata.NDFilter = tag.Value
			case 0x9f:
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
	if makerSubTags, ok := exifMeta.Tags[fmt.Sprintf("%s: %s/%s", exif.MakerIFD, exif.Canon, exif.GroupCanonShotInfo)]; ok {
		values := make(map[uint16]string, len(makerSubTags))
		for i := range makerSubTags {
			tag := makerSubTags[i]
			values[tag.ID] = tag.Value
		}
		drMetadata.NDFilter = values[28]
	}
	if makerSubTags, ok := exifMeta.Tags[fmt.Sprintf("%s: %s/%s", exif.MakerIFD, exif.Canon, exif.GroupCanonProcessingInfo)]; ok {
		for i := range makerSubTags {
			tag := makerSubTags[i]
			if tag.ID == 9 {
				drMetadata.WhitePoint = tag.Value
			}
		}
	}
	if makerSubTags, ok := exifMeta.Tags[fmt.Sprintf("%s: %s/%s", exif.MakerIFD, exif.Canon, exif.GroupCanonLogInfo)]; ok {
		for i := range makerSubTags {
			tag := makerSubTags[i]
			if tag.ID == 9 {
				drMetadata.ColorSpaceNotes = tag.Value
			} else if tag.ID == 11 {
				drMetadata.GammaNotes = tag.Value
			}
		}
	}
	return nil
}

func drMetadataFromMetadataItems(metadataItems *metadata.Items, drMetadata *DRMetadata) error {
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
	if value, ok := itemsMap["com.atomos.hdr.range"]; ok {
		if drMetadata.CameraNotes != "" {
			drMetadata.CameraNotes = fmt.Sprintf("%s\nRange: %s", drMetadata.CameraNotes, value)
		}
		drMetadata.CameraNotes = fmt.Sprintf("Range: %s", value)
	}
	if value, ok := itemsMap["com.atomos.hdr.gamut"]; ok {
		drMetadata.ColorSpaceNotes = value
	}
	if value, ok := itemsMap["com.apple.proapps.image.{TIFF}.Model"]; ok {
		drMetadata.CameraType = value
	}
	return nil
}

func drMetadataFromUuidProfile(profile *mp4.Profile, drMetadata *DRMetadata) error {
	drMetadata.CodecBitrate = profile.VideoProfile.VideoAvgBitrate
	drMetadata.PARNotes = profile.VideoProfile.PixelAspectRatio
	return nil
}

func drMetadataFromNctg(nctg *nikon.NCTG, drMetadata *DRMetadata) error {
	drMetadata.CameraType = nctg.Model
	drMetadata.CameraManufacturer = nctg.Make
	drMetadata.CameraSerial = nctg.SerialNumber
	drMetadata.CameraFps = fmt.Sprintf("%.2f", float32(nctg.FrameRate.Numerator)/float32(nctg.FrameRate.Denominator))
	drMetadata.Shutter = nctg.ExposureTime.ShutterFormat()
	if nctg.ISOInfo.ISOExpansion == "Off" {
		drMetadata.ISO = nctg.ISOInfo.ISO
	} else {
		drMetadata.ISO = fmt.Sprintf("%s %s", nctg.ISOInfo.ISOExpansion, nctg.ISOInfo.ISO)
	}
	if _, err := strconv.Atoi(strings.TrimRight(nctg.WhiteBalance, "K")); err == nil {
		drMetadata.WhitePoint = nctg.WhiteBalance
	}
	drMetadata.CameraFirmware = nctg.Software
	drMetadata.LensType = nctg.LensModel
	drMetadata.LensNotes = nctg.LensInfo
	drMetadata.LensNumber = nctg.LensSerialNumber
	drMetadata.CameraAperture = nctg.FNumber.FNumberFormat()
	drMetadata.FocalPoint = nctg.FocalLength.FocalLengthFormat()
	//drMetadata.Distance =
	//drMetadata.AspectRatioNotes = nctg.CropHiSpeed
	if nctg.PictureControlData.PictureControlBase == "N-LOG" {
		drMetadata.GammaNotes = nctg.PictureControlData.PictureControlBase
	}
	return nil
}

func DRMetadataFromMeta(meta *media.Meta, drMetadata *DRMetadata) error {
	for i := range meta.Items {
		item := meta.Items[i]
		switch item.(type) {
		case *nrtmd.NonRealTimeMeta:
			if err := drMetadataFromSonyXML(item.(*nrtmd.NonRealTimeMeta), drMetadata); err != nil {
				return err
			}
		case *rtmd.RTMD:
			if err := drMetadataFromSonyRTMD(item.(*rtmd.RTMD), drMetadata); err != nil {
				return err
			}
		case *panasonic.ClipMain:
			if err := drMetadataFromPanasonicXML(item.(*panasonic.ClipMain), drMetadata); err != nil {
				return err
			}
		case *exif.ExifMeta:
			if err := drMetadataFromExif(item.(*exif.ExifMeta), drMetadata); err != nil {
				return err
			}
		case *metadata.Items:
			if err := drMetadataFromMetadataItems(item.(*metadata.Items), drMetadata); err != nil {
				return err
			}
		case *mp4.Profile:
			if err := drMetadataFromUuidProfile(item.(*mp4.Profile), drMetadata); err != nil {
				return err
			}
		case *nikon.NCTG:
			if err := drMetadataFromNctg(item.(*nikon.NCTG), drMetadata); err != nil {
				return err
			}
		default:
			continue
		}
	}
	return nil
}
