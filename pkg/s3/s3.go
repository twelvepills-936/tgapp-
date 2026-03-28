package s3

import (
    "context"
    "os"

    "github.com/aws/aws-sdk-go-v2/aws"
    awsCfg "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/credentials"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewClient(ctx context.Context) (*s3.Client, error) {
    region := os.Getenv("S3_REGION")
    endpoint := os.Getenv("S3_ENDPOINT")
    accessKey := os.Getenv("S3_ACCESS_KEY_ID")
    secretKey := os.Getenv("S3_SECRET_ACCESS_KEY")

    optFns := []func(*awsCfg.LoadOptions) error{}
    if region != "" {
        optFns = append(optFns, awsCfg.WithRegion(region))
    }
    if accessKey != "" && secretKey != "" {
        optFns = append(optFns, awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")))
    }
    if endpoint != "" {
        optFns = append(optFns, awsCfg.WithEndpointResolverWithOptions(
            aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
                return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
            })))
    }
    c, err := awsCfg.LoadDefaultConfig(ctx, optFns...)
    if err != nil { return nil, err }
    return s3.NewFromConfig(c, func(o *s3.Options) { o.UsePathStyle = true }), nil
}


