package rtmd

import (
	"encoding/binary"
	"github.com/fukco/media-meta-parser/common"
	"math"
)

type tagName [2]byte

type hexStr string

// UDAM ID
// ILCE-7SM3: 966908004678031c20510000f0c01181
// F65: 20500000f0c01181966908004678031c
const (
	LensUnitMetadataHex               hexStr = "060e2b34025301010c02010101010000"
	CameraUnitMetadataHex             hexStr = "060e2b34025301010c02010102010000"
	UserDefinedAcquisitionMetadataHex hexStr = "060e2b34025301010c0201017f010000"
)

var rtmdMap map[tagName]func(tag *tag, rtmd *RTMD) error

// LensUnitMetadata
// 8000  "Iris (F)" 16-bit unsigned integer
// 8001  "Focus Position (Image Plane)"
// 8002  "Focus Position (Front Lens Vertex)"
// 8003  "Macro Setting"
// 8004  "LensZoom (35mm Still Camera Equivalent)"
// 8005  "LensZoom (Actual Focal Length)"
// 8006  "Optical Extender Magnification"
// 8007  "Lens Attributes"
// 8008  "Iris (T)"
// 8009  "Iris Ring Position"
// 800A  "Focus Ring Position"
// 800B  "Zoom Ring Position"
//
// CameraUnitMetadata
// 3210,  "Capture Gamma Equation"
// 3219,  "Color Primaries"
// 321A,  "Coding Equations"
// 8100,  "AutoExposure Mode"
// 8101,  "Auto Focus Sensing Area Setting"
// 8102,  "Color Correction Filter Wheel Setting"
// 8103,  "Neutral Density Filter Wheel Setting"
// 8104,  "Imager Dimension (Effective Width)"
// 8105,  "Imager Dimension (Effective Height)"
// 8106,  "Capture Frame Rate"
// 8107,  "Image Sensor Readout Mode"
// 8108,  "Shutter Speed (Angle)"
// 8109,  "Shutter Speed (Time)"
// 810A,  "Camera Master Gain Adjustment"
// 810B,  "ISO Sensitivity"
// 810C,  "Electrical Extender Magnification"
// 810D,  "Auto White Balance Mode"
// 810E,  "White Balance"
// 810F,  "Camera Master BlackLevel"
// 8110,  "Camera Knee Point"
// 8111,  "Camera Knee Slope"
// 8112,  "Camera Luminance Dynamic Range"
// 8113,  "Camera Setting File URI"
// 8114,  "Camera Attributes"
// 8115,  "Exposure Index of Photo Meter"
// 8116,  "Gamma for CDL"
// 8117,  "ASC CDL V1.2"
// 8118,  "ColorMatrix"
// 8119?
// 811E?

// UserDefinedAcquisitionMetadata
// 8007, "Lens Attributes"
// E000, "UDAM Set Identifier" 10 bytes
// E101, "Effective Marker Coverage"
// E102, "Effective Marker Aspect Ratio"
// E103, "Camera Process Discrimination Code"
// E104, "Rotary Shutter Mode"
// E105, "Raw Black Code Value"
// E106, "Raw Gray Code Value"
// E107, "Raw White Code Value"
// E109, "Monitoring Descriptions"
// E10B, "Monitoring Base Curve"
// E201, "Cooke Protocol Binary Metadata"
// E202, "Cooke Protocol User Metadata"
// E203, "Cooke Protocol Calibration Type"
// E300?
// E301?
// E303, "Lighting Preset"
// E304, "Current record date and time"
func init() {
	rtmdMap = map[tagName]func(*tag, *RTMD) error{
		// Lens Unit Metadata
		{0x80, 0x00}: processIrisFNumber,
		{0x80, 0x01}: processFocusPositionFromImagePlane,
		{0x80, 0x04}: processLensZoom35mm,
		{0x80, 0x05}: processLensZoom,
		{0x80, 0x0a}: processFocusRingPosition,
		{0x80, 0x0b}: processZoomRingPosition,
		// Camera Unit Metadata
		{0x81, 0x00}: processAutoExposureMode,
		{0x81, 0x01}: processAutoFocusSensingAreaSetting,
		{0x81, 0x04}: processImagerDimensionWidth,
		{0x81, 0x05}: processImagerDimensionHeight,
		{0x81, 0x06}: processCaptureFrameRate,
		{0x81, 0x08}: processShutterSpeedAngel,
		{0x81, 0x09}: processShutterSpeedTime,
		{0x81, 0x0a}: processCameraMasterGainAdjustment,
		{0x81, 0x0b}: processISOSensitivity,
		{0x81, 0x0c}: processElectricalExtenderMagnification,
		{0x81, 0x0d}: processAutoWhiteBalanceMode,
		{0x81, 0x15}: processExposureIndexOfPhotoMeter,
		{0x32, 0x10}: processCaptureGammaEquation,
		{0x32, 0x19}: processColorPrimaries,
		{0x32, 0x1a}: processCodingEquations,
		{0xe3, 0x03}: processLightingPreset,
		// User Defined Acquisition Metadata
		{0xe0, 0x00}: processUDAMId,
	}
}

