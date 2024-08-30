package resolve

import (
	"fmt"
	"github.com/fukco/media-metadata/internal/exif"
	"github.com/fukco/media-metadata/internal/manufacturer/nikon"
	"github.com/fukco/media-metadata/internal/manufacturer/panasonic"
	"github.com/fukco/media-metadata/internal/manufacturer/sony/nrtmd"
	"github.com/fukco/media-metadata/internal/manufacturer/sony/rtmd"
	"github.com/fukco/media-metadata/internal/meta"
	"strconv"
	"strings"
	"time"
)

type DRMetadata struct {
	DateRecorded       string
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
	ShutterAngle       string
	ISO                string
	WhitePoint         string
	WhiteBalanceTint   string
	CameraFirmware     string
	LUTUsed            string
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

func (drMetadata *DRMetadata) parseFromSonyXML(xml *nrtmd.NonRealTimeMeta) {
	drMetadata.DateRecorded = xml.CreationDate.Value
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
			break
		}
	}
	for _, relatedTo := range xml.RelevantFiles.RelatedTo {
		if relatedTo.Rel == "LUT" {
			drMetadata.LUTUsed = relatedTo.File
			break
		}
	}
}

func (drMetadata *DRMetadata) parseFromSonyRTMD(rtmd *rtmd.RTMD) {
	if rtmd.CameraUnitMetadata.WhiteBalance > 0 {
		drMetadata.WhitePoint = strconv.Itoa(int(rtmd.CameraUnitMetadata.WhiteBalance))
	}
	if rtmd.CameraUnitMetadata.ShutterSpeedTime != nil {
		drMetadata.Shutter = rtmd.CameraUnitMetadata.ShutterSpeedTime.String()
	}
	if rtmd.CameraUnitMetadata.ShutterSpeedAngle > 0 {
		drMetadata.ShutterAngle = fmt.Sprintf("%.1f°", rtmd.CameraUnitMetadata.ShutterSpeedAngle)
	}
	if rtmd.CameraUnitMetadata.ISOSensitivity > 0 {
		if rtmd.CameraUnitMetadata.ISOSensitivity == rtmd.CameraUnitMetadata.ExposureIndexOfPhotoMeter {
			drMetadata.ISO = strconv.Itoa(int(rtmd.CameraUnitMetadata.ISOSensitivity))
		} else {
			drMetadata.ISO = fmt.Sprintf("%d EI:%d", rtmd.CameraUnitMetadata.ISOSensitivity, rtmd.CameraUnitMetadata.ExposureIndexOfPhotoMeter)
		}
	}
	if rtmd.CameraUnitMetadata.ImagerDimensionWidth != 0 && rtmd.CameraUnitMetadata.ImagerDimensionHeight != 0 {
		drMetadata.SensorAreaCaptured = fmt.Sprintf("%dμm * %dμm", rtmd.CameraUnitMetadata.ImagerDimensionWidth, rtmd.CameraUnitMetadata.ImagerDimensionHeight)
	}
	if rtmd.CameraUnitMetadata.ColorPrimaries != "" && drMetadata.ColorSpaceNotes == "" {
		drMetadata.ColorSpaceNotes = rtmd.CameraUnitMetadata.ColorPrimaries
	}
	if rtmd.CameraUnitMetadata.CaptureGammaEquation != "" && drMetadata.GammaNotes == "" {
		drMetadata.GammaNotes = rtmd.CameraUnitMetadata.CaptureGammaEquation
	}
	if rtmd.LensUnitMetadata.IrisFNumber > 0 {
		drMetadata.CameraAperture = fmt.Sprintf("%.1f", rtmd.LensUnitMetadata.IrisFNumber)
	}
	if rtmd.LensUnitMetadata.LensZoomPtr != 0 {
		drMetadata.FocalPoint = fmt.Sprintf("%.0f", rtmd.LensUnitMetadata.LensZoomPtr*1000)
	}
	if rtmd.LensUnitMetadata.FocusPositionFromImagePlane > 0 {
		drMetadata.Distance = fmt.Sprintf("%.2f m", rtmd.LensUnitMetadata.FocusPositionFromImagePlane)
	}
}

func (drMetadata *DRMetadata) parseFromPanasonicXML(xml *panasonic.ClipMain) {
	drMetadata.DateRecorded = xml.ClipContent.ClipMetadata.Shoot.StartDate
	drMetadata.ISO = xml.CameraUnitMetadata.ISOSensitivity
	drMetadata.CameraFps = xml.ClipContent.Video.FrameRate
	drMetadata.GammaNotes = xml.CameraUnitMetadata.Gamma.CaptureGamma
	drMetadata.ColorSpaceNotes = xml.CameraUnitMetadata.Gamut.CaptureGamut
	drMetadata.CodecBitrate = xml.ClipContent.Video.Codec.Codec
}

