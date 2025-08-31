package handlers

import (
	"encoding/csv"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// WeatherOffice represents a National Weather Service office
type WeatherOffice struct {
	City      string `json:"city"`
	State     string `json:"state"`
	CWA       string `json:"cwa"`       // County Warning Area code (3-letter)
	StationID string `json:"stationId"` // Full station ID (4-letter, starts with K or P)
}

// GetWeatherOffices returns all available NWS weather offices
func GetWeatherOffices() gin.HandlerFunc {
	return func(c *gin.Context) {
		offices, err := loadWeatherOffices()
		if err != nil {
			// Return hardcoded list if CSV fails to load
			offices = getDefaultOffices()
		}

		// Sort offices by state, then by city
		sort.Slice(offices, func(i, j int) bool {
			if offices[i].State != offices[j].State {
				return offices[i].State < offices[j].State
			}
			return offices[i].City < offices[j].City
		})

		c.JSON(http.StatusOK, gin.H{
			"offices": offices,
			"count":   len(offices),
		})
	}
}

func loadWeatherOffices() ([]WeatherOffice, error) {
	// Try to load from CSV file
	file, err := os.Open("office_names_and_codes.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var offices []WeatherOffice
	for {
		record, err := reader.Read()
		if err != nil {
			break // End of file or error
		}
		
		if len(record) >= 4 {
			offices = append(offices, WeatherOffice{
				City:      record[0],
				State:     record[1],
				CWA:       record[2],
				StationID: record[3],
			})
		}
	}

	return offices, nil
}

func getDefaultOffices() []WeatherOffice {
	// Hardcoded list of major weather offices as fallback
	return []WeatherOffice{
		{City: "Albuquerque", State: "New Mexico", CWA: "ABQ", StationID: "KABQ"},
		{City: "Anchorage", State: "Alaska", CWA: "AFC", StationID: "PAFC"},
		{City: "Albany", State: "New York", CWA: "ALY", StationID: "KALY"},
		{City: "Amarillo", State: "Texas", CWA: "AMA", StationID: "KAMA"},
		{City: "Atlanta", State: "Georgia", CWA: "FFC", StationID: "KFFC"},
		{City: "Birmingham", State: "Alabama", CWA: "BMX", StationID: "KBMX"},
		{City: "Bismarck", State: "North Dakota", CWA: "BIS", StationID: "KBIS"},
		{City: "Boise", State: "Idaho", CWA: "BOI", StationID: "KBOI"},
		{City: "Boston", State: "Massachusetts", CWA: "BOX", StationID: "KBOX"},
		{City: "Buffalo", State: "New York", CWA: "BUF", StationID: "KBUF"},
		{City: "Burlington", State: "Vermont", CWA: "BTV", StationID: "KBTV"},
		{City: "Charleston", State: "South Carolina", CWA: "CHS", StationID: "KCHS"},
		{City: "Charleston", State: "West Virginia", CWA: "RLX", StationID: "KRLX"},
		{City: "Chicago", State: "Illinois", CWA: "LOT", StationID: "KLOT"},
		{City: "Cleveland", State: "Ohio", CWA: "CLE", StationID: "KCLE"},
		{City: "Columbia", State: "South Carolina", CWA: "CAE", StationID: "KCAE"},
		{City: "Dallas/Fort Worth", State: "Texas", CWA: "FWD", StationID: "KFWD"},
		{City: "Denver", State: "Colorado", CWA: "BOU", StationID: "KBOU"},
		{City: "Des Moines", State: "Iowa", CWA: "DMX", StationID: "KDMX"},
		{City: "Detroit", State: "Michigan", CWA: "DTX", StationID: "KDTX"},
		{City: "Duluth", State: "Minnesota", CWA: "DLH", StationID: "KDLH"},
		{City: "Fairbanks", State: "Alaska", CWA: "AFG", StationID: "PAFG"},
		{City: "Grand Rapids", State: "Michigan", CWA: "GRR", StationID: "KGRR"},
		{City: "Gray", State: "Maine", CWA: "GYX", StationID: "KGYX"},
		{City: "Green Bay", State: "Wisconsin", CWA: "GRB", StationID: "KGRB"},
		{City: "Honolulu", State: "Hawaii", CWA: "HFO", StationID: "PHFO"},
		{City: "Houston", State: "Texas", CWA: "HGX", StationID: "KHGX"},
		{City: "Indianapolis", State: "Indiana", CWA: "IND", StationID: "KIND"},
		{City: "Jackson", State: "Mississippi", CWA: "JAN", StationID: "KJAN"},
		{City: "Jacksonville", State: "Florida", CWA: "JAX", StationID: "KJAX"},
		{City: "Kansas City", State: "Missouri", CWA: "EAX", StationID: "KEAX"},
		{City: "Key West", State: "Florida", CWA: "KEY", StationID: "KKEY"},
		{City: "Las Vegas", State: "Nevada", CWA: "VEF", StationID: "KVEF"},
		{City: "Little Rock", State: "Arkansas", CWA: "LZK", StationID: "KLZK"},
		{City: "Los Angeles", State: "California", CWA: "LOX", StationID: "KLOX"},
		{City: "Louisville", State: "Kentucky", CWA: "LMK", StationID: "KLMK"},
		{City: "Lubbock", State: "Texas", CWA: "LUB", StationID: "KLUB"},
		{City: "Memphis", State: "Tennessee", CWA: "MEG", StationID: "KMEG"},
		{City: "Miami", State: "Florida", CWA: "MFL", StationID: "KMFL"},
		{City: "Milwaukee", State: "Wisconsin", CWA: "MKX", StationID: "KMKX"},
		{City: "Minneapolis", State: "Minnesota", CWA: "MPX", StationID: "KMPX"},
		{City: "Nashville", State: "Tennessee", CWA: "OHX", StationID: "KOHX"},
		{City: "New Orleans", State: "Louisiana", CWA: "LIX", StationID: "KLIX"},
		{City: "New York", State: "New York", CWA: "OKX", StationID: "KOKX"},
		{City: "Norman", State: "Oklahoma", CWA: "OUN", StationID: "KOUN"},
		{City: "Omaha", State: "Nebraska", CWA: "OAX", StationID: "KOAX"},
		{City: "Philadelphia", State: "Pennsylvania", CWA: "PHI", StationID: "KPHI"},
		{City: "Phoenix", State: "Arizona", CWA: "PSR", StationID: "KPSR"},
		{City: "Pittsburgh", State: "Pennsylvania", CWA: "PBZ", StationID: "KPBZ"},
		{City: "Portland", State: "Maine", CWA: "GYX", StationID: "KGYX"},
		{City: "Portland", State: "Oregon", CWA: "PQR", StationID: "KPQR"},
		{City: "Raleigh", State: "North Carolina", CWA: "RAH", StationID: "KRAH"},
		{City: "Reno", State: "Nevada", CWA: "REV", StationID: "KREV"},
		{City: "Sacramento", State: "California", CWA: "STO", StationID: "KSTO"},
		{City: "Salt Lake City", State: "Utah", CWA: "SLC", StationID: "KSLC"},
		{City: "San Antonio", State: "Texas", CWA: "EWX", StationID: "KEWX"},
		{City: "San Diego", State: "California", CWA: "SGX", StationID: "KSGX"},
		{City: "San Francisco", State: "California", CWA: "MTR", StationID: "KMTR"},
		{City: "San Juan", State: "Puerto Rico", CWA: "SJU", StationID: "TJSJ"},
		{City: "Seattle", State: "Washington", CWA: "SEW", StationID: "KSEW"},
		{City: "Shreveport", State: "Louisiana", CWA: "SHV", StationID: "KSHV"},
		{City: "Sioux Falls", State: "South Dakota", CWA: "FSD", StationID: "KFSD"},
		{City: "Spokane", State: "Washington", CWA: "OTX", StationID: "KOTX"},
		{City: "St. Louis", State: "Missouri", CWA: "LSX", StationID: "KLSX"},
		{City: "Tallahassee", State: "Florida", CWA: "TAE", StationID: "KTAE"},
		{City: "Tampa", State: "Florida", CWA: "TBW", StationID: "KTBW"},
		{City: "Tucson", State: "Arizona", CWA: "TWC", StationID: "KTWC"},
		{City: "Tulsa", State: "Oklahoma", CWA: "TSA", StationID: "KTSA"},
		{City: "Washington", State: "District of Columbia", CWA: "LWX", StationID: "KLWX"},
	}
}

// Helper function to check if a station code matches an office
func IsOfficeStation(stationCode string, officeID string) bool {
	// Compare both 3-letter CWA codes and 4-letter station IDs
	stationUpper := strings.ToUpper(stationCode)
	officeUpper := strings.ToUpper(officeID)
	
	// Check if it's a full match or partial match
	return stationUpper == officeUpper || 
		   strings.HasSuffix(stationUpper, officeUpper) ||
		   strings.Contains(stationUpper, officeUpper)
}