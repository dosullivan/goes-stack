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

// WeatherProduct represents a weather data product category
type WeatherProduct struct {
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Category    string `json:"category"`
}

// ProductImage represents an individual image in a product
type ProductImage struct {
	URL       string    `json:"url"`
	Timestamp time.Time `json:"timestamp"`
	Filename  string    `json:"filename"`
}

// ProductPath contains the path within the bucket for a product
type ProductPath struct {
	Path string
}

// GetWeatherProducts returns all available weather product categories
func GetWeatherProducts(s3Client *s3.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		products := []WeatherProduct{
			// ABI Full Disk Products
			{Key: "fd_color", Title: "GOES-19 Full Disk Color", Description: "False color composite imagery", Icon: "globe-americas", Category: "abi"},
			{Key: "fd_ch02", Title: "Channel 2 (Red)", Description: "Visible red channel", Category: "abi"},
			{Key: "fd_ch07", Title: "Channel 7 (Shortwave IR)", Description: "Shortwave infrared", Category: "abi"},
			{Key: "fd_ch07_enhanced", Title: "Channel 7 Enhanced", Description: "Enhanced shortwave infrared", Category: "abi"},
			{Key: "fd_ch08", Title: "Channel 8 (Upper Troposphere)", Description: "Upper tropospheric water vapor", Category: "abi"},
			{Key: "fd_ch08_enhanced", Title: "Channel 8 Enhanced", Description: "Enhanced upper troposphere", Category: "abi"},
			{Key: "fd_ch09", Title: "Channel 9 (Mid Troposphere)", Description: "Mid-level water vapor", Category: "abi"},
			{Key: "fd_ch09_enhanced", Title: "Channel 9 Enhanced", Description: "Enhanced mid troposphere", Category: "abi"},
			{Key: "fd_ch13", Title: "Channel 13 (Clean Longwave IR)", Description: "Clean longwave infrared", Category: "abi"},
			{Key: "fd_ch13_enhanced", Title: "Channel 13 Enhanced", Description: "Enhanced clean longwave IR", Category: "abi"},
			{Key: "fd_ch14", Title: "Channel 14 (Longwave IR)", Description: "Standard longwave infrared", Category: "abi"},
			{Key: "fd_ch14_enhanced", Title: "Channel 14 Enhanced", Description: "Enhanced longwave IR", Category: "abi"},
			{Key: "fd_ch15", Title: "Channel 15 (Dirty Longwave IR)", Description: "Dirty longwave infrared", Category: "abi"},
			{Key: "fd_ch15_enhanced", Title: "Channel 15 Enhanced", Description: "Enhanced dirty longwave IR", Category: "abi"},

			// EMWIN Graphics Products
			{Key: "radar_northeast", Title: "US Northeast Radar", Description: "Regional radar composite", Icon: "cloud-rain", Category: "emwin"},
			{Key: "radar_southeast", Title: "US Southeast Radar", Description: "Regional radar composite", Category: "emwin"},
			{Key: "radar_greatlakes", Title: "US Great Lakes Radar", Description: "Regional radar composite", Category: "emwin"},
			{Key: "radar_southernplains", Title: "US Southern Plains Radar", Description: "Regional radar composite", Category: "emwin"},
			{Key: "radar_northrockies", Title: "US Northern Rockies Radar", Description: "Regional radar composite", Category: "emwin"},
			{Key: "radar_southrockies", Title: "US Southern Rockies Radar", Description: "Regional radar composite", Category: "emwin"},
			{Key: "radar_uppermiss", Title: "US Upper Mississippi Valley Radar", Description: "Regional radar composite", Category: "emwin"},
			{Key: "radar_lowermiss", Title: "US Lower Mississippi Valley Radar", Description: "Regional radar composite", Category: "emwin"},
			{Key: "radar_pacnw", Title: "US Pacific Northwest Radar", Description: "Regional radar composite", Category: "emwin"},
			{Key: "radar_pacsw", Title: "US Pacific Southwest Radar", Description: "Regional radar composite", Category: "emwin"},
			{Key: "radar_alaska", Title: "Alaska Radar", Description: "Alaska radar composite", Category: "emwin"},
			{Key: "radar_hawaii", Title: "Hawaii Radar", Description: "Hawaii radar composite", Category: "emwin"},
			{Key: "radar_guam", Title: "Guam Radar", Description: "Guam radar composite", Category: "emwin"},
			{Key: "radar_pr", Title: "Puerto Rico Radar", Description: "Puerto Rico radar composite", Category: "emwin"},
			{Key: "radar_us_composite", Title: "US Composite Radar", Description: "National radar composite", Category: "emwin"},
			{Key: "sat_meteosat", Title: "METEOSAT Infrared", Description: "European satellite infrared", Category: "emwin"},
			{Key: "sat_himawari", Title: "Himawari Infrared", Description: "Japanese satellite infrared", Category: "emwin"},
			{Key: "sat_goes19_us", Title: "GOES-19 Enhanced US", Description: "GOES-19 Channel 13 Enhanced", Category: "emwin"},
			{Key: "sat_goes19_hurricane", Title: "GOES-19 Hurricane Basin", Description: "Atlantic hurricane region", Category: "emwin"},
			{Key: "sat_goes19_pr", Title: "GOES-19 Puerto Rico", Description: "Puerto Rico region", Category: "emwin"},
			{Key: "sat_goeswest_fd", Title: "GOES West Full Disk", Description: "GOES West full disk enhanced", Category: "emwin"},
			{Key: "sat_goeswest_meso", Title: "GOES West Coast", Description: "GOES West coastal region", Category: "emwin"},
			{Key: "analysis_np_surface", Title: "North Pacific Surface Analysis", Description: "Surface pressure analysis", Category: "emwin"},
			{Key: "analysis_np_ice", Title: "North Pacific Sea Ice", Description: "Sea ice analysis", Category: "emwin"},
			{Key: "analysis_caribbean", Title: "Caribbean Surface Analysis", Description: "Caribbean surface analysis", Category: "emwin"},
			{Key: "outlook_convective_day1", Title: "Convective Outlook Day 1", Description: "Severe weather outlook", Category: "emwin"},
			{Key: "outlook_convective_day2", Title: "Convective Outlook Day 2", Description: "Severe weather outlook", Category: "emwin"},
			{Key: "alerts_watches_warnings", Title: "Watches, Warnings, Advisories", Description: "Active weather alerts", Category: "emwin"},
			{Key: "fronts_map", Title: "Fronts / Weather Type", Description: "Weather fronts and precipitation type", Category: "emwin"},
			{Key: "flood_outlook", Title: "River Flood Outlook", Description: "Significant river flooding outlook", Category: "emwin"},
			{Key: "qpf_6hour_color", Title: "6-Hour QPF (Color)", Description: "Quantitative precipitation forecast", Category: "emwin"},
			{Key: "qpf_6hour", Title: "6-Hour QPF", Description: "Quantitative precipitation forecast", Category: "emwin"},
			{Key: "qpf_24hour_day1", Title: "24-Hour QPF Day 1", Description: "Daily precipitation forecast", Category: "emwin"},
			{Key: "qpf_24hour_day2", Title: "24-Hour QPF Day 2", Description: "Daily precipitation forecast", Category: "emwin"},
			{Key: "rainfall_excessive", Title: "Excessive Rainfall Outlook", Description: "Heavy rainfall potential", Category: "emwin"},
			{Key: "analysis_na_surface", Title: "North America Surface Analysis", Description: "Continental surface analysis", Category: "emwin"},
		}

		c.JSON(http.StatusOK, gin.H{"products": products})
	}
}

