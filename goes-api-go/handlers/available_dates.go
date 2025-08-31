package handlers

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"goes-api-go/s3"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func GetAvailableDates(s3Client *s3.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		cstLocation, err := time.LoadLocation("America/Chicago")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load timezone"})
			return
		}

		// Just check one representative product path that's likely to have data
		// This is much faster than scanning all products
		prefix := "goes19/fd/fc/" // GOES-19 Full Disk Color - usually has good coverage
		
		ctx := context.Background()

		// Create a map to store dates by their CST representation
		datesSet := make(map[string]struct{})

		// List objects but not recursively - just get the date directories
		objectCh := s3Client.Client.ListObjects(ctx, s3Client.BucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: false,
		})

		for object := range objectCh {
			if object.Err != nil {
				continue
			}

			// Extract date from the directory structure
			// Expected format: goes19/fd/fc/2024-12-25/
			parts := strings.Split(object.Key, "/")
			if len(parts) >= 4 {
				dateStr := parts[3]
				// Validate it's a date format
				if len(dateStr) == 10 && strings.Count(dateStr, "-") == 2 {
					datesSet[dateStr] = struct{}{}
				}
			}
		}

		// If we didn't find dates using directory listing, fall back to checking files
		if len(datesSet) == 0 {
			objectCh = s3Client.Client.ListObjects(ctx, s3Client.BucketName, minio.ListObjectsOptions{
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

				// Extract UTC timestamp from the object key
				utcTime, err := extractTimestamp(object.Key)
				if err != nil {
					continue
				}

				// Convert UTC time to CST
				cstTime := utcTime.In(cstLocation)

				// Store the CST date
				cstDate := cstTime.Format("2006-01-02")
				datesSet[cstDate] = struct{}{}
				
				// Limit to recent dates to avoid scanning too many files
				if len(datesSet) >= 30 {
					break
				}
			}
		}

		// Convert the set to a sorted slice
		dates := make([]string, 0, len(datesSet))
		for date := range datesSet {
			dates = append(dates, date)
		}

		// Sort dates in descending order
		sort.Slice(dates, func(i, j int) bool {
			return dates[i] > dates[j] // Simple string comparison is sufficient for YYYY-MM-DD format
		})

		c.JSON(http.StatusOK, gin.H{"availableDates": dates})
	}
}
