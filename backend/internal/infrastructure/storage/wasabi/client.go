package wasabi

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/newt239/chat/internal/infrastructure/config"
)

// PresignService は S3 互換のストレージに署名付き URL でアップロード・ダウンロードさせます
type PresignService struct {
	presignClient *s3.PresignClient
	s3Client      *s3.Client
	bucket        string
}

func NewPresignService(cfg config.WasabiConfig) *PresignService {
	s3Client := s3.New(s3.Options{
		Region:       cfg.Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		UsePathStyle: true,
		BaseEndpoint: aws.String(cfg.Endpoint),
	})
	return &PresignService{presignClient: s3.NewPresignClient(s3Client), s3Client: s3Client, bucket: cfg.BucketName}
}

func (p *PresignService) DeleteObject(ctx context.Context, key string) error {
	_, err := p.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(p.bucket), Key: aws.String(key)})
	return err
}

func (p *PresignService) GenerateUploadURL(ctx context.Context, key, contentType string, sizeBytes int64, expires time.Duration) (string, error) {
	request, err := p.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(p.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(sizeBytes),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return request.URL, nil
}

func (p *PresignService) GenerateDownloadURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	request, err := p.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return request.URL, nil
}
