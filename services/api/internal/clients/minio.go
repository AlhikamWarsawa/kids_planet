package clients

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
)

type MinIO struct {
	cli *minio.Client
}

func NewMinIO(ctx context.Context, cfg config.MinIOConfig) (*MinIO, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	endpoint, secure, err := normalizeMinioEndpoint(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("minio new: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if _, err := cli.ListBuckets(pingCtx); err != nil {
		return nil, fmt.Errorf("minio ping failed: %w", err)
	}

	return &MinIO{cli: cli}, nil
}

func (m *MinIO) PutObject(ctx context.Context, bucket, objectKey string, reader io.Reader, size int64, contentType string) (string, error) {
	opts := minio.PutObjectOptions{
		ContentType: strings.TrimSpace(contentType),
	}
	info, err := m.cli.PutObject(ctx, bucket, objectKey, reader, size, opts)
	if err != nil {
		return "", err
	}
	return info.ETag, nil
}

func normalizeMinioEndpoint(raw string) (endpoint string, secure bool, err error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", false, fmt.Errorf("minio endpoint is empty")
	}

	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		u, err := url.Parse(s)
		if err != nil {
			return "", false, fmt.Errorf("invalid MINIO_ENDPOINT=%q: %w", raw, err)
		}
		if u.Host == "" {
			return "", false, fmt.Errorf("invalid MINIO_ENDPOINT=%q: missing host", raw)
		}
		return u.Host, u.Scheme == "https", nil
	}

	return s, false, nil
}
