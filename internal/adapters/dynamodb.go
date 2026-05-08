package adapters

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBAdapter struct {
	client *dynamodb.Client
	table  string
}

func NewDynamoDBAdapter(region, table, accessKeyId, secretAccessKey string) (Adapter, error) {
	if region == "" {
		return nil, fmt.Errorf("dynamodb region is required")
	}
	if table == "" {
		return nil, fmt.Errorf("dynamodb table is required")
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

	return &DynamoDBAdapter{
		client: dynamodb.NewFromConfig(cfg),
		table:  table,
	}, nil
}

func (d *DynamoDBAdapter) GetData(ctx context.Context, key string) (map[string]interface{}, error) {
	result, err := d.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(d.table),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item %s from DynamoDB: %v", key, err)
	}
	if result.Item == nil {
		return nil, fmt.Errorf("item %s not found in DynamoDB", key)
	}

	var data map[string]interface{}
	if err := attributevalue.UnmarshalMap(result.Item, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DynamoDB item %s: %v", key, err)
	}
	return data, nil
}
