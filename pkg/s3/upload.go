package s3

import (
    "context"
    "fmt"
    "os"
    "path"
    "bytes"

    "github.com/aws/aws-sdk-go-v2/service/s3"
    s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func UploadBytes(ctx context.Context, client *s3.Client, bucket string, key string, data []byte, contentType string) (string, error) {
    _, err := client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: &bucket,
        Key:    &key,
        Body:   bytesReader(data),
        ContentType: &contentType,
        ACL:    s3types.ObjectCannedACLPrivate,
    })
    if err != nil { return "", err }

    endpoint := os.Getenv("S3_PUBLIC_BASE")
    if endpoint == "" {
        endpoint = "https://storage.yandexcloud.net"
    }
    return fmt.Sprintf("%s/%s/%s", endpoint, bucket, path.Clean(key)), nil
}

// local helper to avoid importing bytes widely
func bytesReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }


