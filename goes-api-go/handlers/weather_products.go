package handlers

import (
	"context"
	"log"
	"net/http"
	"net/url"
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
			// GOES-18 Full Disk Products
			{Key: "goes18_fd_ch13", Title: "GOES-18 Channel 13", Description: "Clean longwave infrared", Category: "goes18"},
			{Key: "goes18_fd_ch13_enhanced", Title: "GOES-18 Channel 13 Enhanced", Description: "Enhanced clean longwave IR", Category: "goes18"},
			
			// GOES-19 Full Disk Products
			{Key: "fd_color", Title: "GOES-19 Full Disk Color", Description: "False color composite imagery", Icon: "globe-americas", Category: "goes19_fd"},
			{Key: "fd_ch02", Title: "Channel 2 (Red)", Description: "Visible red channel", Category: "goes19_fd"},
			{Key: "fd_ch07", Title: "Channel 7 (Shortwave IR)", Description: "Shortwave infrared", Category: "goes19_fd"},
			{Key: "fd_ch07_enhanced", Title: "Channel 7 Enhanced", Description: "Enhanced shortwave infrared", Category: "goes19_fd"},
			{Key: "fd_ch08", Title: "Channel 8 (Upper Troposphere)", Description: "Upper tropospheric water vapor", Category: "goes19_fd"},
			{Key: "fd_ch08_enhanced", Title: "Channel 8 Enhanced", Description: "Enhanced upper troposphere", Category: "goes19_fd"},
			{Key: "fd_ch09", Title: "Channel 9 (Mid Troposphere)", Description: "Mid-level water vapor", Category: "goes19_fd"},
			{Key: "fd_ch09_enhanced", Title: "Channel 9 Enhanced", Description: "Enhanced mid troposphere", Category: "goes19_fd"},
			{Key: "fd_ch13", Title: "Channel 13 (Clean Longwave IR)", Description: "Clean longwave infrared", Category: "goes19_fd"},
			{Key: "fd_ch13_enhanced", Title: "Channel 13 Enhanced", Description: "Enhanced clean longwave IR", Category: "goes19_fd"},
			{Key: "fd_ch14", Title: "Channel 14 (Longwave IR)", Description: "Standard longwave infrared", Category: "goes19_fd"},
			{Key: "fd_ch14_enhanced", Title: "Channel 14 Enhanced", Description: "Enhanced longwave IR", Category: "goes19_fd"},
			{Key: "fd_ch15", Title: "Channel 15 (Dirty Longwave IR)", Description: "Dirty longwave infrared", Category: "goes19_fd"},
			{Key: "fd_ch15_enhanced", Title: "Channel 15 Enhanced", Description: "Enhanced dirty longwave IR", Category: "goes19_fd"},

			// GOES-19 Mesoscale 1 Products
			{Key: "m1_color", Title: "Mesoscale 1 Color", Description: "False color composite", Category: "goes19_m1"},
			{Key: "m1_ch02", Title: "Mesoscale 1 Channel 2", Description: "Visible red channel", Category: "goes19_m1"},
			{Key: "m1_ch07", Title: "Mesoscale 1 Channel 7", Description: "Shortwave infrared", Category: "goes19_m1"},
			{Key: "m1_ch07_enhanced", Title: "Mesoscale 1 Ch7 Enhanced", Description: "Enhanced shortwave IR", Category: "goes19_m1"},
			{Key: "m1_ch13", Title: "Mesoscale 1 Channel 13", Description: "Clean longwave IR", Category: "goes19_m1"},
			{Key: "m1_ch13_enhanced", Title: "Mesoscale 1 Ch13 Enhanced", Description: "Enhanced clean IR", Category: "goes19_m1"},

			// GOES-19 Mesoscale 2 Products
			{Key: "m2_color", Title: "Mesoscale 2 Color", Description: "False color composite", Category: "goes19_m2"},
			{Key: "m2_ch02", Title: "Mesoscale 2 Channel 2", Description: "Visible red channel", Category: "goes19_m2"},
			{Key: "m2_ch07", Title: "Mesoscale 2 Channel 7", Description: "Shortwave infrared", Category: "goes19_m2"},
			{Key: "m2_ch07_enhanced", Title: "Mesoscale 2 Ch7 Enhanced", Description: "Enhanced shortwave IR", Category: "goes19_m2"},
			{Key: "m2_ch13", Title: "Mesoscale 2 Channel 13", Description: "Clean longwave IR", Category: "goes19_m2"},
			{Key: "m2_ch13_enhanced", Title: "Mesoscale 2 Ch13 Enhanced", Description: "Enhanced clean IR", Category: "goes19_m2"},

			// GOES-19 Non-CMIP Products
			{Key: "non_cmip_acha", Title: "Cloud Top Height", Description: "ABI Cloud Height Algorithm", Category: "goes19_derived"},
			{Key: "non_cmip_acht", Title: "Cloud Top Temperature", Description: "Cloud top temperature product", Category: "goes19_derived"},
			{Key: "non_cmip_dsi", Title: "Derived Stability Index", Description: "Atmospheric stability", Category: "goes19_derived"},
			{Key: "non_cmip_rrqpe", Title: "Rainfall Rate QPE", Description: "Quantitative precipitation estimate", Category: "goes19_derived"},
			{Key: "non_cmip_tpw", Title: "Total Precipitable Water", Description: "Atmospheric water vapor", Category: "goes19_derived"},

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
			{Key: "analysis_na_surface", Title: "North America Surface Analysis", Description: "Continental surface analysis", Category: "emwin"},
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
		}

		c.JSON(http.StatusOK, gin.H{"products": products})
	}
}

