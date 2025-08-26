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

// GetEMWINTextFiles returns EMWIN text files, optionally filtered by category, date, and station
func GetEMWINTextFiles(s3Client *s3.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		category := c.Query("category")
		dateStr := c.Query("date")
		station := c.Query("station")

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

		// Apply filters
		if category != "" {
			files = filterEMWINTextByCategory(files, category)
		}

		if station != "" {
			files = filterEMWINTextByStation(files, station)
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

	for _, utcDate := range utcDates {
		prefix := emwinPath + utcDate + "/"

		objectCh := s3Client.Client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		})

		for object := range objectCh {
			if object.Err != nil {
				return nil, object.Err
			}

			// Only process .TXT files
			if !strings.HasSuffix(strings.ToUpper(object.Key), ".TXT") {
				continue
			}

			// Try to parse timestamp from filename (EMWIN files have timestamps in names)
			timestamp := object.LastModified
			if fileTimestamp := parseEMWINTimestamp(object.Key); !fileTimestamp.IsZero() {
				timestamp = fileTimestamp
			}

			cstTime := timestamp.In(cstLocation)
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

	for i := 0; i < maxDaysBack; i++ {
		checkDate := nowCST.AddDate(0, 0, -i).Format("2006-01-02")
		prefix := emwinPath + checkDate + "/"

		objectCh := s3Client.Client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		})

		for object := range objectCh {
			if object.Err != nil {
				return nil, object.Err
			}

			// Only process .TXT files
			if !strings.HasSuffix(strings.ToUpper(object.Key), ".TXT") {
				continue
			}

			timestamp := object.LastModified
			if fileTimestamp := parseEMWINTimestamp(object.Key); !fileTimestamp.IsZero() {
				timestamp = fileTimestamp
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

// extractStation extracts the station identifier from EMWIN filename
func extractStation(filename string) string {
	// Example: A_ASUS41KGYX250310_C_KWIN_20250825031112_921174-3-RWRGYXME.TXT
	// Extract KGYX part (station identifier)
	parts := strings.Split(filename, "_")
	if len(parts) > 1 {
		productPart := parts[1]
		// Station codes are typically 4 characters starting with K
		if len(productPart) >= 10 {
			station := productPart[6:10]
			if strings.HasPrefix(station, "K") {
				return station
			}
		}
	}
	return ""
}

// getProductDescription provides human-readable descriptions for product codes
func getProductDescription(filename string) string {
	productCode := extractProductCode(filename)
	
	descriptions := map[string]string{
		"ASUS41": "Area Forecast Discussion",
		"ASUS42": "Zone Forecast Product",
		"ASUS43": "Area Forecast Discussion",
		"ASUS44": "Zone Forecast Product", 
		"ASUS45": "Zone Forecast Product",
		"ASUS46": "Zone Forecast Product",
		"FOUS52": "Area Forecast Matrices",
		"FOUS53": "Area Forecast Matrices",
		"FOUS54": "Area Forecast Matrices",
		"FOUS55": "Area Forecast Matrices",
		"FOUS56": "Area Forecast Matrices",
		"FPUS52": "Zone Forecast Product",
		"FPUS53": "Zone Forecast Product",
		"FPUS54": "Zone Forecast Product",
		"FPUS55": "Zone Forecast Product",
		"FPUS56": "Zone Forecast Product",
		"SAUS70": "Surface Hourly Observation",
		"SAUS80": "Surface Hourly Observation",
		"SPUS70": "Special Surface Observation",
		"SPUS80": "Special Surface Observation",
		"FTUS80": "Terminal Aerodrome Forecast",
		"NWUS51": "Local Storm Report",
		"NWUS55": "Local Storm Report",
		"FZUS51": "Zone Forecast Product",
		"FZUS52": "Coastal Waters Forecast",
		"FZUS53": "Zone Forecast Product",
		"FZUS56": "Coastal Waters Forecast",
	}

	if desc, exists := descriptions[productCode]; exists {
		return desc
	}
	return "Weather Product"
}

func filterEMWINTextByCategory(files []EMWINTextFile, category string) []EMWINTextFile {
	var filtered []EMWINTextFile
	
	for _, file := range files {
		switch category {
		case "weather_warnings":
			if isWarningProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		case "forecasts":
			if isForecastProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		case "observations":
			if isObservationProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		case "discussions":
			if isDiscussionProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		case "marine":
			if isMarineProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		case "aviation":
			if isAviationProduct(file.ProductCode) {
				filtered = append(filtered, file)
			}
		default:
			filtered = append(filtered, file)
		}
	}
	
	return filtered
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

// Helper functions to categorize products
func isWarningProduct(productCode string) bool {
	warningCodes := []string{"WOUS", "WWUS", "WFUS"}
	for _, code := range warningCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isForecastProduct(productCode string) bool {
	forecastCodes := []string{"FPUS", "FZUS", "FOUS"}
	for _, code := range forecastCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isObservationProduct(productCode string) bool {
	observationCodes := []string{"SAUS", "SPUS", "SACA"}
	for _, code := range observationCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isDiscussionProduct(productCode string) bool {
	discussionCodes := []string{"ASUS", "FXUS"}
	for _, code := range discussionCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isMarineProduct(productCode string) bool {
	marineCodes := []string{"FZUS", "FZAK", "FZHW"}
	for _, code := range marineCodes {
		if strings.HasPrefix(productCode, code) {
			return true
		}
	}
	return false
}

func isAviationProduct(productCode string) bool {
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