package config

// Config struct to hold the configuration values for reaching the S3 storage
type Config struct {
    BucketName       string `env:"BUCKET_NAME,default=goes-data"`  // Single bucket for all satellite data
    S3Endpoint       string `env:"S3_ENDPOINT,required"`
    AccessKeyID      string `env:"ACCESS_KEY_ID,required"`
    SecretAccessKey  string `env:"SECRET_ACCESS_KEY,required"`
    Port             string `env:"PORT,default=3000"`
    UseSSLforS3      bool   `env:"USE_SSL_FOR_S3,default=true"`
    TrustedProxies   string `env:"TRUSTED_PROXIES,default="`      // Comma-separated list of trusted proxies, if any
}