// GetProductImages returns images for a specific weather product (UTC version)
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
		var images []ProductImage

		// Use the single bucket for all data
		bucketName := s3Client.BucketName
		baseURL := s3Client.BaseURL

		// Get the filter pattern for EMWIN products
		filterPattern := getEMWINPattern(productKey)
		
		// If date specified, get images for that date, otherwise get recent images
		if dateStr != "" {
			images = getProductImagesForDateUTC(ctx, s3Client.Client, bucketName, baseURL, productPath, dateStr, filterPattern)
		} else {
			// Get images from the last 7 days
			images = getRecentProductImagesUTC(ctx, s3Client.Client, bucketName, baseURL, productPath, 7, filterPattern)
		}

		c.JSON(http.StatusOK, gin.H{
			"product": productKey,
			"images":  images,
			"count":   len(images),
		})
	}
}

func getProductImagesForDateUTC(ctx context.Context, client *minio.Client, bucketName, baseURL, bucketPath, dateStr, filterPattern string) []ProductImage {
	var images []ProductImage
	
	// Use the date directly as UTC
	prefix := bucketPath + dateStr + "/"
	log.Printf("DEBUG: Looking for images with prefix: %s in bucket: %s\n", prefix, bucketName)

	objectCh := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			log.Printf("DEBUG: Error listing object: %v\n", object.Err)
			continue
		}

		// Skip non-image files
		if !strings.HasSuffix(strings.ToLower(object.Key), ".png") &&
			!strings.HasSuffix(strings.ToLower(object.Key), ".jpg") &&
			!strings.HasSuffix(strings.ToLower(object.Key), ".gif") {
			continue
		}
		
		// Apply filter pattern if specified (for EMWIN products)
		if filterPattern != "" && !strings.Contains(object.Key, filterPattern) {
			continue
		}

		utcTime, err := extractTimestamp(object.Key)
		if err != nil {
			log.Printf("DEBUG: Could not extract timestamp from %s: %v\n", object.Key, err)
			continue
		}

		// Use proxied URL instead of direct MinIO URL
		proxiedURL := "/proxy/image?url=" + url.QueryEscape(baseURL + object.Key)
		images = append(images, ProductImage{
			URL:       proxiedURL,
			Timestamp: utcTime, // Keep in UTC
			Filename:  object.Key[strings.LastIndex(object.Key, "/")+1:],
		})
	}

	// Sort by timestamp (newest first)
	sort.Slice(images, func(i, j int) bool {
		return images[i].Timestamp.After(images[j].Timestamp)
	})

	log.Printf("DEBUG: Found %d images for date %s\n", len(images), dateStr)
	return images
}

func getRecentProductImagesUTC(ctx context.Context, client *minio.Client, bucketName, baseURL, bucketPath string, maxDaysBack int, filterPattern string) []ProductImage {
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
				continue
			}

			// Skip non-image files
			if !strings.HasSuffix(strings.ToLower(object.Key), ".png") &&
				!strings.HasSuffix(strings.ToLower(object.Key), ".jpg") &&
				!strings.HasSuffix(strings.ToLower(object.Key), ".gif") {
				continue
			}
			
			// Apply filter pattern if specified (for EMWIN products)
			if filterPattern != "" && !strings.Contains(object.Key, filterPattern) {
				continue
			}

			utcTime, err := extractTimestamp(object.Key)
			if err != nil {
				continue
			}

			// Use proxied URL instead of direct MinIO URL
			proxiedURL := "/proxy/image?url=" + url.QueryEscape(baseURL + object.Key)
			images = append(images, ProductImage{
				URL:       proxiedURL,
				Timestamp: utcTime, // Keep in UTC
				Filename:  object.Key[strings.LastIndex(object.Key, "/")+1:],
			})
		}
	}

	// Sort by timestamp (newest first)
	sort.Slice(images, func(i, j int) bool {
		return images[i].Timestamp.After(images[j].Timestamp)
	})

	return images
}

