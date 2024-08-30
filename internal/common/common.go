package common

import (
	"fmt"
	"time"
)

const (
	secondsPerMinute       = 60
	secondsPerHour         = 60 * secondsPerMinute
	secondsPerDay          = 24 * secondsPerHour
	mp4EpochToUnix   int64 = ((1903*365 + 1903/4 - 1903/100 + 1903/400) - (1969*365 + 1969/4 - 1969/100 + 1969/400)) * secondsPerDay
)

func Mp4Epoch(sec int64) *time.Time {
	t := time.Unix(sec+mp4EpochToUnix, 0)
	return &t
}

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
