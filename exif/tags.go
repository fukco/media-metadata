package exif

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var exifTagDefinitionMap = map[uint16]*TagDefinition{
	// IFD0
	0x010f: {Name: "Make"},
	0x0110: {Name: "Model"},
	// EXIF IFD
	0x829a: {Name: "Exposure time", Fn: func(v interface{}) string {
		data := v.([]int64)
		if data[0]*3 > data[1] {
			return fmt.Sprintf("%.1f″", float64(data[0])/float64(data[1]))
		} else if data[0] == 0 || data[1] == 0 {
			return "undef"
		} else {
			return fmt.Sprintf("1/%d", data[1]/data[0])
		}
	}},
	0x829d: {Name: "F number", Fn: func(v interface{}) string {
		data := v.([]int64)
		if data[0] == 0 || data[1] == 0 {
			return "undef"
		} else if data[0] < data[1] {
			return fmt.Sprintf("F/%.2f", float64(data[0])/float64(data[1]))
		} else {
			return fmt.Sprintf("F/%.1f", float64(data[0])/float64(data[1]))
		}
	}},
	0x8822: {Name: "Exposure program", Fn: func(v interface{}) string {
		data := v.(int64)
		// 0 未定义 1 手动 2 自动（程序曝光） 3 光圈优先 4 快门优先 5 创意（偏向景深） 6 动作（偏向高速快门） 7 肖像模式 8 风景模式
		switch data {
		case 0:
			return "Not defined"
		case 1:
			return "Manual"
		case 2:
			return "Normal program"
		case 3:
			return "Aperture priority"
		case 4:
			return "Shutter priority"
		case 5:
			return "Creative program (based toward depth of field)"
		case 6:
			return "Action program (based toward fast shutter speed)"
		case 7:
			return "Portrait mode (for closeup photos with the background out of focus)"
		case 8:
			return "Landscape mode (for landscape photos with the background in focus)"
		default:
			return "Custom"
		}
	}},
	0x8827: {Name: "Photographic Sensitivity"},
	0x8830: {Name: "Sensitivity Type", Fn: func(v interface{}) string {
		data := v.(int64)
		switch data {
		case 0:
			return "Unknown"
		case 1:
			return "Standard Output Sensitivity"
		case 2:
			return "Recommended Exposure Index"
		case 3:
			return "ISO Speed"
		case 4:
			return "Standard Output Sensitivity and Recommended Exposure Index"
		case 5:
			return "Standard Output Sensitivity and ISO Speed"
		case 6:
			return "Recommended Exposure Index and ISO Speed"
		case 7:
			return "Standard Output Sensitivity, Recommended Exposure Index and ISO Speed"
		default:
			return "Reserved"
		}
	}},
	0x8831: {Name: "Standard Output Sensitivity"},
	0x9204: {Name: "Exposure bias", Fn: func(v interface{}) string {
		data := v.([]int64)
		return fmt.Sprintf("%.1f", float64(data[0])/float64(data[1]))
	}},
	0x9207: {Name: "Metering mode", Fn: func(v interface{}) string {
		data := v.(int64)
		switch data {
		// 0 未知 1 平均测光 2 中央重点测光 3 点测光 4 多点测光 5 矩阵测光 6 部分测光 255 其他
		case 0:
			return "unknown"
		case 1:
			return "Average"
		case 2:
			return "CenterWeightedAverage"
		case 3:
			return "Spot"
		case 4:
			return "MultiSpot"
		case 5:
			return "Pattern"
		case 6:
			return "Partial"
		case 255:
			return "other"
		default:
			return "reserved"
		}
	}},
	0x920a: {Name: "Lens focal length", Fn: func(v interface{}) string {
		data := v.([]int64)
		return fmt.Sprintf("%.1f mm", float64(data[0])/float64(data[1]))
	}},
	0xa431: {Name: "Body Serial Number", Fn: func(v interface{}) string {
		str := v.(string)
		return strings.TrimRight(str, string(byte(0)))
	}},
	0xa432: {Name: "Lens Specification", Fn: func(v interface{}) string {
		data := v.([][]int64)
		minFocalLength := data[0]
		maxFocalLength := data[1]
		minFNumber := data[2]
		maxFNumber := data[3]
		str := fmt.Sprintf("%d", minFocalLength[0]/minFocalLength[1])
		if minFocalLength[0] != maxFocalLength[0] || minFocalLength[1] != maxFocalLength[1] {
			str = str + fmt.Sprintf("-%d", maxFocalLength[0]/maxFocalLength[1])
		}
		if minFNumber[0] != 0 {
			if minFNumber[0] < minFNumber[1] {
				str = str + fmt.Sprintf("mm f/%.2f", float64(minFNumber[0])/float64(minFNumber[1]))
			} else {
				str = str + fmt.Sprintf("mm f/%.1f", float64(minFNumber[0])/float64(minFNumber[1]))
			}
			if minFNumber[0] != maxFNumber[0] || minFNumber[1] != maxFNumber[1] {
				if maxFNumber[0] < maxFNumber[1] {
					str = str + fmt.Sprintf("-%.2f", float64(maxFNumber[0])/float64(maxFNumber[1]))
				} else {
					str = str + fmt.Sprintf("-%.1f", float64(maxFNumber[0])/float64(maxFNumber[1]))
				}
			}
		} else {
			str = str + "mm"
		}
		return str
	}},
	0xa434: {Name: "Lens Model"},
	0xa435: {Name: "Lens Serial Number"},
}