func getProductPath(productKey string) string {
	// Map product keys to their paths in the structured bucket
	productPaths := map[string]string{
		// GOES-18 ABI products
		"goes18_fd_ch13":          "goes18/fd/ch13/",
		"goes18_fd_ch13_enhanced":  "goes18/fd/ch13_enhanced/",
		
		// GOES-19 Full Disk ABI products
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
		
		// GOES-19 Mesoscale 1 products
		"m1_color":          "goes19/m1/fc/",
		"m1_ch02":           "goes19/m1/ch02/",
		"m1_ch07":           "goes19/m1/ch07/",
		"m1_ch07_enhanced":  "goes19/m1/ch07_enhanced/",
		"m1_ch13":           "goes19/m1/ch13/",
		"m1_ch13_enhanced":  "goes19/m1/ch13_enhanced/",
		
		// GOES-19 Mesoscale 2 products
		"m2_color":          "goes19/m2/fc/",
		"m2_ch02":           "goes19/m2/ch02/",
		"m2_ch07":           "goes19/m2/ch07/",
		"m2_ch07_enhanced":  "goes19/m2/ch07_enhanced/",
		"m2_ch13":           "goes19/m2/ch13/",
		"m2_ch13_enhanced":  "goes19/m2/ch13_enhanced/",
		
		// GOES-19 Non-CMIP products
		"non_cmip_acha":     "goes19/non-cmip/fd/acha/",
		"non_cmip_acht":     "goes19/non-cmip/fd/acht/",
		"non_cmip_dsi":      "goes19/non-cmip/fd/dsi/",
		"non_cmip_rrqpe":    "goes19/non-cmip/fd/rrqpe/",
		"non_cmip_tpw":      "goes19/non-cmip/fd/tpw/",
		
		// EMWIN products
		"radar_northeast":          "emwin/",
		"radar_southeast":          "emwin/",
		"radar_greatlakes":         "emwin/",
		"radar_southernplains":     "emwin/",
		"radar_northrockies":       "emwin/",
		"radar_southrockies":       "emwin/",
		"radar_uppermiss":          "emwin/",
		"radar_lowermiss":          "emwin/",
		"radar_pacnw":              "emwin/",
		"radar_pacsw":              "emwin/",
		"radar_alaska":             "emwin/",
		"radar_hawaii":             "emwin/",
		"radar_guam":               "emwin/",
		"radar_pr":                 "emwin/",
		"radar_us_composite":       "emwin/",
		"sat_meteosat":            "emwin/",
		"sat_himawari":            "emwin/",
		"sat_goes19_us":           "emwin/",
		"sat_goes19_hurricane":     "emwin/",
		"sat_goes19_pr":           "emwin/",
		"sat_goeswest_fd":         "emwin/",
		"sat_goeswest_meso":       "emwin/",
		"analysis_np_surface":      "emwin/",
		"analysis_np_ice":          "emwin/",
		"analysis_na_surface":      "emwin/",
		"analysis_caribbean":       "emwin/",
		"outlook_convective_day1":  "emwin/",
		"outlook_convective_day2":  "emwin/",
		"alerts_watches_warnings":  "emwin/",
		"fronts_map":              "emwin/",
		"flood_outlook":           "emwin/",
		"qpf_6hour":               "emwin/",
		"qpf_24hour_day1":         "emwin/",
		"qpf_24hour_day2":         "emwin/",
		"rainfall_excessive":      "emwin/",
	}
	
	return productPaths[productKey]
}

func getEMWINPattern(productKey string) string {
	// Map EMWIN product keys to their filename patterns
	emwinPatterns := map[string]string{
		"radar_northeast":          "RADNTHES",
		"radar_southeast":          "RADSTHES",
		"radar_greatlakes":         "RADGRTLK",
		"radar_southernplains":     "RADSTHPL",
		"radar_northrockies":       "RADRCKNT",
		"radar_southrockies":       "RADRCKST",
		"radar_uppermiss":          "RADUMSVY",
		"radar_lowermiss":          "RADSMSVY",
		"radar_pacnw":              "RADPACNW",
		"radar_pacsw":              "RADPACSW",
		"radar_alaska":             "RADALLAK",
		"radar_hawaii":             "RADALLHI",
		"radar_guam":               "RADALLGU",
		"radar_pr":                 "RADALLPR",
		"radar_us_composite":       "RADREFUS",
		"sat_meteosat":            "INDCIRUS",
		"sat_himawari":            "GMS008JA",
		"sat_goes19_us":           "G16CIRUS",
		"sat_goes19_hurricane":     "G02HURUS",
		"sat_goes19_pr":           "IMGSJUPR",
		"sat_goeswest_fd":         "G10FDIUS",
		"sat_goeswest_meso":       "G10CIRUS",
		"analysis_np_surface":      "NPSA01US",
		"analysis_np_ice":          "NPIC01US",
		"analysis_na_surface":      "IMGFNT12",
		"analysis_caribbean":       "CSA001US",
		"outlook_convective_day1":  "MODDY1US",
		"outlook_convective_day2":  "MODDY2US",
		"alerts_watches_warnings":  "IMGWWAUS",
		"fronts_map":              "MOD96FBW",
		"flood_outlook":           "GPHJ88US",
		"qpf_6hour":               "MOD93SUS",
		"qpf_24hour_day1":         "MODQP1US",
		"qpf_24hour_day2":         "MODQP2US",
		"rainfall_excessive":       "MOD94SUS",
	}
	
	return emwinPatterns[productKey]
}