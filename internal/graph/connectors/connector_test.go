package connectors

import "testing"

func TestRenderTemplateUsesDynamicArgs(t *testing.T) {
	got, err := renderTemplate("/customers/{customerID}/contracts/{contractID}", map[string]interface{}{
		"customerID": 10,
		"contractID": "ABC",
	})
	if err != nil {
		t.Fatalf("renderTemplate returned error: %v", err)
	}
	if got != "/customers/10/contracts/ABC" {
		t.Fatalf("renderTemplate = %q", got)
	}
}

func TestRenderTemplateRequiresArgs(t *testing.T) {
	if _, err := renderTemplate("CVN_{codigoConvenio}", map[string]interface{}{}); err == nil {
		t.Fatal("expected missing argument error")
	}
}

func TestNewConnectorValidatesRedisEndpoint(t *testing.T) {
	_, err := NewConnector(ConnectorConfig{
		Field:      "convenio",
		Adapter:    "redis",
		KeyPattern: "CVN_{codigoConvenio}",
	})
	if err == nil {
		t.Fatal("expected redis endpoint validation error")
	}
}

func TestNewConnectorValidatesRDSConfig(t *testing.T) {
	_, err := NewConnector(ConnectorConfig{
		Field:      "convenio",
		Adapter:    "rds",
		KeyPattern: "select * from convenio where codigo = {codigoConvenio}",
		AdapterConfig: map[string]interface{}{
			"driverName": "postgres",
		},
	})
	if err == nil {
		t.Fatal("expected rds dsn validation error")
	}
}
