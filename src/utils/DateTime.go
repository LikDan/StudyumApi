package utils

import "time"

func Date() time.Time {
	return ToDateWithoutTime(time.Now())
}

func ToDateWithoutTime(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func GetTimeDuration(hour, minute int) time.Duration {
	return time.Duration(hour*60*60+minute*60) * time.Second
}