func processIrisFNumber(tag *tag, rtmd *RTMD) error {
	data := tag.data.BigEndianUint16()
	rtmd.LensUnitMetadata.IrisFNumber = math.Pow(2, (1-float64(data)/0x10000)*8)
	return nil
}

func processFocusPositionFromImagePlane(tag *tag, rtmd *RTMD) error {
	rtmd.LensUnitMetadata.FocusPositionFromImagePlane = tag.data.CommonDistanceFormat()
	return nil
}

func processLensZoom35mm(tag *tag, rtmd *RTMD) error {
	result := tag.data.CommonDistanceFormat()
	rtmd.LensUnitMetadata.LensZoom35mmPtr = &result
	return nil
}

func processLensZoom(tag *tag, rtmd *RTMD) error {
	result := tag.data.CommonDistanceFormat()
	rtmd.LensUnitMetadata.LensZoomPtr = &result
	return nil
}

func processFocusRingPosition(tag *tag, rtmd *RTMD) error {
	rtmd.LensUnitMetadata.FocusRingPosition = tag.data.BigEndianUint16()
	return nil
}

func processZoomRingPosition(tag *tag, rtmd *RTMD) error {
	rtmd.LensUnitMetadata.ZoomRingPosition = tag.data.BigEndianUint16()
	return nil
}

func processAutoExposureMode(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.AutoExposureMode = AutoExposureMode(tag.data).String()
	return nil
}

func processAutoFocusSensingAreaSetting(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.AutoFocusSensingAreaSetting = AFMode(tag.data[0]).String()
	return nil
}

func processImagerDimensionWidth(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.ImagerDimensionWidth = tag.data.BigEndianUint16()
	return nil
}

func processImagerDimensionHeight(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.ImagerDimensionHeight = tag.data.BigEndianUint16()
	return nil
}

func processCaptureFrameRate(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.CaptureFrameRate = &common.Fraction{
		Numerator:   int32(binary.BigEndian.Uint32(tag.data[:4])),
		Denominator: int32(binary.BigEndian.Uint32(tag.data[4:])),
	}
	return nil
}

func processShutterSpeedAngel(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.ShutterSpeedAngle = float64(tag.data.BigEndianUint32()) / 60
	return nil
}

func processShutterSpeedTime(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.ShutterSpeedTime = &common.Fraction{
		Numerator:   int32(binary.BigEndian.Uint32(tag.data[:4])),
		Denominator: int32(binary.BigEndian.Uint32(tag.data[4:])),
	}
	return nil
}

func processCameraMasterGainAdjustment(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.CameraMasterGainAdjustment = float64(int16(binary.BigEndian.Uint16(tag.data)) / 100)
	return nil
}

func processISOSensitivity(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.ISOSensitivity = tag.data.BigEndianUint16()
	return nil
}

func processElectricalExtenderMagnification(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.ElectricalExtenderMagnification = float64(tag.data.BigEndianUint16()) / 100
	return nil
}

func processAutoWhiteBalanceMode(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.AutoWhiteBalanceMode = AutoWhiteBalanceMode(tag.data[0]).String()
	return nil
}

func processExposureIndexOfPhotoMeter(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.ExposureIndexOfPhotoMeter = tag.data.BigEndianUint16()
	return nil
}

func processCaptureGammaEquation(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.CaptureGammaEquation = GammaEquation(tag.data).String()
	return nil
}

func processColorPrimaries(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.ColorPrimaries = ColorPrimaries(tag.data).String()
	return nil
}

func processCodingEquations(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.CodingEquations = CodingEquations(tag.data).String()
	return nil
}

func processLightingPreset(tag *tag, rtmd *RTMD) error {
	rtmd.CameraUnitMetadata.LightingPreset = LightingPreset(tag.data[0]).String()
	return nil
}

func processUDAMId(tag *tag, rtmd *RTMD) error {
	if len(rtmd.UserDefinedAcquisitionMetadataSlice) > 0 {
		metadata := rtmd.UserDefinedAcquisitionMetadataSlice[len(rtmd.UserDefinedAcquisitionMetadataSlice)-1]
		metadata.ID = tag.data
	}
	return nil
}