// GetProductImages returns images for a specific weather product
func GetProductImages(s3Client *s3.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		productKey := c.Param("product")
		dateStr := c.Query("date")
		
		if productKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product key is required"})
			return
		}

		// Map product keys to bucket paths
		productPath := getProductPath(productKey)
		if productPath == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown product key"})
			return
		}

		ctx := context.Background()
		cstLocation, err := time.LoadLocation("America/Chicago")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load timezone"})
			return
		}

		var images []ProductImage

		// Use the single bucket for all data
		bucketName := s3Client.BucketName
		baseURL := s3Client.BaseURL

		// If date specified, get images for that date, otherwise get recent images
		if dateStr != "" {
			images, err = getProductImagesForDate(ctx, s3Client.Client, bucketName, baseURL, productPath, dateStr, cstLocation)
		} else {
			// Get images from the last 7 days
			images, err = getRecentProductImages(ctx, s3Client.Client, bucketName, baseURL, productPath, cstLocation, 7)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"product": productKey,
			"images":  images,
			"count":   len(images),
		})
	}
}

// getProductPath maps product keys to their paths within the bucket
func getProductPath(productKey string) string {
	// Map product keys to their paths in the structured bucket
	productPaths := map[string]string{
		// GOES-19 ABI products
		"fd_color":          "goes19/fd/fc/",
		"fd_ch02":           "goes19/fd/ch02/",
		"fd_ch07":           "goes19/fd/ch07/", 
		"fd_ch07_enhanced":  "goes19/fd/ch07_enhanced/",
		"fd_ch08":           "goes19/fd/ch08/",
		"fd_ch08_enhanced":  "goes19/fd/ch08_enhanced/",
		"fd_ch09":           "goes19/fd/ch09/",
		"fd_ch09_enhanced":  "goes19/fd/ch09_enhanced/",
		"fd_ch13":           "goes19/fd/ch13/",
		"fd_ch13_enhanced":  "goes19/fd/ch13_enhanced/",
		"fd_ch14":           "goes19/fd/ch14/",
		"fd_ch14_enhanced":  "goes19/fd/ch14_enhanced/",
		"fd_ch15":           "goes19/fd/ch15/",
		"fd_ch15_enhanced":  "goes19/fd/ch15_enhanced/",
		
		// EMWIN products - using product code patterns
		"radar_northeast":          "emwin/",  // Will filter by RADNTHES
		"radar_southeast":          "emwin/",  // Will filter by RADSTHES
		"radar_greatlakes":         "emwin/",  // Will filter by RADGRTLK
		"radar_southernplains":     "emwin/",  // Will filter by RADSTHPL
		"radar_northrockies":       "emwin/",  // Will filter by RADRCKNT
		"radar_southrockies":       "emwin/",  // Will filter by RADRCKST
		"radar_uppermiss":          "emwin/",  // Will filter by RADUMSVY
		"radar_lowermiss":          "emwin/",  // Will filter by RADSMSVY
		"radar_pacnw":              "emwin/",  // Will filter by RADPACNW
		"radar_pacsw":              "emwin/",  // Will filter by RADPACSW
		"radar_alaska":             "emwin/",  // Will filter by RADALLAK
		"radar_hawaii":             "emwin/",  // Will filter by RADALLHI
		"radar_guam":               "emwin/",  // Will filter by RADALLGU
		"radar_pr":                 "emwin/",  // Will filter by RADALLPR
		"radar_us_composite":       "emwin/",  // Will filter by RADREFUS
		"sat_meteosat":            "emwin/",  // Will filter by INDCIRUS
		"sat_himawari":            "emwin/",  // Will filter by GMS008JA
		"sat_goes19_us":           "emwin/",  // Will filter by G16CIRUS
		"sat_goes19_hurricane":     "emwin/",  // Will filter by G02HURUS
		"sat_goes19_pr":           "emwin/",  // Will filter by IMGSJUPR
		"sat_goeswest_fd":         "emwin/",  // Will filter by G10FDIUS
		"sat_goeswest_meso":       "emwin/",  // Will filter by G10CIRUS
		"analysis_np_surface":      "emwin/",  // Will filter by NPSA01US
		"analysis_np_ice":          "emwin/",  // Will filter by NPIC01US
		"analysis_caribbean":       "emwin/",  // Will filter by CSA001US
		"outlook_convective_day1":  "emwin/",  // Will filter by MODDY1US
		"outlook_convective_day2":  "emwin/",  // Will filter by MODDY2US
		"alerts_watches_warnings":  "emwin/",  // Will filter by IMGWWAUS
		"fronts_map":              "emwin/",  // Will filter by MOD96FBW
		"flood_outlook":           "emwin/",  // Will filter by GPHJ88US
		"qpf_6hour_color":         "emwin/",  // Will filter by MOD91EUS
		"qpf_6hour":               "emwin/",  // Will filter by MOD93SUS
		"qpf_24hour_day1":         "emwin/",  // Will filter by MODQP1US
		"qpf_24hour_day2":         "emwin/",  // Will filter by MODQP2US
		"rainfall_excessive":       "emwin/",  // Will filter by MOD94SUS
		"analysis_na_surface":      "emwin/",  // Will filter by IMGFNT
	}

	// Return the path for the product key
	if path, exists := productPaths[productKey]; exists {
		return path
	}

	return ""
}

