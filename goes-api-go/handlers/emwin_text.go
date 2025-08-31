package handlers

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"goes-api-go/s3"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

// EMWINTextFile represents a text weather bulletin
type EMWINTextFile struct {
	URL         string    `json:"url"`
	Timestamp   time.Time `json:"timestamp"`
	Filename    string    `json:"filename"`
	ProductCode string    `json:"productCode,omitempty"`
	Station     string    `json:"station,omitempty"`
	Description string    `json:"description,omitempty"`
}

// EMWINTextCategory represents categories of EMWIN text products
type EMWINTextCategory struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Examples    []string `json:"examples,omitempty"`
}

// GetEMWINTextCategories returns available EMWIN text product categories
func GetEMWINTextCategories() gin.HandlerFunc {
	return func(c *gin.Context) {
		categories := []EMWINTextCategory{
			{
				Key:         "weather_warnings",
				Title:       "Weather Warnings & Watches",
				Description: "Severe weather warnings, watches, and advisories",
				Examples:    []string{"tornado warnings", "severe thunderstorm warnings", "winter storm watches"},
			},
			{
				Key:         "forecasts",
				Title:       "Weather Forecasts",
				Description: "Zone forecasts, marine forecasts, and aviation forecasts",
				Examples:    []string{"zone forecasts", "marine forecasts", "TAF", "METAR"},
			},
			{
				Key:         "observations",
				Title:       "Weather Observations",
				Description: "Surface observations, upper air data, and marine observations",
				Examples:    []string{"hourly surface obs", "RAOB", "buoy data"},
			},
			{
				Key:         "discussions",
				Title:       "Forecast Discussions",
				Description: "Area forecast discussions and technical bulletins",
				Examples:    []string{"AFD", "technical discussions", "model guidance"},
			},
			{
				Key:         "climate",
				Title:       "Climate & Records",
				Description: "Climate summaries, daily records, and monthly data",
				Examples:    []string{"daily climate", "monthly summaries", "record data"},
			},
			{
				Key:         "marine",
				Title:       "Marine Products",
				Description: "Marine forecasts, warnings, and observations",
				Examples:    []string{"marine forecasts", "coastal warnings", "wave data"},
			},
			{
				Key:         "aviation",
				Title:       "Aviation Products",
				Description: "Aviation forecasts, METARs, TAFs, and pilot reports",
				Examples:    []string{"METAR", "TAF", "PIREP", "SIGMET"},
			},
			{
				Key:         "fire_weather",
				Title:       "Fire Weather",
				Description: "Fire weather forecasts and red flag warnings",
				Examples:    []string{"fire weather forecasts", "red flag warnings", "spot forecasts"},
			},
		}

		c.JSON(http.StatusOK, gin.H{"categories": categories})
	}
}

// GetEMWINTextFiles returns EMWIN text files, optionally filtered by category, date, station, and office
func GetEMWINTextFiles(s3Client *s3.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Query("category")
		dateStr := c.Query("date")
		station := c.Query("station")
		office := c.Query("office")

		ctx := context.Background()
		cstLocation, err := time.LoadLocation("America/Chicago")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load timezone"})
			return
		}

		var files []EMWINTextFile

		// Use the single bucket with EMWIN path (adjust for current structure)
		bucketName := s3Client.BucketName
		emwinPath := "emwin/"  // EMWIN files are directly in date folders
		
		// If date specified, search that date, otherwise search recent files
		if dateStr != "" {
			files, err = getEMWINTextFilesForDate(ctx, s3Client, bucketName, emwinPath, dateStr, cstLocation)
		} else {
			// Get files from the last 3 days
			files, err = getRecentEMWINTextFiles(ctx, s3Client, bucketName, emwinPath, cstLocation, 3)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Apply category filter
		if category != "" && category != "all" {
			files = filterEMWINTextByCategory(files, category)
		}

		if station != "" {
			files = filterEMWINTextByStation(files, station)
		}
		
		if office != "" {
			files = filterEMWINTextByOffice(files, office)
		}

		// Apply limit
		if len(files) > 100 {
			files = files[:100]
		}

		c.JSON(http.StatusOK, gin.H{
			"files": files,
			"count": len(files),
			"filters": gin.H{
				"category": category,
				"station":  station,
				"date":     dateStr,
			},
		})
	}
}