var panasonicTagDefinitionMap = map[uint16]*TagDefinition{
	0x02: {Name: "Firmware Version", Fn: func(v interface{}) string {
		data := v.([]uint8)
		str := make([]string, 0, len(data))
		for _, num := range data {
			str = append(str, strconv.FormatUint(uint64(num), 10))
		}
		res := strings.Join(str, ".")
		return res
	}},
	0x03: {Name: "White Balance", Fn: func(v interface{}) string {
		switch v.(int64) {
		case 1:
			return "Auto"
		case 2:
			return "Daylight"
		case 3:
			return "Cloudy"
		case 4:
			return "Incandescent"
		case 5:
			return "Manual"
		case 8:
			return "Flash"
		case 10:
			return "Black & White"
		case 11:
			return "Manual 2"
		case 12:
			return "Shade"
		case 13:
			return "Kelvin"
		case 14:
			return "'Manual 3"
		case 15:
			return "Manual 4"
		case 19:
			return "Auto (cool)"
		default:
			return ""
		}
	}},
	0x07: {Name: "Focus Mode", Fn: func(v interface{}) string {
		switch v.(int64) {
		case 1:
			return "Auto"
		case 2:
			return "Manual"
		case 4:
			return "Auto, Focus button"
		case 5:
			return "Auto, Continuous"
		case 6:
			return "AF-S"
		case 7:
			return "AF-C"
		case 8:
			return "AF-F"
		default:
			return ""
		}
	}},
	0x0f: {Name: "AF Area Mode", Fn: func(v interface{}) string {
		data := v.([]int64)
		byteSlice := make([]byte, 0, len(data))
		for _, datum := range data {
			byteSlice = append(byteSlice, byte(datum))
		}
		switch {
		case bytes.Compare(byteSlice, []byte{0, 1}) == 0:
			return "9-area"
		case bytes.Compare(byteSlice, []byte{0, 16}) == 0:
			return "3-area (high speed)"
		case bytes.Compare(byteSlice, []byte{0, 23}) == 0:
			return "23-area"
		case bytes.Compare(byteSlice, []byte{0, 49}) == 0:
			return "49-area"
		case bytes.Compare(byteSlice, []byte{0, 225}) == 0:
			return "225-area"
		case bytes.Compare(byteSlice, []byte{1, 0}) == 0:
			return "Spot Focusing"
		case bytes.Compare(byteSlice, []byte{1, 1}) == 0:
			return "5-area"
		case bytes.Compare(byteSlice, []byte{16}) == 0:
			return "Normal?"
		case bytes.Compare(byteSlice, []byte{16, 0}) == 0:
			return "1-area"
		case bytes.Compare(byteSlice, []byte{16, 16}) == 0:
			return "1-area (high speed)"
		case bytes.Compare(byteSlice, []byte{32, 0}) == 0:
			return "Tracking"
		case bytes.Compare(byteSlice, []byte{32, 1}) == 0:
			return "3-area (left)?"
		case bytes.Compare(byteSlice, []byte{32, 2}) == 0:
			return "3-area (center)?"
		case bytes.Compare(byteSlice, []byte{32, 3}) == 0:
			return "3-area (right)?"
		case bytes.Compare(byteSlice, []byte{64, 0}) == 0:
			return "Face Detect"
		case bytes.Compare(byteSlice, []byte{64, 1}) == 0:
			return "Face Detect (animal detect on)"
		case bytes.Compare(byteSlice, []byte{64, 2}) == 0:
			return "Face Detect (animal detect off)"
		case bytes.Compare(byteSlice, []byte{128, 0}) == 0:
			return "Pinpoint focus"
		case bytes.Compare(byteSlice, []byte{240, 0}) == 0:
			return "Tracking"
		default:
			return ""
		}
	}},
	0x25: {Name: "Internal Serial Number", Fn: func(v interface{}) string {
		str := string(v.([]byte))
		return strings.TrimRight(str, string(byte(0)))
	}},
	0x44: {Name: "Color Temp Kelvin"},
	0x51: {Name: "Lens Type"},
	0x52: {Name: "Lens Serial Number", Fn: func(v interface{}) string {
		str := v.(string)
		return strings.TrimRight(str, string(byte(0)))
	}},
	0x60: {Name: "Lens Firmware Version", Fn: func(v interface{}) string {
		data := v.([]uint8)
		str := make([]string, 0, len(data))
		for _, num := range data {
			str = append(str, strconv.FormatUint(uint64(num), 10))
		}
		res := strings.Join(str, ".")
		return res
	}},
	0x9d: {Name: "Internal ND Filter", Fn: func(v interface{}) string {
		data := v.([]int64)
		return fmt.Sprintf("%.1f", float64(data[0])/float64(data[1]))
	}},
	0x9f: {Name: "Shutter Type", Fn: func(v interface{}) string {
		switch v.(int64) {
		case 0:
			return "Mechanical"
		case 1:
			return "Electronic"
		case 2:
			return "Hybrid"
		default:
			return ""
		}
	}},
}

