package datetime

import (
	"github.com/pkg/errors"
	"strconv"
	"time"
)

var DurationError = errors.New("Duration error")

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

func ParseDuration(str string) (time.Duration, error) {
	if len(str) != 5 {
		return 0, errors.Wrap(DurationError, "length not equal 5")
	}

	return time.ParseDuration(str[:2] + "h" + str[3:] + "m")
}