func getProductImagesForDate(ctx context.Context, client *minio.Client, bucketName, baseURL, bucketPath, dateStr string, cstLocation *time.Location) ([]ProductImage, error) {
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

	var images []ProductImage

	for _, utcDate := range utcDates {
		prefix := bucketPath + utcDate + "/"

		objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		})

		for object := range objectCh {
			if object.Err != nil {
				return nil, object.Err
			}

			// Skip non-image files
			if !strings.HasSuffix(strings.ToLower(object.Key), ".png") &&
				!strings.HasSuffix(strings.ToLower(object.Key), ".jpg") &&
				!strings.HasSuffix(strings.ToLower(object.Key), ".gif") {
				continue
			}

			// Extract UTC timestamp from the object key
			utcTime, err := extractTimestamp(object.Key)
			if err != nil {
				continue
			}

			// Convert to CST and check if it falls on our requested date
			cstTime := utcTime.In(cstLocation)
			if cstTime.Format("2006-01-02") == dateStr {
				images = append(images, ProductImage{
					URL:       baseURL + object.Key,
					Timestamp: cstTime,
					Filename:  object.Key[strings.LastIndex(object.Key, "/")+1:],
				})
			}
		}
	}

	// Sort by timestamp
	sort.Slice(images, func(i, j int) bool {
		return images[i].Timestamp.Before(images[j].Timestamp)
	})

	return images, nil
}

