package datetime

import (
	"strconv"
	"time"
)

func Date() time.Time {
	return ToDateWithoutTime(time.Now())
}

func ToDateWithoutTime(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func FormatDuration(d time.Duration) string {
	return strconv.Itoa(int(d.Hours())) + ":" + strconv.Itoa(int(d.Minutes()))
}
