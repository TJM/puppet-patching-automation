package functions

import (
	"time"
)

const (
	// TimeFormatISO8601 - ISO8601 Time Format
	TimeFormatISO8601 = "2006-01-02 15:04:05"

	// TimeFormatISO8601NoSpace - ISO8601 Time Format with an underscore instead of space
	TimeFormatISO8601NoSpace = "2006-01-02_15:04:05"

	// TimeFormatDateTimeLocal - Time Format used by html datetime-local input type
	TimeFormatDateTimeLocal = "2006-01-02T15:04"
)

// Check : Error checking
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// FormatAsDateTimeLocal : Format local date and time
// 2006-01-02T15:04
func FormatAsDateTimeLocal(t time.Time) string {
	if t == (time.Time{}) { // Catch "zero" time value and return an empty string
		return ""
	}
	return t.Format(TimeFormatDateTimeLocal)
}

// FormatAsISO8601 : Format local date and time
// 2006-01-02 15:04:05
func FormatAsISO8601(t time.Time) string {
	if t == (time.Time{}) { // Catch "zero" time value and return an empty string
		return ""
	}
	return t.Format(TimeFormatISO8601)
}

// FormatAsISO8601NoSpace : Format local date and time
// 2006-01-02 15:04:05
func FormatAsISO8601NoSpace(t time.Time) string {
	if t == (time.Time{}) { // Catch "zero" time value and return an empty string
		return ""
	}
	return t.Format(TimeFormatISO8601NoSpace)
}

// Contains will determine whether a string is in a []string
func Contains(sl []string, name string) bool {
	for _, value := range sl {
		if value == name {
			return true
		}
	}
	return false
}
