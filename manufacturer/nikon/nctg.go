package nikon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/fukco/media-meta-parser/common"
	"github.com/fukco/media-meta-parser/exif"
	"github.com/fukco/media-meta-parser/media"
	"math"
	"reflect"
	"strings"
)

type NCTG struct {
	Make                 string
	Model                string
	Software             string
	CreateDate           string
	DateTimeOriginal     string
	FrameCount           int
	FrameRate            *common.UFraction
	TimeZone             string
	FrameWidth           int
	FrameHeight          int
	AudioChannels        int
	AudioBitsPerSample   int
	AudioSampleRate      int
	NikonDateTime        string
	ElectronicVR         string
	ExposureTime         *common.UFraction
	FNumber              *common.UFraction
	ExposureProgram      string
	ExposureCompensation *common.Fraction
	MeteringMode         string
	FocalLength          *common.UFraction
	SerialNumber         string
	LensInfo             string
	LensMake             string
	LensModel            string
	LensSerialNumber     string
	//MakerNoteVersion      string
	WhiteBalance string
	FocusMode    string
	CropHiSpeed  string
	//Vibration Reduction
	VRInfoVersion      string
	VibrationReduction string
	VRMode             string
	ActiveDLighting    string
	PictureControlData *PictureControlData
	ISOInfo            *ISOInfo
	//LensType              string
	Lens                  string
	FlashMode             string
	LensData              *LensData
	ShutterCount          int
	HighISONoiseReduction string
	AFInfo2               *AFInfo2
}

type PictureControlData struct {
	PictureControlVersion     string
	PictureControlName        string
	PictureControlBase        string
	PictureControlAdjust      string
	PictureControlQuickAdjust string
	Sharpness                 string
	MidRangeSharpness         string
	Clarity                   string
	Contrast                  string
	Brightness                string
	Saturation                string
	Hue                       string
	FilterEffect              string
	ToningEffect              string
	ToningSaturation          string
}

type ISOInfo struct {
	ISO           string
	ISOExpansion  string
	ISO2          string
	ISOExpansion2 string
}

type AFInfo2 struct {
	AFInfo2Version   string
	ContrastDetectAF string
	AFAreaMode       string
	PhaseDetectAF    string
	PrimaryAFPoint   string
	AFPointsUsed     string
}

type LensData struct {
	LensDataVersion string
	FocusDistance   string
}

var nctgNameMap = map[uint32]string{
	0x01:      "Make",
	0x02:      "Model",
	0x03:      "Software",
	0x11:      "CreateDate",
	0x12:      "DateTimeOriginal",
	0x13:      "FrameCount",
	0x16:      "FrameRate",
	0x19:      "TimeZone",
	0x22:      "FrameWidth",
	0x23:      "FrameHeight",
	0x32:      "AudioChannels",
	0x33:      "AudioBitsPerSample",
	0x34:      "AudioSampleRate",
	0x1002:    "NikonDateTime",
	0x1013:    "ElectronicVR",
	0x110829a: "ExposureTime",
	0x110829d: "FNumber",
	0x1108822: "ExposureProgram",
	0x1109204: "ExposureCompensation",
	0x1109207: "MeteringMode",
	0x110920a: "FocalLength",
	0x110a431: "SerialNumber",
	0x110a432: "LensInfo",
	0x110a433: "LensMake",
	0x110a434: "LensModel",
	0x110a435: "LensSerialNumber",
	//0x2000001: "MakerNoteVersion",
	0x2000005: "WhiteBalance",
	0x2000007: "FocusMode",
	0x200001b: "CropHiSpeed",
	0x200001f: "VRInfo",
	0x2000022: "ActiveD-Lighting",
	0x2000023: "PictureControlData",
	0x2000025: "ISOInfo",
	//0x2000083: "LensType",
	0x2000084: "Lens",
	0x2000087: "FlashMode",
	0x2000098: "LensData",
	0x20000a7: "ShutterCount",
	0x20000b1: "HighISONoiseReduction",
	0x20000b7: "AFInfo2",
}

var exposureProgramMap = map[uint16]string{
	0: "Not Defined",
	1: "Manual",
	2: "Program AE",
	3: "Aperture-priority AE",
	4: "Shutter speed priority AE",
	5: "Creative (Slow speed)",
	6: "Action (High speed)",
	7: "Portrait",
	8: "Landscape",
}

var meteringModeMap = map[uint16]string{
	0:   "Unknown",
	1:   "Average",
	2:   "Center-weighted average",
	3:   "Spot",
	4:   "Multi-spot",
	5:   "Multi-segment",
	6:   "Partial",
	255: "Other",
}

