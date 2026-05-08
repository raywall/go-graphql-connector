package connectors

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/raywall/go-graphql-connector/internal/adapters"
)

type ConnectorConfig struct {
	Field         string                 `json:"field"`
	Adapter       string                 `json:"adapter"`
	AdapterConfig map[string]interface{} `json:"adapterConfig"`
	KeyPattern    string                 `json:"keyPattern"`
	Optional      bool                   `json:"optional,omitempty"`
	TimeoutMS     int                    `json:"timeoutMs,omitempty"`
	Retries       int                    `json:"retries,omitempty"`
}

type Config struct {
	Connectors []ConnectorConfig `json:"connectors"`
}

type Connector interface {
	Field() string
	GetData(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)
	Optional() bool
}

type connector struct {
	field      string
	adapter    adapters.Adapter
	keyPattern string
	optional   bool
	timeout    time.Duration
	retries    int
}

func NewConnector(config ConnectorConfig) (Connector, error) {
	if config.Field == "" {
		return nil, fmt.Errorf("field is required")
	}
	if config.Adapter == "" {
		return nil, fmt.Errorf("adapter is required for field %s", config.Field)
	}
	if config.AdapterConfig == nil {
		config.AdapterConfig = map[string]interface{}{}
	}

	var adapter adapters.Adapter
	var err error
	switch config.Adapter {
	case "redis":
		endpoint, _ := config.AdapterConfig["endpoint"].(string)
		password, _ := config.AdapterConfig["password"].(string)
		adapter, err = adapters.NewRedisAdapter(endpoint, password)

	case "s3":
		region, _ := config.AdapterConfig["region"].(string)
		bucket, _ := config.AdapterConfig["bucket"].(string)
		accessKeyId, _ := config.AdapterConfig["accessKeyId"].(string)
		secretAccessKey, _ := config.AdapterConfig["secretAccessKey"].(string)
		adapter, err = adapters.NewS3Adapter(region, bucket, accessKeyId, secretAccessKey)

	case "dynamodb":
		region, _ := config.AdapterConfig["region"].(string)
		table, _ := config.AdapterConfig["table"].(string)
		accessKeyId, _ := config.AdapterConfig["accessKeyId"].(string)
		secretAccessKey, _ := config.AdapterConfig["secretAccessKey"].(string)
		adapter, err = adapters.NewDynamoDBAdapter(region, table, accessKeyId, secretAccessKey)

	case "rest":
		baseURL, _ := config.AdapterConfig["baseUrl"].(string)
		method, _ := config.AdapterConfig["method"].(string)
		body, _ := config.AdapterConfig["body"].(string)
		headers := stringMap(config.AdapterConfig["headers"])
		adapter, err = adapters.NewRestAdapter(baseURL, method, headers, body)
		if config.KeyPattern == "" {
			config.KeyPattern, _ = config.AdapterConfig["endpoint"].(string)
		}

	case "rds":
		driverName, _ := config.AdapterConfig["driverName"].(string)
		dsn, _ := config.AdapterConfig["dsn"].(string)
		resultMode, _ := config.AdapterConfig["resultMode"].(string)
		adapter, err = adapters.NewRDSAdapter(driverName, dsn, resultMode)
		if config.KeyPattern == "" {
			config.KeyPattern, _ = config.AdapterConfig["query"].(string)
		}

	default:
		return nil, fmt.Errorf("unsupported adapter: %s", config.Adapter)
	}
	if err != nil {
		return nil, err
	}
	if config.KeyPattern == "" {
		return nil, fmt.Errorf("keyPattern is required for field %s", config.Field)
	}
	if config.Retries < 0 {
		return nil, fmt.Errorf("retries cannot be negative for field %s", config.Field)
	}

	timeout := 3 * time.Second
	if config.TimeoutMS > 0 {
		timeout = time.Duration(config.TimeoutMS) * time.Millisecond
	}

	return &connector{
		field:      config.Field,
		adapter:    adapter,
		keyPattern: config.KeyPattern,
		optional:   config.Optional,
		timeout:    timeout,
		retries:    config.Retries,
	}, nil
}

func (c *connector) Field() string {
	return c.field
}

func (c *connector) Optional() bool {
	return c.optional
}

func (c *connector) GetData(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	key, err := renderTemplate(c.keyPattern, args)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for attempt := 0; attempt <= c.retries; attempt++ {
		callCtx, cancel := context.WithTimeout(ctx, c.timeout)
		data, err := c.adapter.GetData(callCtx, key)
		cancel()
		if err == nil {
			return data, nil
		}
		lastErr = err
		if ctx.Err() != nil {
			break
		}
	}

	return nil, lastErr
}

func LoadConnectors(connectorConfig string) (map[string]Connector, error) {
	var config Config
	if err := json.Unmarshal([]byte(connectorConfig), &config); err != nil {
		return nil, fmt.Errorf("error parsing connectors config: %v", err)
	}

	connectors := make(map[string]Connector)
	for _, connConfig := range config.Connectors {
		conn, err := NewConnector(connConfig)
		if err != nil {
			return nil, fmt.Errorf("error creating connector for %s: %v", connConfig.Field, err)
		}
		connectors[connConfig.Field] = conn
	}

	return connectors, nil
}

var templateToken = regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_]*)\}`)

func renderTemplate(pattern string, args map[string]interface{}) (string, error) {
	missing := make([]string, 0)
	rendered := templateToken.ReplaceAllStringFunc(pattern, func(token string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(token, "{"), "}")
		value, ok := args[name]
		if !ok {
			missing = append(missing, name)
			return token
		}
		return fmt.Sprintf("%v", value)
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("missing argument(s) for template %q: %s", pattern, strings.Join(missing, ", "))
	}
	return rendered, nil
}

func stringMap(value interface{}) map[string]string {
	result := map[string]string{}
	switch typed := value.(type) {
	case map[string]string:
		return typed
	case map[string]interface{}:
		for key, value := range typed {
			result[key] = fmt.Sprintf("%v", value)
		}
	}
	return result
}
