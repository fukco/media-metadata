package rtmd

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/fukco/media-metadata/internal/common"
	"math"
	"strings"
)

var ErrNotMatchedTag = errors.New("not matched tag")

const (
	LensUnitMetadataHex               = "060e2b34025301010c02010101010000"
	CameraUnitMetadataHex             = "060e2b34025301010c02010102010000"
	UserDefinedAcquisitionMetadataHex = "060e2b34025301010c0201017f010000"
)

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
func (t code) processLensUnitMetadata(rtmd *RTMD, raw rawData) error {
	switch t {
	case 0x8000:
		data := raw.BigEndianUint16()
		rtmd.LensUnitMetadata.IrisFNumber = math.Pow(2, (1-float64(data)/0x10000)*8)
	case 0x8001:
		rtmd.LensUnitMetadata.FocusPositionFromImagePlane = raw.CommonDistanceFormat()
	case 0x8004:
		rtmd.LensUnitMetadata.LensZoom35mmPtr = raw.CommonDistanceFormat()
	case 0x8005:
		rtmd.LensUnitMetadata.LensZoomPtr = raw.CommonDistanceFormat()
	case 0x800a:
		rtmd.LensUnitMetadata.FocusRingPosition = raw.BigEndianUint16()
	case 0x800b:
		rtmd.LensUnitMetadata.ZoomRingPosition = raw.BigEndianUint16()
	default:
		return ErrNotMatchedTag
	}
	return nil
}

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
// 8119? =ISO
// 811E? =ISO
// 8120?
func (t code) processCameraUnitMetadata(rtmd *RTMD, raw rawData) error {
	switch t {
	case 0x3210:
		rtmd.CameraUnitMetadata.CaptureGammaEquation = GammaEquation(raw).String()
	case 0x3219:
		rtmd.CameraUnitMetadata.ColorPrimaries = ColorPrimaries(raw).String()
	case 0x321a:
		rtmd.CameraUnitMetadata.CodingEquations = CodingEquations(raw).String()
	case 0x8100:
		rtmd.CameraUnitMetadata.AutoExposureMode = AutoExposureMode(raw).String()
	case 0x8101:
		rtmd.CameraUnitMetadata.AutoFocusSensingAreaSetting = AFMode(raw[0]).String()
	case 0x8104:
		rtmd.CameraUnitMetadata.ImagerDimensionWidth = raw.BigEndianUint16()
	case 0x8105:
		rtmd.CameraUnitMetadata.ImagerDimensionHeight = raw.BigEndianUint16()
	case 0x8106:
		rtmd.CameraUnitMetadata.CaptureFrameRate = &common.Rational{
			Numerator:   int32(binary.BigEndian.Uint32(raw[:4])),
			Denominator: int32(binary.BigEndian.Uint32(raw[4:])),
		}
	case 0x8108:
		rtmd.CameraUnitMetadata.ShutterSpeedAngle = float64(raw.BigEndianUint32()) / 60
	case 0x8109:
		rtmd.CameraUnitMetadata.ShutterSpeedTime = &common.Rational{
			Numerator:   int32(binary.BigEndian.Uint32(raw[:4])),
			Denominator: int32(binary.BigEndian.Uint32(raw[4:])),
		}
	case 0x810a:
		rtmd.CameraUnitMetadata.CameraMasterGainAdjustment = float64(int16(binary.BigEndian.Uint16(raw)) / 100)
	case 0x810b:
		rtmd.CameraUnitMetadata.ISOSensitivity = raw.BigEndianUint16()
	case 0x810c:
		rtmd.CameraUnitMetadata.ElectricalExtenderMagnification = float64(raw.BigEndianUint16()) / 100
	case 0x810d:
		rtmd.CameraUnitMetadata.AutoWhiteBalanceMode = AutoWhiteBalanceMode(raw[0]).String()
	case 0x810e:
		rtmd.CameraUnitMetadata.WhiteBalance = raw.BigEndianUint16()
	case 0x8114:
		rtmd.CameraUnitMetadata.CameraAttributes = strings.TrimRight(string(raw), string(byte(0)))
	case 0x8115:
		rtmd.CameraUnitMetadata.ExposureIndexOfPhotoMeter = raw.BigEndianUint16()
	default:
		return ErrNotMatchedTag
	}
	return nil
}

// UserDefinedAcquisitionMetadata
// E000, "UDAM Set Identifier"
// E101, "Effective Marker Coverage"
// E102, "Effective Marker Aspect Ratio"
// E103, "Camera Process Discrimination Code"
// E104, "Rotary Shutter Mode"
// E105, "Raw Black Code Value"
// E106, "Raw Gray Code Value"
// E107, "Raw White Code Value"
// E108?
// E109, "Monitoring Descriptions"
// E10B, "Monitoring Base Curve"
// E111, "Gamma for look?
// E112, "Color for look"?
// E113, "Pre-CDL Transform"?
// E201, "Cooke Protocol Binary Metadata"
// E202, "Cooke Protocol User Metadata"
// E203, "Cooke Protocol Calibration Type"
// E300, "Image Stabilizer" 0-enabled 1-disabled
// E301?
// E303, "Lighting Preset"
// E304, "Current record date and time"
// 8007, "Lens Attributes"
func (t code) processUserDefinedAcquisitionMetadata(rtmd *RTMD, raw rawData) error {
	switch t {
	case 0xe000:
		if len(rtmd.userDefinedAcquisitionMetadataUnKnownSlice) > 0 {
			metadataSet := rtmd.userDefinedAcquisitionMetadataUnKnownSlice[len(rtmd.userDefinedAcquisitionMetadataUnKnownSlice)-1]
			metadataSet.id = raw
		}
	case 0xe103:
		rtmd.UserDefinedAcquisitionMetadata.CameraProcessDiscriminationCode = strings.ToUpper(hex.EncodeToString(raw))
	case 0xe109:
		rtmd.UserDefinedAcquisitionMetadata.MonitoringDescriptions = strings.TrimRight(string(raw), string(byte(0)))
	case 0xe113:
		rtmd.UserDefinedAcquisitionMetadata.PreCDLTransform = strings.TrimRight(string(raw), string(byte(0)))
	case 0xe300:
		rtmd.UserDefinedAcquisitionMetadata.ImageStabilizerEnabled = raw[0]&1 == 0
	case 0xe303:
		rtmd.UserDefinedAcquisitionMetadata.LightingPreset = LightingPreset(raw[0]).String()
	default:
		return ErrNotMatchedTag
	}
	return nil
}
