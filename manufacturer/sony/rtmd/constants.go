package rtmd

import "encoding/hex"

type AutoExposureMode []byte

func (mode AutoExposureMode) String() string {
	switch hex.EncodeToString(mode) {
	case "060e2b340401010b0510010101010000":
		return "Manual"
	case "060e2b340401010b0510010101020000":
		return "Auto"
	case "060e2b340401010b0510010101030000":
		return "GAIN"
	case "060e2b340401010b0510010101040000":
		return "A Mode"
	case "060e2b340401010b0510010101050000":
		return "S Mode"
	default:
		return "Camera specific control"
	}
}

type AFMode byte

func (mode AFMode) String() string {
	switch mode {
	case 0x00:
		return "MF"
	case 0x01:
		return "AF Center"
	case 0x02:
		return "AF Whole"
	case 0x03:
		return "AF Multi"
	case 0x04:
		return "AF Spot"
	case 0xff:
		return "Undefined"
	default:
		return "Reserved"
	}
}

type AutoWhiteBalanceMode byte

func (mode AutoWhiteBalanceMode) String() string {
	switch mode {
	case 0x00:
		return "Preset" // The WB is set to a fixed value
	case 0x01:
		return "Auto"
	case 0x02:
		return "Hold"
	case 0x03:
		return "One Push"
	case 0xff:
		return "Undefined"
	default:
		return "Reserved"
	}
}

type LightingPreset byte

func (mode LightingPreset) String() string {
	switch mode {
	case 0x01:
		return "Incandescent"
	case 0x02:
		return "Fluorescent"
	case 0x04:
		return "SunLight"
	case 0x05:
		return "Cloudy"
	case 0x06:
		return "Other"
	case 0x21:
		return "Custom"
	default:
		return "Reserved"
	}
}

type CodingEquations []byte

func (ce CodingEquations) String() string {
	switch hex.EncodeToString(ce) {
	case "060e2b34040101010401010102020000":
		return "Rec.709"
	case "060e2b340401010d0401010102060000":
		return "Rec.2020ncl"
	default:
		return "Unknown"
	}
}

type GammaEquation []byte

func (ge GammaEquation) String() string {
	switch hex.EncodeToString(ge) {
	case "060e2b34040101010401010101020000":
		return "rec709"
	case "060e2b34040101010401010101030000":
		return "SMPTE ST 240M"
	case "060e2b340401010d0401010101080000":
		return "rec709-xvycc"
	case "060e2b34040101060e06040101010301":
		return "Cine1"
	case "060e2b34040101060e06040101010302":
		return "Cine2"
	case "060e2b34040101060e06040101010303":
		return "Cine3"
	case "060e2b34040101060e06040101010304":
		return "Cine4"
	case "060e2b34040101060e06040101010508":
		return "S-Log2"
	case "060e2b34040101060e06040101010602":
		return "Still"
	case "060e2b34040101060e06040101010604":
		return "S-Log3"
	case "060e2b34040101060e06040101010605":
		return "S-Log3-Cine"
	case "060e2b34040101060e06040101010705":
		return "S-Cinetone"
	case "060e2b340401010d04010101010b0000":
		return "Rec2100-HLG"
	default:
		return "Gamma: Unkn/Custom"
	}
}

type ColorPrimaries []byte

func (cp ColorPrimaries) String() string {
	switch hex.EncodeToString(cp) {
	case "060e2b34040101060401010103030000":
		return "rec709"
	case "060e2b34040101060e06040101030103":
		return "S-Gamut"
	case "060e2b34040101060e06040101030104":
		return "S-Gamut3"
	case "060e2b34040101060e06040101030105":
		return "S-Gamut3.Cine"
	case "060e2b340401010d0401010103040000":
		return "rec2020"
	default:
		return "ColorSpace Unkn/Custom"
	}
}
