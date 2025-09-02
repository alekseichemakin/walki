package s3client

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Client struct {
	S3      *s3.Client
	Presign *s3.PresignClient
	Bucket  string
}

func New(ctx context.Context) (*Client, error) {
	// IAM будет искать AWS_REGION, AWS_ACCESS_KEY_ID / SECRET, а также профиль ~/.aws/
	cfg, err := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		return nil, err
	}

	endpoint := os.Getenv("S3_ENDPOINT") // e.g. "https://storage.yandexcloud.net"
	bucket := os.Getenv("S3_BUCKET")     // e.g. "walki-prod"

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &Client{
		S3:      s3Client,
		Presign: s3.NewPresignClient(s3Client),
		Bucket:  bucket,
	}, nil
}

func (c *Client) PutObject(ctx context.Context, key string, body io.Reader, size int64, mime string) error {
	_, err := c.S3.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(c.Bucket),
		Key:          aws.String(key),
		Body:         body,
		ContentType:  aws.String(mime),
		ACL:          s3types.ObjectCannedACLPrivate,
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	return err
}

func (c *Client) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	out, err := c.Presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.Bucket),
		Key:    aws.String(key),
	}, func(po *s3.PresignOptions) {
		po.Expires = ttl
	})
	if err != nil {
		return "", err
	}
	return out.URL, nil
}
