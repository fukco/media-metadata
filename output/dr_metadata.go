package output

import (
	"fmt"
	"github.com/fukco/media-meta-parser/manufacturer/sony"
	"github.com/fukco/media-meta-parser/media"
	"strconv"
)

type DRMetadata struct {
	FileName           string `csv:"File Name"`
	ClipDirectory      string `csv:"Clip Directory"`
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
	//drMetadata.CodecBitrate = rtmd.CameraUnitMetadata.
	drMetadata.Shutter = rtmd.CameraUnitMetadata.ShutterSpeedTime.String()
	drMetadata.ISO = strconv.Itoa(int(rtmd.CameraUnitMetadata.ISOSensitivity))
	drMetadata.CameraAperture = fmt.Sprintf("%.1f", rtmd.LensUnitMetadata.IrisFNumber)
	drMetadata.FocalPoint = fmt.Sprintf("%.0f mm", rtmd.LensUnitMetadata.LensZoom*1000)
	if rtmd.LensUnitMetadata.FocusRingPosition == 0xffff {
		drMetadata.Distance = "infinity"
	} else {
		drMetadata.Distance = fmt.Sprintf("%.2f m", rtmd.LensUnitMetadata.FocusPositionFromImagePlane)
	}
	drMetadata.CameraNotes = fmt.Sprintf("Exp mode: %s\n", rtmd.CameraUnitMetadata.AutoExposureMode) +
		fmt.Sprintf("Focus Ares: %s\n", rtmd.CameraUnitMetadata.AutoFocusSensingAreaSetting) +
		fmt.Sprintf("White Balance mode: %s", rtmd.CameraUnitMetadata.AutoWhiteBalanceMode)
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
		}
	}
	return nil
}
