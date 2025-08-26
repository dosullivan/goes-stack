package handlers

import (
	"testing"
	"time"
)

func TestTimezoneLogic(t *testing.T) {
	// Test cases with UTC timestamps and expected CST dates
	tests := []struct {
		utcTimestamp    string
		expectedCSTDate string
	}{
		{"20241226000203", "2024-12-25"}, // 6:02:03 PM CST on Dec 25
		{"20241226030203", "2024-12-25"}, // 9:02:03 PM CST on Dec 25
		{"20241225230203", "2024-12-25"}, // 5:02:03 PM CST on Dec 25
	}

	cstLocation, _ := time.LoadLocation("America/Chicago")

	for _, tc := range tests {
		utcTime, err := time.Parse("20060102150405", tc.utcTimestamp)
		if err != nil {
			t.Errorf("Failed to parse UTC timestamp: %v", err)
			continue
		}

		cstTime := utcTime.In(cstLocation)
		cstDate := cstTime.Format("2006-01-02")

		if cstDate != tc.expectedCSTDate {
			t.Errorf("For UTC timestamp %s, expected CST date %s, but got %s",
				tc.utcTimestamp, tc.expectedCSTDate, cstDate)
		}
	}
}
