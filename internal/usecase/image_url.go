package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitlab16.skiftrade.kz/templates/go/pkg/s3"
)

func imageURLForBytes(ctx context.Context, telegramID string, data []byte, mimeType string) (string, error) {
	bucket := strings.TrimSpace(os.Getenv("S3_BUCKET"))
	accessKey := strings.TrimSpace(os.Getenv("S3_ACCESS_KEY_ID"))
	if bucket == "" || accessKey == "" {
		return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
	}

	client, err := s3.NewClient(ctx)
	if err != nil {
		return "", err
	}

	ext := extensionForMime(mimeType)
	key := fmt.Sprintf("generated/%s/%d%s", sanitizePathSegment(telegramID), time.Now().UnixNano(), ext)
	return s3.UploadBytes(ctx, client, bucket, key, data, mimeType)
}

func extensionForMime(mimeType string) string {
	switch strings.ToLower(mimeType) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	default:
		return ".png"
	}
}

func sanitizePathSegment(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "anonymous"
	}
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	return filepath.Clean(s)
}
