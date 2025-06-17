package graphql

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/raywall/cloud-easy-connector/pkg/cloud"
)

type PathType string

const (
	LOCAL       PathType = "local"
	ENV_VARS    PathType = "env_vars"
	AWS_SSM     PathType = "aws_ssm"
	AWS_SECRETS PathType = "aws_secrets"
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
			path = Path{
				Type: AWS_SSM,
				Path: parts[1],
				Args: make(map[string]interface{}),
			}

			path.Args["enc"] = false
			if len(parts) > 2 {
				path.Args["enc"] = (parts[2] == "true")
			}

		case "secrets":
			path = Path{
				Type: AWS_SECRETS,
				Path: parts[1],
				Args: make(map[string]interface{}),
			}

			path.Args["type"] = "text"
			if len(parts) > 2 {
				path.Args["type"] = parts[2]
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
		strings.HasPrefix(config, "secret:"))
}
