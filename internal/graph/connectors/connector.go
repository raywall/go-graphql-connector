package connectors

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/raywall/go-graphql-integrator/internal/adapters"
)

type ConnectorConfig struct {
	Field         string                 `json:"field"`
	Adapter       string                 `json:"adapter"`
	AdapterConfig map[string]interface{} `json:"adapterConfig"`
	KeyPattern    string                 `json:"keyPattern"`
}

type Config struct {
	Connectors []ConnectorConfig `json:"connectors"`
}

type Connector interface {
	GetData(codigoConvenio int) (map[string]interface{}, error)
}

type connector struct {
	adapter    adapters.Adapter
	keyPattern string
}

func NewConnector(config ConnectorConfig) (Connector, error) {
	var adapter adapters.Adapter
	switch config.Adapter {
	case "redis":
		endpoint, _ := config.AdapterConfig["endpoint"].(string)
		password, _ := config.AdapterConfig["password"].(string)
		adapter = adapters.NewRedisAdapter(endpoint, password)

	case "s3":
		region, _ := config.AdapterConfig["region"].(string)
		bucket, _ := config.AdapterConfig["bucket"].(string)
		accessKeyId, _ := config.AdapterConfig["accessKeyId"].(string)
		secretAccessKey, _ := config.AdapterConfig["secretAccessKey"].(string)
		adapter = adapters.NewS3Adapter(region, bucket, accessKeyId, secretAccessKey)

	case "dynamodb":
		region, _ := config.AdapterConfig["region"].(string)
		table, _ := config.AdapterConfig["table"].(string)
		accessKeyId, _ := config.AdapterConfig["accessKeyId"].(string)
		secretAccessKey, _ := config.AdapterConfig["secretAccessKey"].(string)
		adapter = adapters.NewDynamoDBAdapter(region, table, accessKeyId, secretAccessKey)

	case "rest":
		baseUrl, _ := config.AdapterConfig["baseUrl"].(string)
		endpoint, _ := config.AdapterConfig["endpoint"].(string)
		adapter = adapters.NewRestAdapter(baseUrl, endpoint)

	default:
		return nil, fmt.Errorf("unsupported adapter: %s", config.Adapter)
	}

	return &connector{
		adapter:    adapter,
		keyPattern: config.KeyPattern,
	}, nil
}

func (c *connector) GetData(codigoConvenio int) (map[string]interface{}, error) {
	key := strings.Replace(c.keyPattern, "{codigoConvenio}", fmt.Sprintf("%d", codigoConvenio), 1)
	return c.adapter.GetData(key)
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
