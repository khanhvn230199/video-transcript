package uploads

import (
	"context"
	"fmt"
	"io"

	"video-transcript/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

var r2Client *s3.Client

// InitR2 khởi tạo S3 client trỏ vào Cloudflare R2 (dùng endpoint + key trong .env).
func InitR2(ctx context.Context) error {
	if config.SvcCfg.AWS_ACCESS_KEY_ID == "" || config.SvcCfg.AWS_SECRET_ACCESS_KEY == "" || config.SvcCfg.AWS_CUSTOM_ENDPOINT == "" {
		zap.S().Error("R2 config is missing (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY / AWS_CUSTOM_ENDPOINT)")
		return fmt.Errorf("R2 config is missing (AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY / AWS_CUSTOM_ENDPOINT)")
	}

	endpoint := config.SvcCfg.AWS_CUSTOM_ENDPOINT

	zap.S().Infow("Init R2",
		"endpoint", endpoint,
		"region", config.SvcCfg.AWS_REGION,
		"bucket", config.SvcCfg.BUCKET_KEY,
	)

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           endpoint,
			PartitionID:   "aws",
			SigningRegion: config.SvcCfg.AWS_REGION,
		}, nil
	})

	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(config.SvcCfg.AWS_REGION),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				config.SvcCfg.AWS_ACCESS_KEY_ID,
				config.SvcCfg.AWS_SECRET_ACCESS_KEY,
				"",
			),
		),
		awsconfig.WithEndpointResolverWithOptions(resolver),
	)
	if err != nil {
		zap.S().Error("load R2 config: %w", err)
		return fmt.Errorf("load R2 config: %w", err)
	}

	r2Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return nil
}

// UploadToR2 upload dữ liệu lên bucket R2, trả về public URL (dựa trên AWS_BASE_URL + key).
func UploadToR2(ctx context.Context, key string, body io.Reader, size int64, contentType string) (string, error) {
	if r2Client == nil {
		zap.S().Error("R2 client is not initialized")
		return "", fmt.Errorf("R2 client is not initialized")
	}

	zap.S().Infow("UploadToR2 start",
		"bucket", config.SvcCfg.BUCKET_KEY,
		"key", key,
		"size", size,
		"content_type", contentType,
	)
	if config.SvcCfg.BUCKET_KEY == "" {
		zap.S().Error("R2 bucket (BUCKET_KEY) is not configured")
		return "", fmt.Errorf("R2 bucket (BUCKET_KEY) is not configured")
	}

	out, err := r2Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(config.SvcCfg.BUCKET_KEY),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		zap.S().Errorw("upload to R2 failed",
			"error", err,
			"bucket", config.SvcCfg.BUCKET_KEY,
			"key", key,
		)
		return "", fmt.Errorf("upload to R2: %w", err)
	}

	zap.S().Infow("UploadToR2 success",
		"bucket", config.SvcCfg.BUCKET_KEY,
		"key", key,
		"etag", aws.ToString(out.ETag),
	)

	baseURL := config.SvcCfg.AWS_BASE_URL
	if baseURL == "" {
		// Nếu không cấu hình sẵn base URL, chỉ trả về key.
		zap.S().Error("base URL is not configured")
		return key, nil
	}

	return fmt.Sprintf("%s/%s", baseURL, key), nil
}
