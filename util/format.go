package util

import "time"

func GetZeroTime() time.Time {
	t, _ := time.Parse("2006-01-02 15:04:05", "1999-01-01 00:00:00")
	return t
}
