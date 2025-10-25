package wasabi

import "time"

type Config struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	MaxFileSize     int64
	UploadExpires   time.Duration
	DownloadExpires time.Duration
}

func NewConfig() *Config {
	return &Config{
		MaxFileSize:     1073741824, // 1GB
		UploadExpires:   15 * time.Minute,
		DownloadExpires: 5 * time.Minute,
	}
}

func (c *Config) GetMaxFileSize() int64 {
	return c.MaxFileSize
}

func (c *Config) GetUploadExpires() interface{} {
	return c.UploadExpires
}

func (c *Config) GetDownloadExpires() interface{} {
	return c.DownloadExpires
}
