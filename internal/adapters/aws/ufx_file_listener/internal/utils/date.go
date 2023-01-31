package utils

import "time"

func ConvertToDate(date string) (time.Time, error) {
	// January 2nd
	layout := "2006-01-02"
	t, err := time.Parse(layout, date)
	return t, err
}
