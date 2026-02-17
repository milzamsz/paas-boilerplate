package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Provider implements Service using AWS S3 or S3-compatible stores (MinIO, R2, etc.).
type S3Provider struct {
	client   *s3.Client
	bucket   string
	endpoint string // e.g. "https://s3.us-east-1.amazonaws.com" or MinIO URL
	region   string
}

// S3Config holds configuration for the S3 provider.
type S3Config struct {
	Endpoint        string // e.g. "http://localhost:9000" for MinIO
	Region          string // e.g. "us-east-1"
	Bucket          string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool   // true for MinIO, false for AWS S3
	PublicURL       string // optional: custom public URL prefix (e.g. CDN)
}

// NewS3Provider creates a new S3-compatible storage provider.
func NewS3Provider(ctx context.Context, cfg S3Config) (*S3Provider, error) {
	resolver := aws.EndpointResolverWithOptionsFunc(
		func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			if cfg.Endpoint != "" {
				return aws.Endpoint{
					URL:               cfg.Endpoint,
					HostnameImmutable: cfg.UsePathStyle,
				}, nil
			}
			return aws.Endpoint{}, &aws.EndpointNotFoundError{}
		},
	)

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
		awsconfig.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	})

	return &S3Provider{
		client:   client,
		bucket:   cfg.Bucket,
		endpoint: cfg.Endpoint,
		region:   cfg.Region,
	}, nil
}

// Upload stores a file in S3.
func (p *S3Provider) Upload(ctx context.Context, key string, reader io.Reader, contentType string, size int64) (*FileInfo, error) {
	input := &s3.PutObjectInput{
		Bucket:        aws.String(p.bucket),
		Key:           aws.String(key),
		Body:          reader,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
		CacheControl:  aws.String("public, max-age=31536000, immutable"),
	}

	if _, err := p.client.PutObject(ctx, input); err != nil {
		return nil, fmt.Errorf("s3: failed to upload %s: %w", key, err)
	}

	return &FileInfo{
		Key:         key,
		ContentType: contentType,
		Size:        size,
	}, nil
}

// Delete removes a file from S3.
func (p *S3Provider) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}

	if _, err := p.client.DeleteObject(ctx, input); err != nil {
		return fmt.Errorf("s3: failed to delete %s: %w", key, err)
	}
	return nil
}

// GetPresignedURL returns a time-limited download URL.
func (p *S3Provider) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(p.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(key),
	}

	result, err := presignClient.PresignGetObject(ctx, input, func(o *s3.PresignOptions) {
		o.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("s3: failed to presign %s: %w", key, err)
	}

	return result.URL, nil
}

// GetPublicURL returns the public URL for a key.
// Uses the endpoint + bucket + key pattern.
func (p *S3Provider) GetPublicURL(key string) string {
	if p.endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", p.endpoint, p.bucket, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", p.bucket, p.region, key)
}

// detectContentType sniffs the content type from the first 512 bytes.
// Falls back to "application/octet-stream" if unknown.
func detectContentType(data []byte) string {
	ct := http.DetectContentType(data)
	return ct
}