// GetEMWINTextContent returns the content of a specific EMWIN text file
func GetEMWINTextContent(s3Client *s3.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		objectKey := c.Query("key")
		if objectKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Object key is required"})
			return
		}

		ctx := context.Background()
		bucketName := s3Client.BucketName

		// Get the object from MinIO
		object, err := s3Client.Client.GetObject(ctx, bucketName, objectKey, minio.GetObjectOptions{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve file"})
			return
		}
		defer object.Close()

		// Read content
		buf := make([]byte, 64*1024) // 64KB buffer
		n, err := object.Read(buf)
		if err != nil && n == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
			return
		}

		content := string(buf[:n])

		c.JSON(http.StatusOK, gin.H{
			"objectKey": objectKey,
			"content":   content,
		})
	}
}

func getEMWINTextFilesForDate(ctx context.Context, s3Client *s3.S3Client, bucketName, emwinPath, dateStr string, cstLocation *time.Location) ([]EMWINTextFile, error) {
	cstMidnight, err := time.ParseInLocation("2006-01-02", dateStr, cstLocation)
	if err != nil {
		return nil, fmt.Errorf("invalid date format")
	}

	utcStart := cstMidnight.UTC()
	utcEnd := cstMidnight.Add(24 * time.Hour).UTC()
	utcDates := []string{
		utcStart.Format("2006-01-02"),
		utcEnd.Format("2006-01-02"),
	}

	var files []EMWINTextFile

	// Check both the expected date folders and 1969-12-31 (where misfiled items go)
	prefixes := []string{}
	for _, utcDate := range utcDates {
		prefixes = append(prefixes, emwinPath + utcDate + "/")
	}
	// Also check 1969-12-31 folder for files with parsing issues
	prefixes = append(prefixes, emwinPath + "1969-12-31/")

	for _, prefix := range prefixes {
		objectCh := s3Client.Client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		})

		for object := range objectCh {
			if object.Err != nil {
				continue // Skip errors for individual prefixes
			}

			// Only process .TXT files
			if !strings.HasSuffix(strings.ToUpper(object.Key), ".TXT") {
				continue
			}

			// Try to parse timestamp from filename (EMWIN files have timestamps in names)
			timestamp := parseEMWINTimestamp(object.Key)
			if timestamp.IsZero() {
				// If we can't parse from filename, use LastModified
				timestamp = object.LastModified
			}

			cstTime := timestamp.In(cstLocation)
			// Check if this file's actual timestamp matches the requested date
			if cstTime.Format("2006-01-02") == dateStr {
				files = append(files, EMWINTextFile{
					URL:         s3Client.BaseURL + object.Key,
					Timestamp:   cstTime,
					Filename:    object.Key[strings.LastIndex(object.Key, "/")+1:],
					ProductCode: extractProductCode(object.Key),
					Station:     extractStation(object.Key),
					Description: getProductDescription(object.Key),
				})
			}
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Timestamp.After(files[j].Timestamp)
	})

	return files, nil
}

func getRecentEMWINTextFiles(ctx context.Context, s3Client *s3.S3Client, bucketName, emwinPath string, cstLocation *time.Location, maxDaysBack int) ([]EMWINTextFile, error) {
	var files []EMWINTextFile
	nowCST := time.Now().In(cstLocation)

	// Check recent date folders
	for i := 0; i < maxDaysBack; i++ {
		checkDate := nowCST.AddDate(0, 0, -i).Format("2006-01-02")
		prefix := emwinPath + checkDate + "/"

		objectCh := s3Client.Client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		})

		for object := range objectCh {
			if object.Err != nil {
				continue
			}

			// Only process .TXT files
			if !strings.HasSuffix(strings.ToUpper(object.Key), ".TXT") {
				continue
			}

			timestamp := parseEMWINTimestamp(object.Key)
			if timestamp.IsZero() {
				timestamp = object.LastModified
			}

			cstTime := timestamp.In(cstLocation)
			files = append(files, EMWINTextFile{
				URL:         s3Client.BaseURL + object.Key,
				Timestamp:   cstTime,
				Filename:    object.Key[strings.LastIndex(object.Key, "/")+1:],
				ProductCode: extractProductCode(object.Key),
				Station:     extractStation(object.Key),
				Description: getProductDescription(object.Key),
			})
		}
	}

	// Also check 1969-12-31 folder for recent misfiled items
	prefix := emwinPath + "1969-12-31/"
	objectCh := s3Client.Client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	cutoffTime := nowCST.AddDate(0, 0, -maxDaysBack)
	for object := range objectCh {
		if object.Err != nil {
			continue
		}

		// Only process .TXT files
		if !strings.HasSuffix(strings.ToUpper(object.Key), ".TXT") {
			continue
		}

		timestamp := parseEMWINTimestamp(object.Key)
		if timestamp.IsZero() {
			timestamp = object.LastModified
		}

		cstTime := timestamp.In(cstLocation)
		// Only include files from the last maxDaysBack days
		if cstTime.After(cutoffTime) {
			files = append(files, EMWINTextFile{
				URL:         s3Client.BaseURL + object.Key,
				Timestamp:   cstTime,
				Filename:    object.Key[strings.LastIndex(object.Key, "/")+1:],
				ProductCode: extractProductCode(object.Key),
				Station:     extractStation(object.Key),
				Description: getProductDescription(object.Key),
			})
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].Timestamp.After(files[j].Timestamp)
	})

	return files, nil
}