var cropHiSpeedMap = map[uint16]string{
	0:  "Off",
	1:  "1.3x Crop",
	2:  "DX Crop",
	3:  "5:4 Crop",
	4:  "3:2 Crop",
	6:  "16:9 Crop",
	8:  "2.7x Crop",
	9:  "DX Movie Crop",
	10: "1.3x Movie Crop",
	11: "FX Uncropped",
	12: "DX Uncropped",
	13: "2.8x Movie Crop",
	14: "1.4x Movie Crop",
	15: "1.5x Movie Crop",
	17: "1:1 Crop",
}

var activeDLightingMap = map[uint16]string{
	0:      "Off",
	1:      "Low",
	3:      "Normal",
	5:      "High",
	7:      "Extra High",
	8:      "Extra High 1",
	9:      "Extra High 2",
	10:     "Extra High 3",
	11:     "Extra High 4",
	0xffff: "Auto",
}

var highISONoiseReductionMap = map[uint16]string{
	0: "Off",
	1: "Minimal",
	2: "Low",
	3: "Medium Low",
	4: "Normal",
	5: "Medium High",
	6: "High",
}

var afAreaModeOnContrastDetectAFOffMap = map[byte]string{
	0:   "Single Area",
	1:   "Dynamic Area",
	2:   "Dynamic Area (closest subject)",
	3:   "Group Dynamic",
	4:   "Dynamic Area (9 points)",
	5:   "Dynamic Area (21 points)",
	6:   "Dynamic Area (51 points)",
	7:   "Dynamic Area (51 points, 3D-tracking)",
	8:   "Auto-area",
	9:   "Dynamic Area (3D-tracking)",
	10:  "Single Area (wide)",
	11:  "Dynamic Area (wide)",
	12:  "Dynamic Area (wide, 3D-tracking)",
	13:  "Group Area",
	14:  "Dynamic Area (25 points)",
	15:  "Dynamic Area (72 points)",
	16:  "Group Area (HL)",
	17:  "Group Area (VL)",
	18:  "Dynamic Area (49 points)",
	128: "Single",
	129: "Auto (41 points)",
	130: "Subject Tracking (41 points)",
	131: "Face Priority (41 points)",
	192: "Pinpoint",
	193: "Single",
	195: "Wide (S)",
	196: "Wide (L)",
	197: "Auto",
}

var afAreaModeOnContrastDetectAFOnMap = map[byte]string{
	0:   "Contrast-detect",
	1:   "Contrast-detect (normal area)",
	2:   "Contrast-detect (wide area)",
	3:   "Contrast-detect (face priority)",
	4:   "Contrast-detect (subject tracking)",
	128: "Single",
	129: "Auto (41 points)",
	130: "Subject Tracking (41 points)",
	131: "Face Priority (41 points)",
	192: "Pinpoint",
	193: "Single",
	194: "Dynamic",
	195: "Wide (S)",
	196: "Wide (L)",
	197: "Auto",
	198: "Auto (People)",
	199: "Auto (Animal)",
	200: "Normal-area AF",
	201: "Wide-area AF",
	202: "Face-priority AF",
	203: "Subject-tracking AF",
	204: "Dynamic Area (S)",
	205: "Dynamic Area (M)",
	206: "Dynamic Area (L)",
	207: "3D-tracking",
}

var isoExpansionMap = map[uint16]string{
	0x000: "Off",
	0x101: "Hi 0.3",
	0x102: "Hi 0.5",
	0x103: "Hi 0.7",
	0x104: "Hi 1.0",
	0x105: "Hi 1.3",
	0x106: "Hi 1.5",
	0x107: "Hi 1.7",
	0x108: "Hi 2.0",
	0x109: "Hi 2.3",
	0x10a: "Hi 2.5",
	0x10b: "Hi 2.7",
	0x10c: "Hi 3.0",
	0x10d: "Hi 3.3",
	0x10e: "Hi 3.5",
	0x10f: "Hi 3.7",
	0x110: "Hi 4.0",
	0x111: "Hi 4.3",
	0x112: "Hi 4.5",
	0x113: "Hi 4.7",
	0x114: "Hi 5.0",
	0x201: "Lo 0.3",
	0x202: "Lo 0.5",
	0x203: "Lo 0.7",
	0x204: "Lo 1.0",
}

var flashModeMap = map[byte]string{
	0:  "Did Not Fire",
	1:  "Fired, Manual",
	3:  "Not Ready",
	7:  "Fired, External",
	8:  "Fired, Commander Mode",
	9:  "Fired, TTL Mode",
	18: "LED Light",
}

