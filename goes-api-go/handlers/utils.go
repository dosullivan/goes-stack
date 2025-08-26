package handlers

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// extractTimestamp extracts a timestamp from an S3 object key
func extractTimestamp(objectKey string) (time.Time, error) {
	// Example: OR_ABI-L2-CMIPF-M6CFC_G16_s20243600000202_e20243600009510_c20243600009587.png
	//                                     ^-- start time (s)
	re := regexp.MustCompile(`_s(\d{4})(\d{3})(\d{6})\d`) // Match YYYY DDD HHMMSS d
	match := re.FindStringSubmatch(objectKey)
	if len(match) < 4 {
		return time.Time{}, fmt.Errorf("no timestamp found in filename")
	}

	year, _ := strconv.Atoi(match[1])     // YYYY
	doy, _ := strconv.Atoi(match[2])      // DDD (day of year)
	hour, _ := strconv.Atoi(match[3][:2]) // HH
	min, _ := strconv.Atoi(match[3][2:4]) // MM
	sec, _ := strconv.Atoi(match[3][4:])  // SS

	// Create time using day-of-year
	t := time.Date(year, time.January, 1, hour, min, sec, 0, time.UTC)
	t = t.AddDate(0, 0, doy-1) // subtract 1 because day-of-year is 1-based

	return t, nil
}