// parseEMWINTimestamp tries to extract timestamp from EMWIN filename format
func parseEMWINTimestamp(filename string) time.Time {
	// EMWIN files often have format: A_PRODUCT_TIMESTAMP_C_KWIN_...
	// Try to find and parse the timestamp portion
	parts := strings.Split(filename, "_")
	for _, part := range parts {
		// Look for timestamp format like 20250825033003 (YYYYMMDDHHMMSS)
		if len(part) == 14 && isNumeric(part) {
			if t, err := time.Parse("20060102150405", part); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}

// extractProductCode extracts the product code from EMWIN filename
func extractProductCode(filename string) string {
	// Example: A_ASUS41KGYX250310_C_KWIN_20250825031112_921174-3-RWRGYXME.TXT
	// Extract ASUS41 part
	parts := strings.Split(filename, "_")
	if len(parts) > 1 {
		productPart := parts[1]
		// Product codes are typically 6 characters
		if len(productPart) >= 6 {
			return productPart[:6]
		}
	}
	return ""
}

// extractStation extracts the station identifier (CCCC) from EMWIN filename
func extractStation(filename string) string {
	// EMWIN File format: (A/Z)_TTAAiiCCCCYYGGgg[BBB]_C_KWIN_yyyyMMddhhmmss_nnnnnn-p-NNNxxxqq.(TXT/ZIS/GIF/PNG/JPG)
	// Where:
	// - TTAAii is the 6-character WMO product type (T1T2A1A2ii)
	// - CCCC is the 4-character station identifier (position 6-10 in the WMO identifier)
	// - YYGGgg is the date/time group
	// - [BBB] is optional amendment indicator
	// Example: A_WFUS54KBMX260104_C_KWIN_20160126010535_000179-1-TORBMXAL.TXT
	//          Station is KBMX (Birmingham, AL)
	
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return ""
	}
	
	// Get the WMO product identifier (second part after pflag)
	wmoProduct := parts[1]
	
	// CCCC (station code) is always at position 6-10 (0-indexed)
	if len(wmoProduct) >= 10 {
		station := wmoProduct[6:10]
		// Station codes are 4 uppercase letters, typically starting with:
		// K - Continental US
		// P - Pacific/Alaska (PAFC, PAFG, PAJK, PHFO, etc.)
		// T - Caribbean (TJSJ for San Juan)
		// C - Canada (though less common in EMWIN)
		return station
	}
	
	return ""
}

// getProductDescription provides human-readable descriptions for product codes
func getProductDescription(filename string) string {
	// Extract the 3-letter product code from the filename
	// The product code appears after TTAAii in the WMO header
	// For example, in A_FXUS65KABQ... the product code would be derived from FXUS
	
	parts := strings.Split(filename, "_")
	if len(parts) < 2 {
		return "Weather Product"
	}
	
	wmoProduct := parts[1]
	if len(wmoProduct) < 4 {
		return "Weather Product"
	}
	
	// Extract T1T2 (first 2 chars) which helps identify the category
	t1t2 := wmoProduct[0:2]
	
	// Try to find product code in the filename (usually in the last part)
	// Format: nnnnnn-p-NNNxxxqq where NNN is the product category
	productCode := ""
	if len(parts) >= 6 {
		freeFormat := parts[5] // nnnnnn-p-NNNxxxqq part
		dashParts := strings.Split(freeFormat, "-")
		if len(dashParts) >= 3 {
			lastPart := dashParts[2] // NNNxxxqq part
			if len(lastPart) >= 3 {
				productCode = strings.ToUpper(lastPart[0:3])
			}
		}
	}
	
	// Map common product codes to descriptions
	productDescriptions := map[string]string{
		"ABV": "Rawinsonde Data Above 100 Millibars",
		"ADA": "Alarm/Alert Administrative Message",
		"ADM": "Alert Administrative Message",
		"ADR": "NWS Administrative Message",
		"ADV": "Generic Space Environment Advisory",
		"AFD": "Area Forecast Discussion",
		"AFM": "Area Forecast Matrices",
		"AFP": "Area Forecast Product",
		"AFW": "Fire Weather Matrix",
		"AGF": "Agricultural Forecast",
		"AGO": "Agricultural Observations",
		"ALT": "Space Environment Alert",
		"AQA": "Air Quality Alert",
		"AQI": "Air Quality Index Statement",
		"ASA": "Air Stagnation Advisory",
		"AVA": "Avalanche Watch",
		"AVG": "Avalanche Weather Guidance",
		"AVW": "Avalanche Warning",
		"AWO": "Area Weather Outlook",
		"AWS": "Area Weather Summary",
		"AWU": "Area Weather Update",
		"AWW": "Airport Weather Warning",
		"BOY": "Buoy Report",
		"BRG": "Coast Guard Observations",
		"CAE": "Child Abduction Emergency",
		"CCF": "Coded City Forecast",
		"CDW": "Civil Danger Warning",
		"CEM": "Civil Emergency Message",
		"CF6": "WFO Monthly/Daily Climate Data",
		"CFP": "Convective Forecast Product",
		"CFW": "Coastal Flood Warnings/Watches/Statements",
		"CGR": "Coast Guard Surface Report",
		"CHG": "Computer Hurricane Guidance",
		"CLI": "Climatological Report (Daily)",
		"CLM": "Climatological Report (Monthly)",
		"CWA": "Center Weather Advisory",
		"CWF": "Coastal Waters Forecast",
		"CWS": "Center Weather Statement",
		"DSA": "Unnumbered Depression / Suspicious Area Advisory",
		"DSM": "ASOS Daily Summary",
		"DSW": "Dust Storm Warning and Dust Advisory",
		"EFP": "3 To 5 Day Extended Forecast",
		"EQI": "Tsunami Bulletin",
		"EQR": "Earthquake Report",
		"EQW": "Earthquake Warning",
		"ESF": "Flood Potential Outlook",
		"EVI": "Evacuation Immediate",
		"EWW": "Extreme Wind Warning",
		"FFA": "Flash Flood Watch",
		"FFG": "Flash Flood Guidance",
		"FFH": "Headwater Guidance",
		"FFS": "Flash Flood Statement",
		"FFW": "Flash Flood Warning",
		"FLS": "Flood Statement",
		"FLW": "Flood Warning",
		"FRW": "Fire Warning",
		"FWA": "Fire Weather Administrative Message",
		"FWD": "Fire Weather Outlook Discussion",
		"FWF": "Routine Fire Weather Forecast",
		"FWL": "Land Management Forecasts",
		"FWM": "Miscellaneous Fire Weather Product",
		"FWN": "Fire Weather Notification",
		"FWO": "Fire Weather Observation",
		"FWS": "Spot Forecast",
		"GLF": "Great Lakes Forecast",
		"GLS": "Great Lakes Storm Summary",
		"HLS": "Hurricane Local Statement",
		"HMD": "Hydrometeorological Discussion",
		"HMW": "Hazardous Materials Warning",
		"HSF": "High Seas Forecast",
		"HWO": "Hazardous Weather Outlook",
		"HWR": "Hourly Weather Roundup",
		"ICE": "Ice Forecast",
		"LAE": "Local Area Emergency",
		"LCD": "Preliminary Local Climatological Data",
		"LEW": "Law Enforcement Warning",
		"LFP": "Local Forecast",
		"LSR": "Local Storm Report",
		"MAN": "Rawinsonde Observation Mandatory Levels",
		"MAP": "Mean Areal Precipitation",
		"MAW": "Amended Marine Forecast",
		"MFM": "Marine Forecast Matrix",
		"MIS": "Miscellaneous Local Product",
		"MSM": "ASOS Monthly Summary Message",
		"MTR": "METAR Formatted Surface Weather Observation",
		"MWS": "Marine Weather Statement",
		"MWW": "Marine Weather Message",
		"NOW": "Short Term Forecast",
		"NPW": "Non-Precipitation Warnings/Watches/Advisories",
		"NSH": "Nearshore Marine Forecast",
		"NUW": "Nuclear Power Plant Warning",
		"NWR": "NOAA Weather Radio Forecast",
		"OFF": "Offshore Forecast",
		"PFM": "Point Forecast Matrices",
		"PFW": "Fire Weather Point Forecast Matrices",
		"PLS": "Plain Language Ship Report",
		"PMD": "Prognostic Meteorological Discussion",
		"PNS": "Public Information Statement",
		"PSH": "Post Storm Hurricane Report",
		"PWO": "Public Severe Weather Outlook",
		"PWS": "Tropical Cyclone Probabilities",
		"QPF": "Quantitative Precipitation Forecast",
		"QPS": "Quantitative Precipitation Statement",
		"REC": "Recreational Report",
		"RER": "Record Report",
		"RFD": "Rangeland Fire Danger Forecast",
		"RFW": "Red Flag Warning",
		"RHW": "Radiological Hazard Warning",
		"RMT": "Required Monthly Test",
		"RNS": "Rain Information Statement",
		"RRM": "Miscellaneous Hydrologic Data",
		"RVA": "River Summary",
		"RVD": "Daily River Forecasts",
		"RVF": "River Forecast",
		"RVI": "River Ice Statement",
		"RVM": "Miscellaneous River Product",
		"RVR": "River Recreation Statement",
		"RVS": "River Statement",
		"RWR": "Regional Weather Roundup",
		"RWS": "Regional Weather Summary",
		"RWT": "Required Weekly Test",
		"SAB": "Special Avalanche Bulletin",
		"SAF": "Special Agricultural Weather Forecast/Advisory",
		"SAG": "Snow Avalanche Guidance",
		"SAW": "Preliminary Notice of Watch & Cancellation Message",
		"SCC": "Storm Summary",
		"SCD": "Supplementary Climatological Data",
		"SCP": "Satellite Cloud Product",
		"SCS": "Selected Cities Summary",
		"SDO": "Supplementary Data Observation",
		"SDS": "Special Dispersion Statement",
		"SEL": "Severe Local Storm Watch and Watch Cancellation",
		"SEV": "SPC Watch Point Information Message",
		"SFP": "State Forecast",
		"SFT": "Tabular State Forecast",
		"SGL": "Rawinsonde Observation Significant Levels",
		"SHP": "Surface Ship Report at Synoptic Time",
		"SIG": "International Sigmet / Convective Sigmet",
		"SIM": "Satellite Interpretation Message",
		"SLS": "Severe Local Storm Watch and Areal Outline",
		"SMF": "Smoke Management Weather Forecast",
		"SMW": "Special Marine Warning",
		"SOB": "SOB Observation",
		"SPS": "Special Weather Statement",
		"SPW": "Shelter in Place Warning",
		"SQW": "Snow Squall Warning",
		"SRF": "Surf Forecast",
		"SRG": "Soaring Guidance",
		"SSI": "Water Supply Outlook",
		"SSM": "Main Synoptic Hour Surface Observation",
		"STA": "Network and Severe Weather Statistical Summaries",
		"STD": "Satellite Tropical Disturbance Summary",
		"STO": "Road Condition Reports",
		"STP": "State Max/Min Temperature and Precipitation Table",
		"STQ": "Spot Forecast Request",
		"SUM": "Space Weather Message",
		"SVR": "Severe Thunderstorm Warning",
		"SVS": "Severe Weather Statement",
		"SWO": "Severe Storm Outlook Narrative",
		"SWS": "State Weather Summary",
		"SYN": "Regional Weather Synopsis",
		"TAF": "Terminal Aerodrome Forecast",
		"TAV": "Travelers Forecast Table",
		"TCA": "Aviation Tropical Cyclone Advisory",
		"TCD": "Tropical Cyclone Discussion",
		"TCE": "Tropical Cyclone Position Estimate",
		"TCM": "Marine/Aviation Tropical Cyclone Advisory",
		"TCP": "Public Tropical Cyclone Advisory",
		"TCS": "Satellite Tropical Cyclone Summary",
		"TCU": "Tropical Cyclone Update",
		"TCV": "Tropical Cyclone Watch/Warning Break Points",
		"TID": "Tide Report",
		"TMA": "Tsunami Tide/Seismic Message Acknowledgement",
		"TOE": "911 Telephone Outage Emergency",
		"TOR": "Tornado Warning",
		"TPT": "Temperature Precipitation Table",
		"TSU": "Tsunami Watch/Warning",
		"TUV": "Weather Bulletin",
		"TVL": "Travelers Forecast",
		"TWD": "Tropical Weather Discussion",
		"TWO": "Tropical Weather Outlook and Summary",
		"TWS": "Tropical Weather Summary",
		"URN": "Aircraft Reconnaissance",
		"UVI": "Ultraviolet Index",
		"VAA": "Volcanic Activity Advisory",
		"VER": "Forecast Verification Statistics",
		"VFT": "Terminal Aerodrome Forecast (TAF) Verification",
		"VOW": "Volcano Warning",
		"WA0": "Airmet (Pacific)",
		"WA1": "Airmet (Northeast)",
		"WA2": "Airmet (Southeast)",
		"WA3": "Airmet (North Central)",
		"WA4": "Airmet (South Central)",
		"WA5": "Airmet (Rocky Mountains)",
		"WA6": "Airmet (West Coast)",
		"WA7": "Airmet (Juneau, AK)",
		"WA8": "Airmet (Anchorage, AK)",
		"WA9": "Airmet (Fairbanks, AK)",
		"WAR": "Space Weather Advisory",
		"WAT": "Space Weather Watch",
		"WCN": "Weather Watch Clearance Notification",
		"WCR": "Weekly Weather and Crop Report",
		"WDA": "Weekly Data for Agriculture",
		"WDU": "Warning Decision Update",
		"WEK": "Routine Space Environment Product",
		"WOU": "Tornado/Severe Thunderstorm Watch",
		"WRK": "Work File",
		"WST": "Tropical Cyclone Sigmet",
		"WSV": "Volcanic Activity Sigmet",
		"WSW": "Winter Storm Warning",
		"WWA": "Watch Status Report",
		"WWP": "Severe Thunderstorm/Tornado Watch Probabilities",
		"ZFP": "Zone Forecast Product",
	}
	
	// Check if we found a product code in the filename
	if productCode != "" {
		if desc, exists := productDescriptions[productCode]; exists {
			return desc
		}
	}
	
	// Fall back to WMO T1T2 categories if no specific product code found
	wmoCategories := map[string]string{
		"FX": "Area Forecast Discussion",
		"FO": "Forecast Products",
		"FP": "Zone Forecast",
		"FZ": "Zone/Coastal Forecast",
		"SA": "Surface Observations",
		"SP": "Special Surface Observations",
		"FT": "Terminal Aerodrome Forecast",
		"NW": "Warnings/Watches/Advisories",
		"WW": "Warnings/Watches",
		"WF": "Fire Weather",
		"WS": "Sigmet",
		"AC": "Convective Outlook",
		"AS": "Area Forecast Discussion",
		"CF": "Coastal Forecast",
		"RR": "Hydrologic Data",
		"RV": "River Products",
	}
	
	if desc, exists := wmoCategories[t1t2]; exists {
		return desc
	}
	
	return "Weather Product"
}

func filterEMWINTextByCategory(files []EMWINTextFile, category string) []EMWINTextFile {
	var filtered []EMWINTextFile
	
	for _, file := range files {
		// Extract the actual three-letter product code from filename
		actualCode := extractActualProductCode(file.Filename)
		
		switch category {
		case "weather_warnings":
			// Check actual product code first, then WMO code
			if isWarningProductCode(actualCode) || isWarningProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		case "forecasts":
			if isForecastProductCode(actualCode) || isForecastProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		case "observations":
			if isObservationProductCode(actualCode) || isObservationProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		case "discussions":
			if isDiscussionProductCode(actualCode) || isDiscussionProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		case "marine":
			if isMarineProductCode(actualCode) || isMarineProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		case "aviation":
			if isAviationProductCode(actualCode) || isAviationProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		default:
			filtered = append(filtered, file)
		}
	}
	
	return filtered
}

// extractActualProductCode extracts the three-letter product code from filename
func extractActualProductCode(filename string) string {
	// Format: nnnnnn-p-PPPxxxqq where PPP is the 3-letter product code
	parts := strings.Split(filename, "-")
	if len(parts) >= 3 {
		lastPart := parts[len(parts)-1]
		// Remove file extension
		lastPart = strings.TrimSuffix(lastPart, ".TXT")
		lastPart = strings.TrimSuffix(lastPart, ".txt")
		// The product code is the first 3 letters
		if len(lastPart) >= 3 {
			return strings.ToUpper(lastPart[:3])
		}
	}
	return ""
}

// Helper functions to check actual three-letter product codes
func isWarningProductCode(code string) bool {
	warningCodes := map[string]bool{
		"TOR": true, // Tornado Warning
		"SVR": true, // Severe Thunderstorm Warning
		"SVS": true, // Severe Weather Statement
		"FFW": true, // Flash Flood Warning
		"FLW": true, // Flood Warning
		"WSW": true, // Winter Weather Warnings
		"NPW": true, // Non-Precipitation Warnings
		"EWW": true, // Extreme Wind Warning
		"CFW": true, // Coastal Flood Warning
		"SMW": true, // Special Marine Warning
		"MWW": true, // Marine Weather Warning
		"FRW": true, // Fire Warning
		"AVW": true, // Avalanche Warning
		"DSW": true, // Dust Storm Warning
		"SQW": true, // Snow Squall Warning
		"EQW": true, // Earthquake Warning
		"VOW": true, // Volcano Warning
		"HMW": true, // Hazardous Materials Warning
		"LEW": true, // Law Enforcement Warning
		"LAE": true, // Local Area Emergency
		"CDW": true, // Civil Danger Warning
		"CEM": true, // Civil Emergency Message
		"EVI": true, // Evacuation Immediate
		"SPW": true, // Shelter in Place Warning
		"NUW": true, // Nuclear Power Plant Warning
		"RHW": true, // Radiological Hazard Warning
		"FFA": true, // Flash Flood Watch
		"AVA": true, // Avalanche Watch
		"WOU": true, // Tornado/Severe Thunderstorm Watch
		"TSU": true, // Tsunami Watch/Warning
	}
	return warningCodes[code]
}

func isForecastProductCode(code string) bool {
	forecastCodes := map[string]bool{
		"ZFP": true, // Zone Forecast Product
		"AFD": true, // Area Forecast Discussion  
		"NOW": true, // Short Term Forecast
		"SFP": true, // State Forecast
		"PFM": true, // Point Forecast Matrices
		"AFM": true, // Area Forecast Matrices
		"EFP": true, // Extended Forecast
		"FWF": true, // Fire Weather Forecast
		"CWF": true, // Coastal Waters Forecast
		"GLF": true, // Great Lakes Forecast
		"OFF": true, // Offshore Forecast
		"NSH": true, // Nearshore Marine Forecast
		"TAF": true, // Terminal Aerodrome Forecast
		"MFM": true, // Marine Forecast Matrix
		"RFD": true, // Rangeland Fire Danger Forecast
		"LFP": true, // Local Forecast
		"TAV": true, // Travelers Forecast Table
		"TVL": true, // Travelers Forecast
	}
	return forecastCodes[code]
}

func isObservationProductCode(code string) bool {
	obsCodes := map[string]bool{
		"RWR": true, // Regional Weather Roundup
		"RWS": true, // Regional Weather Summary
		"HWR": true, // Hourly Weather Roundup
		"MTR": true, // METAR Surface Weather Observation
		"SYN": true, // Regional Weather Synopsis
		"AWS": true, // Area Weather Summary
		"CLI": true, // Climatological Report
		"LCD": true, // Local Climatological Data
		"LSR": true, // Local Storm Report
		"DSM": true, // ASOS Daily Summary
		"MSM": true, // ASOS Monthly Summary
		"BOY": true, // Buoy Report
		"CGR": true, // Coast Guard Surface Report
		"SHP": true, // Surface Ship Report
		"TID": true, // Tide Report
		"HYD": true, // Hydrometeorological Products
		"OBS": true, // Observations
		"SCS": true, // Selected Cities Summary
		"RER": true, // Record Report
	}
	return obsCodes[code]
}

func isDiscussionProductCode(code string) bool {
	discussionCodes := map[string]bool{
		"AFD": true, // Area Forecast Discussion
		"FWD": true, // Fire Weather Outlook Discussion
		"PMD": true, // Prognostic Meteorological Discussion
		"HMD": true, // Hydrometeorological Discussion
		"TWD": true, // Tropical Weather Discussion
		"SCD": true, // Supplementary Climatological Data
		"PWO": true, // Public Severe Weather Outlook
	}
	return discussionCodes[code]
}

func isMarineProductCode(code string) bool {
	marineCodes := map[string]bool{
		"MWW": true, // Marine Weather Warning
		"MWS": true, // Marine Weather Statement
		"SMW": true, // Special Marine Warning
		"CWF": true, // Coastal Waters Forecast
		"OFF": true, // Offshore Forecast
		"NSH": true, // Nearshore Marine Forecast
		"GLF": true, // Great Lakes Forecast
		"HSF": true, // High Seas Forecast
		"MFM": true, // Marine Forecast Matrix
		"MIM": true, // Marine Interpretation Message
		"MVF": true, // Marine Verification Coded Message
		"CGR": true, // Coast Guard Surface Report
		"SHP": true, // Surface Ship Report
		"PLS": true, // Plain Language Ship Report
		"TID": true, // Tide Report
		"ICE": true, // Ice Forecast
		"IOB": true, // Ice Observation
		"SRF": true, // Surf Forecast
		"SRD": true, // Surf Discussion
	}
	return marineCodes[code]
}

func isAviationProductCode(code string) bool {
	aviationCodes := map[string]bool{
		"TAF": true, // Terminal Aerodrome Forecast
		"MTR": true, // METAR Formatted Surface Weather Observation
		"AWW": true, // Airport Weather Warning
		"SIG": true, // International Sigmet
		"WST": true, // Tropical Cyclone Sigmet
		"WSV": true, // Volcanic Activity Sigmet
		"TAP": true, // Terminal Alerting Products
		"OFA": true, // Offshore Aviation Area Forecast
		"CWA": true, // Center Weather Advisory
		"CWS": true, // Center Weather Statement
	}
	// Also check for Airmet codes
	if len(code) == 3 && strings.HasPrefix(code, "WA") {
		return true
	}
	// Check for FA codes (Aviation Area Forecasts)
	if len(code) == 3 && strings.HasPrefix(code, "FA") {
		return true
	}
	return aviationCodes[code]
}

func filterEMWINTextByStation(files []EMWINTextFile, station string) []EMWINTextFile {
	var filtered []EMWINTextFile
	
	for _, file := range files {
		if strings.EqualFold(file.Station, station) {
			filtered = append(filtered, file)
		}
	}
	
	return filtered
}

func filterEMWINTextByOffice(files []EMWINTextFile, office string) []EMWINTextFile {
	var filtered []EMWINTextFile
	officeUpper := strings.ToUpper(office)
	
	for _, file := range files {
		// Check if the station code matches the office
		// Office can be a 3-letter CWA code (e.g., "BMX") or 4-letter station ID (e.g., "KBMX")
		stationUpper := strings.ToUpper(file.Station)
		
		// Match if:
		// - Station exactly matches office (e.g., "KBMX" == "KBMX")
		// - Station contains office code (e.g., "KBMX" contains "BMX")
		// - Office is contained in station (e.g., "BMX" is in "KBMX")
		if stationUpper == officeUpper ||
		   strings.Contains(stationUpper, officeUpper) ||
		   strings.Contains(officeUpper, stationUpper) {
			filtered = append(filtered, file)
		}
	}
	
	return filtered
}

// Helper functions to categorize products
func isWarningProduct(productCode string) bool {
	// Check WMO codes for warnings
	warningCodes := []string{"WOUS", "WWUS", "WFUS"}
	for _, code := range warningCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isForecastProduct(productCode string) bool {
	// Check WMO codes for forecasts
	forecastCodes := []string{"FPUS", "FZUS", "FOUS"}
	for _, code := range forecastCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isObservationProduct(productCode string) bool {
	// Check WMO codes for observations
	observationCodes := []string{"SAUS", "SPUS", "SACA", "ASUS"}
	for _, code := range observationCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isDiscussionProduct(productCode string) bool {
	// Check WMO codes for discussions
	discussionCodes := []string{"FXUS"}
	for _, code := range discussionCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isMarineProduct(productCode string) bool {
	// Check WMO codes for marine products
	marineCodes := []string{"FZUS", "FZAK", "FZHW"}
	for _, code := range marineCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isAviationProduct(productCode string) bool {
	// Check WMO codes for aviation products
	aviationCodes := []string{"FTUS", "FAUS", "FVUS"}
	for _, code := range aviationCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isNumeric(s string) bool {
	for _, char := range s {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}