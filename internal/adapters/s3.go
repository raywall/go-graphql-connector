package adapters

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Adapter struct {
	client *s3.Client
	bucket string
}

func NewS3Adapter(region, bucket, accessKeyId, secretAccessKey string) (Adapter, error) {
	if region == "" {
		return nil, fmt.Errorf("s3 region is required")
	}
	if bucket == "" {
		return nil, fmt.Errorf("s3 bucket is required")
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     accessKeyId,
				SecretAccessKey: secretAccessKey,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	return &S3Adapter{
		client: s3.NewFromConfig(cfg),
		bucket: bucket,
	}, nil
}

func (s *S3Adapter) GetData(ctx context.Context, key string) (map[string]interface{}, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s from S3: %v", key, err)
	}
	defer result.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(result.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode S3 object %s: %v", key, err)
	}
	return data, nil
}
