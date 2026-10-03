package wasabi

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// PresignService は S3 互換のストレージに署名付き URL でアップロード・ダウンロードさせます
type PresignService struct {
	presignClient *s3.PresignClient
	s3Client      *s3.Client
	config        *Config
}

func NewPresignService(cfg *Config) *PresignService {
	options := s3.Options{
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		UsePathStyle: true,
	}
	if cfg.Endpoint != "" {
		options.BaseEndpoint = aws.String(cfg.Endpoint)
	}
	s3Client := s3.New(options)
	return &PresignService{
		presignClient: s3.NewPresignClient(s3Client),
		s3Client:      s3Client,
		config:        cfg,
	}
}

// DeleteObject はストレージ上のオブジェクトを削除します
func (p *PresignService) DeleteObject(ctx context.Context, key string) error {
	_, err := p.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.config.BucketName),
		Key:    aws.String(key),
	})
	return err
}

func (p *PresignService) GenerateUploadURL(ctx context.Context, key, contentType string, _ int64, expires time.Duration) (string, error) {
	request, err := p.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(p.config.BucketName),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(durationOr(expires, p.config.UploadExpires)))
	if err != nil {
		return "", err
	}
	return request.URL, nil
}

func (p *PresignService) GenerateDownloadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	request, err := p.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.config.BucketName),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(durationOr(expires, p.config.DownloadExpires)))
	if err != nil {
		return "", err
	}
	return request.URL, nil
}

func durationOr(d, fallback time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return fallback
}
