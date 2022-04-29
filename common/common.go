package common

import "fmt"

func ConvertBitrate(input uint32) string {
	units := [3]string{"Kbps", "Mbps", "Gbps"}
	i := 0
	for ; i < len(units); i++ {
		if input >= 1000 {
			input /= 1000
		} else {
			break
		}
	}
	return fmt.Sprintf("%d%s", input, units[i])
}

func FormatFNumber(value string) string {
	return ""
}
