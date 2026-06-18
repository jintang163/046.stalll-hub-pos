package nsq

import "time"

func GetCurrentTime() time.Time {
	return time.Now().Local()
}

func GetCurrentTimeString() string {
	return GetCurrentTime().Format("2006-01-02 15:04:05")
}

func GetCurrentDate() string {
	return GetCurrentTime().Format("2006-01-02")
}