func lensInfoFormat(tagData []byte) string {
	result := ""
	if len(tagData) == 32 {
		minFocal := binary.BigEndian.Uint32(tagData[:4]) / binary.BigEndian.Uint32(tagData[4:8])
		maxFocal := binary.BigEndian.Uint32(tagData[8:12]) / binary.BigEndian.Uint32(tagData[12:16])
		minFVal := &common.UFraction{
			Numerator:   binary.BigEndian.Uint32(tagData[16:20]),
			Denominator: binary.BigEndian.Uint32(tagData[20:24]),
		}
		minF := minFVal.FNumberFormat()
		maxFVal := &common.UFraction{
			Numerator:   binary.BigEndian.Uint32(tagData[24:28]),
			Denominator: binary.BigEndian.Uint32(tagData[28:32]),
		}
		maxF := maxFVal.FNumberFormat()
		if minFocal == maxFocal {
			result = fmt.Sprintf("%dmm", minFocal)
		} else {
			result = fmt.Sprintf("%d-%dmm", minFocal, maxFocal)
		}
		if minF == maxF {
			result += fmt.Sprintf(" f/%s", minF)
		} else {
			result += fmt.Sprintf(" f/%s-%s", minF, maxF)
		}
	}
	return result
}

func pictureControlFormat([]byte) string {
	return ""

}

