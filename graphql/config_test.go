package graphql

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/raywall/cloud-easy-connector/pkg/cloud"
)

func TestLoadConfigReadsUnifiedConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "service.json")
	mustWrite(t, configPath, `{
		"schema": "local:`+filepath.Join(dir, "schema.json")+`",
		"connectors": "local:`+filepath.Join(dir, "connectors.json")+`",
		"mock": "local:`+filepath.Join(dir, "mock.json")+`",
		"route": "/variables",
		"pretty": true,
		"graphiql": true,
		"allow_partial": true
	}`)
	mustWrite(t, filepath.Join(dir, "schema.json"), `{"types":[],"query":{"name":"Query","fields":[]}}`)
	mustWrite(t, filepath.Join(dir, "connectors.json"), `{"connectors":[]}`)
	mustWrite(t, filepath.Join(dir, "mock.json"), `{"status":true,"values":{}}`)

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if config.Route != "/variables" {
		t.Fatalf("Route = %q", config.Route)
	}
	if !config.AllowPartial {
		t.Fatal("expected AllowPartial to be true")
	}
	if value, err := config.GetSchemaValue(); err != nil || value == "" {
		t.Fatalf("GetSchemaValue = %q, %v", value, err)
	}
}

func TestEnvPathUsesDefaultValue(t *testing.T) {
	t.Setenv("GRAPHQL_CONNECTOR_TEST", "")
	if err := os.Unsetenv("GRAPHQL_CONNECTOR_TEST"); err != nil {
		t.Fatalf("Unsetenv returned error: %v", err)
	}

	value, err := FromString("env:GRAPHQL_CONNECTOR_TEST:fallback").GetValue(nil)
	if err != nil {
		t.Fatalf("GetValue returned error: %v", err)
	}
	if value != "fallback" {
		t.Fatalf("value = %q", value)
	}
}

func TestFromStringSupportsS3ConfigPath(t *testing.T) {
	path := FromString("s3:us-east-1:config-bucket:graphql/schema.json")
	if path == nil {
		t.Fatal("expected s3 path")
	}
	if path.Type != AWS_S3 {
		t.Fatalf("Type = %q", path.Type)
	}
	if path.Args["region"] != "us-east-1" {
		t.Fatalf("region = %v", path.Args["region"])
	}
	if path.Args["bucket"] != "config-bucket" {
		t.Fatalf("bucket = %v", path.Args["bucket"])
	}
	if path.Path != "graphql/schema.json" {
		t.Fatalf("Path = %q", path.Path)
	}
}

func TestFromStringSupportsDynamoDBConfigPath(t *testing.T) {
	path := FromString("dynamodb:us-east-1:graphql-config:schema:configKey:payload")
	if path == nil {
		t.Fatal("expected dynamodb path")
	}
	if path.Type != AWS_DYNAMO {
		t.Fatalf("Type = %q", path.Type)
	}
	if path.Args["table"] != "graphql-config" {
		t.Fatalf("table = %v", path.Args["table"])
	}
	if path.Path != "schema" {
		t.Fatalf("Path = %q", path.Path)
	}
	if path.Args["keyAttribute"] != "configKey" {
		t.Fatalf("keyAttribute = %v", path.Args["keyAttribute"])
	}
	if path.Args["valueAttribute"] != "payload" {
		t.Fatalf("valueAttribute = %v", path.Args["valueAttribute"])
	}
}

func TestNewPreservesConfigDefaults(t *testing.T) {
	config := &Config{
		Schema: `{
			"types": [
				{
					"name": "CombinedData",
					"fields": [
						{"name": "status", "type": "String"}
					]
				}
			],
			"query": {
				"name": "Query",
				"fields": [
					{"name": "dataSources", "type": "Object", "ofType": "CombinedData"}
				]
			}
		}`,
		Connectors: `{"connectors":[]}`,
	}

	api, err := New(config, testResources(), "us-east-1", "http://localhost:4566")
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	if api.Config.Route != "/graphql" {
		t.Fatalf("Route = %q", api.Config.Route)
	}
	if !api.Config.Pretty {
		t.Fatal("expected Pretty default to be true")
	}
}

func testResources() *cloud.CloudContextList {
	return &cloud.CloudContextList{cloud.SSMContext}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
