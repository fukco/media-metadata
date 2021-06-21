package sony

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

func (mode AutoWhiteBalanceMode)String() string  {
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

type CodingEquations []byte

// TODO Reference: Catalyst Browse(chinese:编程方程) not sure it's correct or not
func (ce CodingEquations)String() string {
	switch hex.EncodeToString(ce) {
	case "060e2b34040101010401010102020000":
		return "Rec.709"
	default:
		return "Unknown"
	}
}