func (drMetadata *DRMetadata) parseFromExif(exifMeta *exif.ExifMeta) {
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

		if values[0x9003] != "" {
			if values[0x9011] != "" {
				parse, err := time.Parse("2006:01:02 15:04:05-07:00", values[0x9003]+values[0x9011])
				if err != nil {
					fmt.Println("not valid time format")
				} else {
					drMetadata.DateRecorded = parse.Format(time.RFC3339)
				}
			} else {
				parse, err := time.Parse("2006:01:02 15:04:05", values[0x9003])
				if err != nil {
					fmt.Println("not valid time format")
				} else {
					drMetadata.DateRecorded = parse.Format(time.DateTime)
				}
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
}

func (drMetadata *DRMetadata) parseFromMetaItems(itemsMap map[string]any) {
	if value, ok := itemsMap["com.atomos.hdr.gamut"]; ok {
		drMetadata.ColorSpaceNotes = value.(string)
	}
	if value, ok := itemsMap["com.atomos.hdr.gamma"]; ok {
		drMetadata.GammaNotes = value.(string)
	}
	if value, ok := itemsMap["com.atomos.hdr.camera"]; ok {
		drMetadata.CameraFormat = value.(string)
	}
	if value, ok := itemsMap["com.atomos.hdr.monitormode"]; ok {
		drMetadata.CameraNotes = fmt.Sprintf("MonitorMode: %s", value)
	}
	if value, ok := itemsMap["com.apple.proapps.image.{TIFF}.Make"]; ok {
		drMetadata.CameraManufacturer = value.(string)
	}
	if value, ok := itemsMap["com.apple.proapps.image.{TIFF}.Model"]; ok {
		drMetadata.CameraType = value.(string)
	}
	if value, ok := itemsMap["com.apple.proapps.image.{TIFF}.Software"]; ok {
		drMetadata.CameraFirmware = value.(string)
	}
	if value, ok := itemsMap["com.atomos.hdr.range"]; ok {
		if drMetadata.CameraNotes != "" {
			drMetadata.CameraNotes = fmt.Sprintf("%s\nRange: %s", drMetadata.CameraNotes, value)
		}
		drMetadata.CameraNotes = fmt.Sprintf("Range: %s", value)
	}
}

func (drMetadata *DRMetadata) parseFromNctg(nctg *nikon.NCTG) {
	parse, err := time.Parse("2006:01:02 15:04:05-07:00", nctg.CreateDate+nctg.TimeZone)
	if err != nil {
		return
	}
	drMetadata.DateRecorded = parse.Format(time.RFC3339)
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
}

func GetDRMetadataFromMeta(m *meta.Metadata) *DRMetadata {
	drMetadata := &DRMetadata{}
	if m.Mp4Meta != nil {
		if m.Mp4Meta.CreationTime != nil {
			drMetadata.DateRecorded = m.Mp4Meta.CreationTime.Format(time.RFC3339)
		}
		if m.Mp4Meta.VideoProfile != nil {
			drMetadata.CodecBitrate = m.Mp4Meta.VideoProfile.VideoAvgBitrate
			drMetadata.PARNotes = m.Mp4Meta.VideoProfile.PixelAspectRatio
		}
	}
	if m.ExifMeta != nil {
		drMetadata.parseFromExif(m.ExifMeta)
	}
	if m.MakerMeta.Nikon != nil && m.MakerMeta.Nikon.NCTG != nil {
		drMetadata.parseFromNctg(m.MakerMeta.Nikon.NCTG)
	}
	if m.MakerMeta.Panasonic != nil && m.MakerMeta.Panasonic.ClipMain != nil {
		drMetadata.parseFromPanasonicXML(m.MakerMeta.Panasonic.ClipMain)
	}
	if m.MakerMeta.Sony != nil {
		if m.MakerMeta.Sony.NonRealTimeMeta != nil {
			drMetadata.parseFromSonyXML(m.MakerMeta.Sony.NonRealTimeMeta)
		}
		if m.MakerMeta.Sony.RTMD != nil {
			drMetadata.parseFromSonyRTMD(m.MakerMeta.Sony.RTMD)
		}
	}
	if len(m.MetaItemKeyValues) > 0 {
		drMetadata.parseFromMetaItems(m.MetaItemKeyValues)
	}
	return drMetadata
}
