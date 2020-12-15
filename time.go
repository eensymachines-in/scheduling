package scheduling

import (
	"fmt"
	"time"
)

const (
	// I tried to change PM to AM in the format, and then it fails to read PM. but when PM in format, it reads both
	// https://stackoverflow.com/questions/44924628/golang-time-parse-1122-pm-something-that-i-can-do-math-with
	format = "03:04 PM"
	mdNt   = "12:00 AM"
)

// TimeStr : custom definition of time as string, indicated of the format above
type TimeStr string

// ToElapsedTm : just converts the string time to
func (ts TimeStr) ToElapsedTm() (int, error) {
	mdntTm, _ := time.Parse(format, mdNt) // getting the midnight time
	tm, err := time.Parse(format, string(ts))
	if err != nil {
		return -1, err
	}
	elapsed := tm.Sub(mdntTm).Seconds()
	return int(elapsed), nil
}

// TmStrFromUnixSecs : for the unix seconds given this can convert that into TimeStr
// this application uses specific format of clock so that its compatible to PArsing of time
func TmStrFromUnixSecs(elapsed int) TimeStr {
	hr, rem := elapsed/3600, elapsed%3600
	min := rem / 60
	// fmt.Printf("%d %d\n", hr, min)
	ampm := "AM"
	if hr >= 12 { // noon 12 is 12:01, while midnight 12 is 00:01
		hr = hr - 12
		ampm = "PM"
	}
	return TimeStr(fmt.Sprintf("%02d:%02d %s", hr, min, ampm))
}

// ElapsedSecondsNow : this can for any given day, calculate the seconds that have elapsed since midnight
func ElapsedSecondsNow() int {
	hr, min, sec := time.Now().Clock()
	return (hr * 3600) + (min * 60) + sec
}