var fujiExifTagDefinitionMap = map[uint16]*TagDefinition{
	0x10: {Name: "Internal Serial Number", Fn: func(v interface{}) string {
		return fmt.Sprintf("%v", v)
	}},
	0x1002: {Name: "White Balance", Fn: func(v interface{}) string {
		data := v.(int64)
		switch data {
		case 0x0:
			return "Auto"
		case 0x1:
			return "Auto (white priority)"
		case 0x2:
			return "Auto (ambiance priority)"
		case 0x100:
			return "Daylight"
		case 0x200:
			return "Cloudy"
		case 0x300:
			return "Daylight Fluorescent"
		case 0x301:
			return "Day White Fluorescent"
		case 0x302:
			return "White Fluorescent"
		case 0x303:
			return "Warm White Fluorescent"
		case 0x304:
			return "Living Room Warm White Fluorescent"
		case 0x400:
			return "Incandescent"
		case 0x500:
			return "Flash"
		case 0x600:
			return "Underwater"
		case 0xf00:
			return "Custom"
		case 0xf01:
			return "Custom2"
		case 0xf02:
			return "Custom3"
		case 0xf03:
			return "Custom4"
		case 0xf04:
			return "Custom5"
		case 0xff0:
			return "Kelvin"
		default:
			return ""
		}
	}},
	0x1031: {Name: "Picture Mode", Fn: func(v interface{}) string {
		data := v.(int64)
		switch data {
		case 0x0:
			return "Auto"
		case 0x1:
			return "Portrait"
		case 0x2:
			return "Landscape"
		case 0x3:
			return "Macro"
		case 0x4:
			return "Sports"
		case 0x5:
			return "Night Scene"
		case 0x6:
			return "Program AE"
		case 0x7:
			return "Natural Light"
		case 0x8:
			return "Anti-blur"
		case 0x9:
			return "Beach & Snow"
		case 0xa:
			return "Sunset"
		case 0xb:
			return "Museum"
		case 0xc:
			return "Party"
		case 0xd:
			return "Flower"
		case 0xe:
			return "Text"
		case 0xf:
			return "Natural Light & Flash"
		case 0x10:
			return "Beach"
		case 0x11:
			return "Snow"
		case 0x12:
			return "Fireworks"
		case 0x13:
			return "Underwater"
		case 0x14:
			return "Portrait with Skin Correction"
		case 0x16:
			return "Panorama"
		case 0x17:
			return "Night (tripod)"
		case 0x18:
			return "Pro Low-light"
		case 0x19:
			return "Pro Focus"
		case 0x1a:
			return "Portrait 2"
		case 0x1b:
			return "Dog Face Detection"
		case 0x1c:
			return "Cat Face Detection"
		case 0x30:
			return "HDR"
		case 0x40:
			return "Advanced Filter"
		case 0x100:
			return "Aperture-priority AE"
		case 0x200:
			return "Shutter speed priority AE"
		case 0x300:
			return "Manual"
		default:
			return ""
		}
	}},
	0x1401: {Name: "Film Mode", Fn: func(v interface{}) string {
		data := v.(int64)
		switch data {
		case 0x0:
			return "F0/Standard (Provia)"
		case 0x100:
			return "F1/Studio Portrait"
		case 0x110:
			return "F1a/Studio Portrait Enhanced Saturation"
		case 0x120:
			return "F1b/Studio Portrait Smooth Skin Tone (Astia)"
		case 0x130:
			return "F1c/Studio Portrait Increased Sharpness"
		case 0x200:
			return "F2/Fujichrome (Velvia)"
		case 0x300:
			return "F3/Studio Portrait Ex"
		case 0x400:
			return "F4/Velvia"
		case 0x500:
			return "Pro Neg. Std"
		case 0x501:
			return "Pro Neg. Hi"
		case 0x600:
			return "Classic Chrome"
		case 0x700:
			return "Eterna"
		case 0x800:
			return "Classic Negative"
		case 0x900:
			return "Bleach Bypass"
		case 0xa00:
			return "Nostalgic Neg"
		default:
			return ""
		}
	}},
	0x3803: {Name: "Video Recording Mode", Fn: func(v interface{}) string {
		data := v.(int64)
		switch data {
		case 0x0:
			return "Normal"
		case 0x10:
			return "F-log"
		case 0x20:
			return "HLG"
		default:
			return ""
		}
	}},
	0x3806: {Name: "Video Compression", Fn: func(v interface{}) string {
		data := v.(int64)
		switch data {
		case 1:
			return "Log GOP"
		case 2:
			return "All Intra"
		default:
			return ""
		}
	}},
	0x3820: {Name: "Frame Rate"},
	0x3821: {Name: "Frame Width"},
	0x3822: {Name: "Frame Height"},
}

