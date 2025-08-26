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

func GetImagesByDate(s3Client *s3.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		cstLocation, err := time.LoadLocation("America/Chicago")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load timezone"})
			return
		}

		requestedDate := c.Param("date")
		cstMidnight, err := time.ParseInLocation("2006-01-02", requestedDate, cstLocation)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
			return
		}

		utcStart := cstMidnight.UTC()
		utcEnd := cstMidnight.Add(24 * time.Hour).UTC()
		utcDates := []string{
			utcStart.Format("2006-01-02"),
			utcEnd.Format("2006-01-02"),
		}

		ctx := context.Background()
		var imageUrls []string

		for _, utcDate := range utcDates {
			prefix := "false-color/fd/" + utcDate + "/"

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
					continue
				}

				// Convert to CST and check if it falls on our requested date
				cstTime := utcTime.In(cstLocation)
				if cstTime.Format("2006-01-02") == requestedDate {
					imageUrls = append(imageUrls, s3Client.BaseURL+object.Key)
				}
			}
		}

		// Sort URLs chronologically
		sort.Strings(imageUrls)

		c.JSON(http.StatusOK, gin.H{"imageUrls": imageUrls})
	}
}
