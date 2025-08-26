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

		prefix := "false-color/fd/"
		ctx := context.Background()

		// Create a map to store dates by their CST representation
		datesSet := make(map[string]struct{})

		objectCh := s3Client.Client.ListObjects(ctx, s3Client.BucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: true,
		})

		for object := range objectCh {
			if object.Err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": object.Err.Error()})
				return
			}

			// Skip non-image files
			if !strings.HasSuffix(strings.ToLower(object.Key), ".png") &&
				!strings.HasSuffix(strings.ToLower(object.Key), ".jpg") {
				continue
			}

			// Extract UTC timestamp from the object key
			utcTime, err := extractTimestamp(object.Key)
			if err != nil {
				continue // Skip files that don't match our expected format
			}

			// Convert UTC time to CST
			cstTime := utcTime.In(cstLocation)

			// Store the CST date
			cstDate := cstTime.Format("2006-01-02")
			datesSet[cstDate] = struct{}{}
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
