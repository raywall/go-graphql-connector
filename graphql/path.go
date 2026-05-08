package graphql

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/raywall/cloud-easy-connector/pkg/cloud"
)

type PathType string

const (
	LOCAL       PathType = "local"
	ENV_VARS    PathType = "env_vars"
	AWS_SSM     PathType = "aws_ssm"
	AWS_SECRETS PathType = "aws_secrets"
	AWS_S3      PathType = "aws_s3"
	AWS_DYNAMO  PathType = "aws_dynamodb"
)

type Path struct {
	Args         map[string]interface{} `json:"args"`
	DefaultValue string                 `json:"default_value"`
	Path         string                 `json:"path"`
	Type         PathType               `json:"type"`
}

// GetValue is a method capable of retrieve a value from a local file or a cloud service using a
// cloud context of the cloud easy connector library
func (p *Path) GetValue(ctx cloud.CloudContext) (interface{}, error) {
	switch p.Type {
	case LOCAL:
		content, err := os.ReadFile(p.Path)
		if err != nil {
			return "", fmt.Errorf("failed to get value on %s: %v", p.Path, err)
		}
		return string(content), nil

	case ENV_VARS:
		value, exists := os.LookupEnv(p.Path)
		if !exists {
			if p.DefaultValue != "" {
				return p.DefaultValue, nil
			}
			return "", fmt.Errorf("environment variable %s was not found", p.Path)
		}
		return string(value), nil

	case AWS_SSM:
		value, err := ctx.GetParameterValue(p.Path, p.Args["enc"].(bool))
		if err != nil {
			return "", fmt.Errorf("failed to get value on AWS SSM parameter %s: %v", p.Path, err)
		}
		return value, nil

	case AWS_SECRETS:
		value, err := ctx.GetSecretValue(p.Path, cloud.SecretType(p.Args["type"].(string)))
		if err != nil {
			return "", fmt.Errorf("failed to get value on AWS Secrets Manager %s: %v", p.Path, err)
		}
		if _, ok := value.(string); !ok {
			obj := make(map[string]interface{})
			err := json.Unmarshal(value.([]byte), &obj)
			if err != nil {
				return "", fmt.Errorf("failed to get value on AWS Secrets Manager %s: %v", p.Path, err)
			}
			return obj, nil
		}
		return value, nil

	case AWS_S3:
		value, err := getS3ConfigValue(context.Background(), p)
		if err != nil {
			return "", err
		}
		return value, nil

	case AWS_DYNAMO:
		value, err := getDynamoDBConfigValue(context.Background(), p)
		if err != nil {
			return "", err
		}
		return value, nil
	}

	return "", fmt.Errorf("configure type is not supported")
}

// FromString is a method capable of load a path from an inline path configuration
func FromString(config string) *Path {
	if !IsPath(config) {
		return nil
	}

	var (
		parts []string = strings.Split(config, ":")
		path  Path
	)

	if len(parts) > 0 {
		switch parts[0] {
		case "env":
			if len(parts) < 2 {
				return nil
			}
			path = Path{
				Type: ENV_VARS,
				Path: parts[1],
			}

			if len(parts) > 2 {
				path.DefaultValue = parts[2]
			}

		case "local":
			path = Path{
				Type: LOCAL,
				Path: strings.ReplaceAll(config, "local:", ""),
			}

		case "ssm":
			if len(parts) < 2 {
				return nil
			}
			path = Path{
				Type: AWS_SSM,
				Path: parts[1],
				Args: make(map[string]interface{}),
			}

			path.Args["enc"] = false
			if len(parts) > 2 {
				path.Args["enc"] = (parts[2] == "true")
			}

		case "secret", "secrets":
			if len(parts) < 2 {
				return nil
			}
			path = Path{
				Type: AWS_SECRETS,
				Path: parts[1],
				Args: make(map[string]interface{}),
			}

			path.Args["type"] = "text"
			if len(parts) > 2 {
				path.Args["type"] = parts[2]
			}

		case "s3":
			if len(parts) < 4 {
				return nil
			}
			path = Path{
				Type: AWS_S3,
				Path: parts[3],
				Args: map[string]interface{}{
					"region": parts[1],
					"bucket": parts[2],
				},
			}

		case "dynamodb":
			if len(parts) < 4 {
				return nil
			}
			path = Path{
				Type: AWS_DYNAMO,
				Path: parts[3],
				Args: map[string]interface{}{
					"region":         parts[1],
					"table":          parts[2],
					"keyAttribute":   "id",
					"valueAttribute": "value",
				},
			}
			if len(parts) > 4 {
				path.Args["keyAttribute"] = parts[4]
			}
			if len(parts) > 5 {
				path.Args["valueAttribute"] = parts[5]
			}
		}
	}

	return &path
}

// IsPath checks if a string value is an inline path
func IsPath(config string) bool {
	return (strings.HasPrefix(config, "env:") ||
		strings.HasPrefix(config, "ssm:") ||
		strings.HasPrefix(config, "local:") ||
		strings.HasPrefix(config, "secret:") ||
		strings.HasPrefix(config, "secrets:") ||
		strings.HasPrefix(config, "s3:") ||
		strings.HasPrefix(config, "dynamodb:"))
}

func getS3ConfigValue(ctx context.Context, path *Path) (string, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(fmt.Sprintf("%v", path.Args["region"])))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config for S3: %v", err)
	}

	result, err := s3.NewFromConfig(cfg).GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(fmt.Sprintf("%v", path.Args["bucket"])),
		Key:    aws.String(path.Path),
	})
	if err != nil {
		return "", fmt.Errorf("failed to get S3 config object %s: %v", path.Path, err)
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read S3 config object %s: %v", path.Path, err)
	}
	return string(body), nil
}

func getDynamoDBConfigValue(ctx context.Context, path *Path) (string, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(fmt.Sprintf("%v", path.Args["region"])))
	if err != nil {
		return "", fmt.Errorf("failed to load AWS config for DynamoDB: %v", err)
	}

	keyAttribute := fmt.Sprintf("%v", path.Args["keyAttribute"])
	valueAttribute := fmt.Sprintf("%v", path.Args["valueAttribute"])
	result, err := dynamodb.NewFromConfig(cfg).GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(fmt.Sprintf("%v", path.Args["table"])),
		Key: map[string]types.AttributeValue{
			keyAttribute: &types.AttributeValueMemberS{Value: path.Path},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get DynamoDB config item %s: %v", path.Path, err)
	}
	if result.Item == nil {
		return "", fmt.Errorf("DynamoDB config item %s not found", path.Path)
	}

	var item map[string]interface{}
	if err := attributevalue.UnmarshalMap(result.Item, &item); err != nil {
		return "", fmt.Errorf("failed to unmarshal DynamoDB config item %s: %v", path.Path, err)
	}

	value, ok := item[valueAttribute]
	if !ok {
		encoded, err := json.Marshal(item)
		if err != nil {
			return "", fmt.Errorf("failed to marshal DynamoDB config item %s: %v", path.Path, err)
		}
		return string(encoded), nil
	}
	if text, ok := value.(string); ok {
		return text, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal DynamoDB config value %s: %v", path.Path, err)
	}
	return string(encoded), nil
}