var canonExifTagDefinitionMap = map[uint16]*TagDefinition{
	0x1:  {Name: string(GroupCanonCameraSettings), SubTagDefinition: canonSubTagDefinitionMap[string(GroupCanonCameraSettings)]},
	0x4:  {Name: string(GroupCanonShotInfo), SubTagDefinition: canonSubTagDefinitionMap[string(GroupCanonShotInfo)]},
	0x6:  {Name: "Canon Image Type"},
	0x7:  {Name: "Canon Firmware Version"},
	0x10: {Name: "Canon Model ID"}, // Ref:https://exiftool.org/TagNames/Canon.html#CanonModelID
	0x38: {Name: "Battery Type", Fn: func(v interface{}) string {
		data := v.([]uint8)
		if len(data) == 76 {
			if bytes.Compare(data[:4], []byte{0x4c, 0, 0, 0}) == 0 {
				replacer := strings.NewReplacer(string(byte(0)), "", string(byte(1)), "")
				replacer.Replace(string(data[4:]))
				return replacer.Replace(string(data[4:]))
			}
		}
		return ""
	}},
	0x95: {Name: "Lens Model", Fn: func(v interface{}) string {
		data := v.(string)
		index := strings.Index(data, string([]byte{1, 1, 1, 1}))
		if index >= 0 {
			return strings.TrimSpace(data[:index])
		}
		return ""
	}},
	0x96:   {Name: "Internal Serial Number"},
	0xa0:   {Name: string(GroupCanonProcessingInfo), SubTagDefinition: canonSubTagDefinitionMap[string(GroupCanonProcessingInfo)]},
	0x4026: {Name: string(GroupCanonLogInfo), SubTagDefinition: canonSubTagDefinitionMap[string(GroupCanonLogInfo)]},
}