func getRecentProductImages(ctx context.Context, client *minio.Client, bucketName, baseURL, bucketPath string, cstLocation *time.Location, maxDaysBack int) ([]ProductImage, error) {
	var images []ProductImage
	nowUTC := time.Now().UTC()

	for i := -1; i < maxDaysBack; i++ { // Start at -1 to include tomorrow
		checkDate := nowUTC.AddDate(0, 0, -i).Format("2006-01-02")
		prefix := bucketPath + checkDate + "/"

		objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		})

		for object := range objectCh {
			if object.Err != nil {
				return nil, object.Err
			}

			// Skip non-image files
			if !strings.HasSuffix(strings.ToLower(object.Key), ".png") &&
				!strings.HasSuffix(strings.ToLower(object.Key), ".jpg") &&
				!strings.HasSuffix(strings.ToLower(object.Key), ".gif") {
				continue
			}

			utcTime, err := extractTimestamp(object.Key)
			if err != nil {
				continue
			}

			cstTime := utcTime.In(cstLocation)
			images = append(images, ProductImage{
				URL:       baseURL + object.Key,
				Timestamp: cstTime,
				Filename:  object.Key[strings.LastIndex(object.Key, "/")+1:],
			})
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(images, func(i, j int) bool {
		return images[i].Timestamp.After(images[j].Timestamp)
	})

	return images, nil
}