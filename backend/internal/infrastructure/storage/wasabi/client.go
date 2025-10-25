package wasabi

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	s3Client *s3.Client
	config   *Config
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = true
	})

	return &Client{
		s3Client: s3Client,
		config:   cfg,
	}, nil
}

type PresignService struct {
	presignClient *s3.PresignClient
	config        *Config
}

func NewPresignService(client *Client) *PresignService {
	return &PresignService{
		presignClient: s3.NewPresignClient(client.s3Client),
		config:        client.config,
	}
}

func (p *PresignService) GenerateUploadURL(key, contentType string, sizeBytes int64, expires interface{}) (string, error) {
	var duration time.Duration
	if expires == nil {
		duration = p.config.UploadExpires
	} else if d, ok := expires.(time.Duration); ok {
		duration = d
	} else {
		duration = p.config.UploadExpires
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(p.config.BucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}

	request, err := p.presignClient.PresignPutObject(context.Background(), input, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return "", err
	}

	return request.URL, nil
}

func (p *PresignService) GenerateDownloadURL(key string, expires interface{}) (string, error) {
	var duration time.Duration
	if expires == nil {
		duration = p.config.DownloadExpires
	} else if d, ok := expires.(time.Duration); ok {
		duration = d
	} else {
		duration = p.config.DownloadExpires
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(p.config.BucketName),
		Key:    aws.String(key),
	}

	request, err := p.presignClient.PresignGetObject(context.Background(), input, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return "", err
	}

	return request.URL, nil
}
