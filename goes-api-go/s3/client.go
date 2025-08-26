package s3

import (
	"context"
	"fmt"
	cfg "goes-api-go/config"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Client is a wrapper around the Minio client, providing additional context like the bucket name and base URL
type S3Client struct {
	Client     *minio.Client // The Minio client used for interacting with the S3-compatible storage
	BucketName string        // The name of the S3 bucket containing all satellite data
	BaseURL    string        // The base URL for accessing objects in the bucket
}

// NewS3Client initializes and returns an S3Client configured to interact with the specified S3-compatible storage
func NewS3Client(ctx context.Context, cfg *cfg.Config) (*S3Client, error) {
	// Create the MinIO client with the provided endpoint, access key, and secret key
	client, err := minio.New(cfg.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),	// Use static credentials for authentication
		Secure: cfg.UseSSLforS3,	// Indicates whether to use a secure (HTTPS) connection
	})
	if err != nil {
		// Log a fatal error message and return the error if the client creation fails
		log.Fatalf("Failed to create MinIO client: %v", err)
		return nil, err
	}

	// Determine the protocol based on the UseSSLforS3 setting
	protocol := "http"
	if cfg.UseSSLforS3 {
		protocol = "https"
	}

	// Construct the base URL for the S3 bucket
	baseURL := fmt.Sprintf("%s://%s/%s/", protocol, cfg.S3Endpoint, cfg.BucketName)
	
	// Initialize and return the S3Client
	return &S3Client{
		Client:     client,
		BucketName: cfg.BucketName,
		BaseURL:    baseURL,
	}, nil
}
