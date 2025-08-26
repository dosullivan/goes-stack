package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"goes-api-go/s3"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func GetLatestImage(s3Client *s3.S3Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		maxDaysBack := 7
		ctx := context.Background()

		cstLocation, err := time.LoadLocation("America/Chicago")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load timezone"})
			return
		}

		var latestTime time.Time
		var latestObjectKey string

		// Get current time in CST
		nowCST := time.Now().In(cstLocation)

		// Check today, tomorrow (for UTC crossover), and previous days
		for i := -1; i < maxDaysBack; i++ { // Start at -1 to include tomorrow
			checkDate := nowCST.AddDate(0, 0, -i).Format("2006-01-02")
			prefix := "false-color/fd/" + checkDate + "/"

			objectCh := s3Client.Client.ListObjects(ctx, s3Client.BucketName, minio.ListObjectsOptions{
				Prefix:    prefix,
				Recursive: true,
			})

			for object := range objectCh {
				if object.Err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": object.Err.Error()})
					return
				}

				if !strings.HasSuffix(strings.ToLower(object.Key), ".png") &&
					!strings.HasSuffix(strings.ToLower(object.Key), ".jpg") {
					continue
				}

				utcTime, err := extractTimestamp(object.Key)
				if err != nil {
					continue
				}

				cstTime := utcTime.In(cstLocation)

				if latestTime.IsZero() || cstTime.After(latestTime) {
					latestTime = cstTime
					latestObjectKey = object.Key
				}
			}
		}

		if latestObjectKey == "" {
			c.JSON(http.StatusNotFound, gin.H{"error": "No recent images found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"imageUrl": s3Client.BaseURL + latestObjectKey,
		})
	}
}