var canonSubTagDefinitionMap = map[string]*SubTagDefinition{
	string(GroupCanonCameraSettings): {tagDefinitionType: byIndex,
		subTagDefinitionMap: map[interface{}]*TagDefinition{
			7: {Name: "Focus Mode", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0:
					return "One-shot AF"
				case 1:
					return "AI Servo AF"
				case 2:
					return "AI Focus AF"
				case 3:
					return "Manual Focus (3)"
				case 4:
					return "Single"
				case 5:
					return "Continuous"
				case 6:
					return "Manual Focus (6)"
				case 16:
					return "Pan Focus"
				case 256:
					return "AF + MF"
				case 257:
					return "Live View"
				case 512:
					return "Movie Snap Focus"
				case 519:
					return "Movie Servo AF"
				default:
					return ""
				}
			}},
			9: {Name: "Record Mode", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 1:
					return "JPEG"
				case 2:
					return "CRW+THM"
				case 3:
					return "AVI+THM"
				case 4:
					return "TIF"
				case 5:
					return "TIF+JPEG"
				case 6:
					return "CR2"
				case 7:
					return "CR2+JPEG"
				case 9:
					return "MOV"
				case 10:
					return "MP4"
				case 11:
					return "CRM"
				case 12:
					return "CR3"
				case 13:
					return "CR3+JPEG"
				case 14:
					return "HIF"
				case 15:
					return "CR3+HIF"
				default:
					return ""
				}
			}},
			11: {Name: "Easy Mode", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0:
					return "Full auto"
				case 1:
					return "Manual"
				case 2:
					return "Landscape"
				case 3:
					return "Fast shutter"
				case 4:
					return "Slow shutter"
				case 5:
					return "Night"
				case 6:
					return "Gray Scale"
				case 7:
					return "Sepia"
				case 8:
					return "Portrait"
				case 9:
					return "Sports"
				case 10:
					return "Macro"
				case 11:
					return "Black & White"
				case 12:
					return "Pan focus"
				case 13:
					return "Vivid"
				case 14:
					return "Neutral"
				case 15:
					return "Flash Off"
				case 16:
					return "Long Shutter"
				case 17:
					return "Super Macro"
				case 18:
					return "Foliage"
				case 19:
					return "Indoor"
				case 20:
					return "Fireworks"
				case 21:
					return "Beach"
				case 22:
					return "Underwater"
				case 23:
					return "Snow"
				case 24:
					return "Kids & Pets"
				case 25:
					return "Night Snapshot"
				case 26:
					return "Digital Macro"
				case 27:
					return "My Colors"
				case 28:
					return "Movie Snap"
				case 29:
					return "Super Macro 2"
				case 30:
					return "Color Accent"
				case 31:
					return "Color Swap"
				case 32:
					return "Aquarium"
				case 33:
					return "ISO 3200"
				case 34:
					return "ISO 6400"
				case 35:
					return "Creative Light Effect"
				case 36:
					return "Easy"
				case 37:
					return "Quick Shot"
				case 38:
					return "Creative Auto"
				case 39:
					return "Zoom Blur"
				case 40:
					return "Low Light"
				case 41:
					return "Nostalgic"
				case 42:
					return "Super Vivid"
				case 43:
					return "Poster Effect"
				case 44:
					return "Face Self-timer"
				case 45:
					return "Smile"
				case 46:
					return "Wink Self-timer"
				case 47:
					return "Fisheye Effect"
				case 48:
					return "Miniature Effect"
				case 49:
					return "High-speed Burst"
				case 50:
					return "Best Image Selection"
				case 51:
					return "High Dynamic Range"
				case 52:
					return "Handheld Night Scene"
				case 53:
					return "Movie Digest"
				case 54:
					return "Live View Control"
				case 55:
					return "Discreet"
				case 56:
					return "Blur Reduction"
				case 57:
					return "Monochrome"
				case 58:
					return "Toy Camera Effect"
				case 59:
					return "Scene Intelligent Auto"
				case 60:
					return "High-speed Burst HQ"
				case 61:
					return "Smooth Skin"
				case 62:
					return "Soft Focus"
				case 68:
					return "Food"
				case 84:
					return "HDR Art Standard"
				case 85:
					return "HDR Art Vivid"
				case 93:
					return "HDR Art Bold"
				case 257:
					return "Spotlight"
				case 258:
					return "Night 2"
				case 259:
					return "Night+"
				case 260:
					return "Super Night"
				case 261:
					return "Sunset"
				case 263:
					return "Night Scene"
				case 264:
					return "Surface"
				case 265:
					return "Low Light 2"
				default:
					return ""
				}
			}},
			17: {Name: "Metering Mode", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0:
					return "Default"
				case 1:
					return "Spot"
				case 2:
					return "Average"
				case 3:
					return "Evaluative"
				case 4:
					return "Partial"
				case 5:
					return "Center-weighted average"
				default:
					return ""
				}
			}},
			18: {Name: "Focus Range", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0:
					return "Manual"
				case 1:
					return "Auto"
				case 2:
					return "Not Known"
				case 3:
					return "Macro"
				case 4:
					return "Very Close"
				case 5:
					return "Close"
				case 6:
					return "Middle Range"
				case 7:
					return "Far Range"
				case 8:
					return "Pan Focus"
				case 9:
					return "Super Macro"
				case 10:
					return "Infinity"
				default:
					return ""
				}
			}},
			19: {Name: "AF Point", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0:
					return "N/A"
				case 0x2005:
					return "Manual AF point selection"
				case 0x3000:
					return "None (MF)"
				case 0x3001:
					return "Auto AF point selection"
				case 0x3002:
					return "Right"
				case 0x3003:
					return "Center"
				case 0x3004:
					return "Left"
				case 0x4001:
					return "Auto AF point selection"
				case 0x4006:
					return "Face Detect"
				default:
					return ""
				}
			}},
			20: {Name: "Canon Exposure Mode", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0:
					return "Easy"
				case 1:
					return "Program AE"
				case 2:
					return "Shutter speed priority AE"
				case 3:
					return "Aperture-priority AE"
				case 4:
					return "Manual"
				case 5:
					return "Depth-of-field AE"
				case 6:
					return "M-Dep"
				case 7:
					return "Bulb"
				case 8:
					return "Flexible-priority AE"
				default:
					return ""
				}
			}},
			32: {Name: "Focus Continuous", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0:
					return "Single"
				case 1:
					return "Continuous"
				case 8:
					return "Manual"
				default:
					return ""
				}
			}},
			34: {Name: "Image Stabilization", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0:
					return "Off"
				case 1:
					return "On"
				case 2:
					return "Shoot Only"
				case 3:
					return "Panning"
				case 4:
					return "Dynamic"
				case 256:
					return "Off (2)"
				case 257:
					return "On (2)"
				case 258:
					return "Shoot Only (2)"
				case 259:
					return "Panning (2)"
				case 260:
					return "Dynamic (2)"
				default:
					return ""
				}
			}},
		},
	},
	string(GroupCanonShotInfo): {tagDefinitionType: byIndex,
		subTagDefinitionMap: map[interface{}]*TagDefinition{
			1: {Name: "Auto ISO", Fn: func(v interface{}) string {
				data := v.(int64)
				value := math.Exp(float64(data)/32*math.Log(2)) * 100
				return fmt.Sprintf("%.0f", value)
			}},
			2: {Name: "Base ISO", Fn: func(v interface{}) string {
				data := v.(int64)
				value := math.Exp(float64(data)/32*math.Log(2)) * 100 / 32
				return fmt.Sprintf("%.0f", value)
			}},
			7: {Name: "WhiteBalance", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0:
					return "Auto"
				case 1:
					return "Daylight"
				case 2:
					return "Cloudy"
				case 3:
					return "Tungsten"
				case 4:
					return "Fluorescent"
				case 5:
					return "Flash"
				case 6:
					return "Custom"
				case 7:
					return "Black & White"
				case 8:
					return "Shade"
				case 9:
					return "Manual Temperature (Kelvin)"
				case 10:
					return "PC Set1"
				case 11:
					return "PC Set2"
				case 12:
					return "PC Set3"
				case 14:
					return "Daylight Fluorescent"
				case 15:
					return "Custom 1"
				case 16:
					return "Custom 2"
				case 17:
					return "Underwater"
				case 18:
					return "Custom 3"
				case 19:
					return "Custom 4"
				case 20:
					return "PC Set4"
				case 21:
					return "PC Set5"
				case 23:
					return "Auto (ambience priority)"
				default:
					return ""
				}
			}},
			12: {Name: "Camera Temperature", Fn: func(v interface{}) string {
				data := v.(int64)
				if data == 0 {
					return "N/A"
				}
				return fmt.Sprintf("%d °C", data-128)
			}},
			28: {Name: "ND Filter", Fn: func(v interface{}) string {
				switch v.(int64) {
				case -1:
					return "n/a"
				case 0:
					return "Off"
				case 1:
					return "On"
				default:
					return ""
				}
			}},
		},
	},
	string(GroupCanonProcessingInfo): {tagDefinitionType: byIndex,
		subTagDefinitionMap: map[interface{}]*TagDefinition{
			9: {Name: "Color Temperature"},
			10: {Name: "Picture Style", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0x00:
					return "None"
				case 0x01:
					return "Standard"
				case 0x02:
					return "Portrait"
				case 0x03:
					return "High Saturation"
				case 0x04:
					return "Adobe RGB"
				case 0x05:
					return "Low Saturation"
				case 0x06:
					return "CM Set 1"
				case 0x07:
					return "CM Set 2"
				case 0x21:
					return "User Def. 1"
				case 0x22:
					return "User Def. 2"
				case 0x23:
					return "User Def. 3"
				case 0x41:
					return "PC 1"
				case 0x42:
					return "PC 2"
				case 0x43:
					return "PC 3"
				case 0x81:
					return "Standard"
				case 0x82:
					return "Portrait"
				case 0x83:
					return "Landscape"
				case 0x84:
					return "Neutral"
				case 0x85:
					return "Faithful"
				case 0x86:
					return "Monochrome"
				case 0x87:
					return "Auto"
				case 0x88:
					return "Fine Detail"
				case 0xff:
					return "n/a"
				case 0xffff:
					return "n/a"
				default:
					return ""
				}
			}},
		}},
	string(GroupCanonLogInfo): {tagDefinitionType: byIndex,
		subTagDefinitionMap: map[interface{}]*TagDefinition{
			4: {Name: "Compression Format", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0x00:
					return "Editing (ALL-I)"
				case 0x01:
					return "Standard (IPB)"
				case 0x02:
					return "Light (IPB)"
				case 0x03:
					return "Motion JPEG"
				case 0x04:
					return "RAW"
				default:
					return ""
				}
			}},
			9: {Name: "ColorSpace2", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0x00:
					return "BT.709"
				case 0x01:
					return "BT.2020"
				case 0x02:
					return "Cinema gamut"
				default:
					return ""
				}
			}},
			11: {Name: "Canon Log Version", Fn: func(v interface{}) string {
				switch v.(int64) {
				case 0x00:
					return "OFF"
				case 0x01:
					return "Canon Log"
				case 0x02:
					return "Canon Log 2"
				case 0x03:
					return "Canon Log 3"
				default:
					return ""
				}
			}},
		}},
}