func ProcessNCTG(meta *media.Meta, content []byte, ctx *media.Context) error {
	buf := bytes.NewBuffer(content)
	nctg := &NCTG{}
	for {
		tagId := binary.BigEndian.Uint32(buf.Next(4))
		tagDataFormat := exif.DataType(binary.BigEndian.Uint16(buf.Next(2)))
		tagSize := int(binary.BigEndian.Uint16(buf.Next(2))) * int(exif.TypeSize[tagDataFormat])
		tagData := buf.Next(tagSize)
		if tagName, ok := nctgNameMap[tagId]; ok {
			if tagDataFormat == exif.DTAscii {
				refValue := reflect.ValueOf(nctg).Elem()
				refValue.FieldByName(tagName).SetString(strings.TrimSpace(string(bytes.TrimRight(tagData, "\x00"))))
			} else if tagDataFormat == exif.DTLong {
				switch tagId {
				case 0x13:
					nctg.FrameCount = int(binary.BigEndian.Uint32(tagData[:4]))
				default:
					refValue := reflect.ValueOf(nctg).Elem()
					if len(tagData) == 4 {
						refValue.FieldByName(tagName).SetInt(int64(binary.BigEndian.Uint32(tagData)))
					}
				}
			} else if tagDataFormat == exif.DTByte {
				switch tagId {
				case 0x2000087:
					nctg.FlashMode = flashModeMap[tagData[0]]
				}
			} else if tagDataFormat == exif.DTShort {
				switch tagId {
				case 0x1013:
					if binary.BigEndian.Uint16(tagData[:2]) == 0 {
						nctg.ElectronicVR = "off"
					} else if binary.BigEndian.Uint16(tagData[:2]) == 1 {
						nctg.ElectronicVR = "on"
					}
				case 0x1108822:
					nctg.ExposureProgram = exposureProgramMap[binary.BigEndian.Uint16(tagData)]
				case 0x1109207:
					nctg.MeteringMode = meteringModeMap[binary.BigEndian.Uint16(tagData)]
				case 0x2000022:
					nctg.ActiveDLighting = activeDLightingMap[binary.BigEndian.Uint16(tagData)]
				case 0x200001b:
					nctg.CropHiSpeed = cropHiSpeedMap[binary.BigEndian.Uint16(tagData[:2])]
					if len(tagData) == 14 {
						nctg.CropHiSpeed = fmt.Sprintf("%s (%dx%d cropped to %dx%d at pixel %d,%d)",
							nctg.CropHiSpeed, binary.BigEndian.Uint16(tagData[2:4]), binary.BigEndian.Uint16(tagData[4:6]),
							binary.BigEndian.Uint16(tagData[6:8]), binary.BigEndian.Uint16(tagData[8:10]),
							binary.BigEndian.Uint16(tagData[10:12]), binary.BigEndian.Uint16(tagData[12:14]))
					}
				case 0x20000b1:
					nctg.HighISONoiseReduction = highISONoiseReductionMap[binary.BigEndian.Uint16(tagData)]
				default:
					refValue := reflect.ValueOf(nctg).Elem()
					if len(tagData) == 2 {
						refValue.FieldByName(tagName).SetInt(int64(binary.BigEndian.Uint16(tagData)))
					}
				}
			} else if tagDataFormat == exif.DTRational {
				switch tagId {
				case 0x110a432:
					nctg.LensInfo = lensInfoFormat(tagData)
				case 0x2000084:
					nctg.Lens = lensInfoFormat(tagData)
				default:
					refValue := reflect.ValueOf(nctg).Elem()
					if len(tagData) == 8 {
						uFraction := &common.UFraction{
							Numerator:   binary.BigEndian.Uint32(tagData[:4]),
							Denominator: binary.BigEndian.Uint32(tagData[4:8]),
						}
						refValue.FieldByName(tagName).Set(reflect.ValueOf(uFraction))
					}
				}
			} else if tagDataFormat == exif.DTSRational {
				if tagId == 0x1109204 {
					refValue := reflect.ValueOf(nctg).Elem()
					if len(tagData) == 8 {
						fraction := &common.Fraction{
							Numerator:   int32(binary.BigEndian.Uint32(tagData[:4])),
							Denominator: int32(binary.BigEndian.Uint32(tagData[4:8])),
						}
						refValue.FieldByName(tagName).Set(reflect.ValueOf(fraction))
					}
				}
			} else if tagDataFormat == exif.DTUndefined {
				switch tagId {
				case 0x200001f:
					nctg.VRInfoVersion = string(tagData[:4])
					if tagData[4] == 0 {
						nctg.VibrationReduction = "n/a"
					} else if tagData[4] == 1 {
						nctg.VibrationReduction = "On"
					} else if tagData[4] == 2 {
						nctg.VibrationReduction = "Off"
					}
					if strings.HasPrefix(nctg.Model, "NIKON Z") {
						if tagData[6] == 0 {
							nctg.VRMode = "n/a"
						} else if tagData[6] == 1 {
							nctg.VRMode = "Sport"
						} else if tagData[6] == 3 {
							nctg.VRMode = "Normal"
						}
					}
				case 0x2000023:
					if !bytes.HasPrefix(tagData, []byte("03")) {
						continue
					}
					pictureControlData := &PictureControlData{}
					nctg.PictureControlData = pictureControlData
					pictureControlData.PictureControlVersion = string(tagData[:4])
					pictureControlData.PictureControlName = string(bytes.TrimRight(tagData[8:28], "\x00"))
					pictureControlData.PictureControlBase = string(bytes.TrimRight(tagData[28:48], "\x00"))
					if tagData[54] == 0 {
						pictureControlData.PictureControlAdjust = "Default Settings"
					} else if tagData[54] == 1 {
						pictureControlData.PictureControlAdjust = "Quick Adjust"
					} else if tagData[54] == 2 {
						pictureControlData.PictureControlAdjust = "Full Control"
					}
					//TODO
					//if tagData[55] == 0 {
					//	pictureControlData.PictureControlQuickAdjust = "Normal"
					//} else if tagData[55] == 0x7f {
					//	pictureControlData.PictureControlQuickAdjust = "n/a"
					//}
				case 0x2000025:
					isoInfo := &ISOInfo{}
					nctg.ISOInfo = isoInfo
					//TODO how to show ISO
					isoInfo.ISO = fmt.Sprintf("%.0f", 100*math.Pow(2, float64(tagData[0])/12-5))
					isoInfo.ISOExpansion = isoExpansionMap[binary.BigEndian.Uint16(tagData[4:6])]
					isoInfo.ISO2 = fmt.Sprintf("%.0f", 100*math.Pow(2, float64(tagData[0])/12-5))
					isoInfo.ISOExpansion2 = isoExpansionMap[binary.BigEndian.Uint16(tagData[10:12])]
				case 0x20000b7:
					//TODO not finished
					afInfo2 := &AFInfo2{}
					nctg.AFInfo2 = afInfo2
					afInfo2.AFInfo2Version = string(tagData[:4])
					if tagData[4] == 0 {
						afInfo2.ContrastDetectAF = "off"
					} else if tagData[4] == 1 {
						afInfo2.ContrastDetectAF = "om"
					} else if tagData[4] == 2 {
						afInfo2.ContrastDetectAF = "om(2)"
					}
					if tagData[4] == 0 {
						afInfo2.AFAreaMode = afAreaModeOnContrastDetectAFOffMap[tagData[5]]
					} else {
						afInfo2.AFAreaMode = afAreaModeOnContrastDetectAFOnMap[tagData[5]]
					}
				case 0x2000098:
					if strings.HasPrefix(string(tagData[:4]), "080") {
						lensData := &LensData{}
						nctg.LensData = lensData
						lensData.LensDataVersion = string(tagData[:4])
						//TODO
					}
				default:
					println(tagName)
				}
			}
		}
		if buf.Len() <= 0 {
			break
		}
	}
	meta.Items = append(meta.Items, nctg)
	return nil
}
