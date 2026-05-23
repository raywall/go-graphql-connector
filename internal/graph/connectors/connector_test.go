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

func TestApplyResponseTransformUnwrapsPath(t *testing.T) {
	got, err := applyResponseTransform(map[string]interface{}{
		"data": []interface{}{
			map[string]interface{}{"id": "1"},
		},
		"errors": []interface{}{},
	}, ResponseTransformConfig{
		UnwrapPath:   "data",
		ErrorsPath:   "errors",
		FailOnErrors: true,
	})
	if err != nil {
		t.Fatalf("applyResponseTransform returned error: %v", err)
	}
	items, ok := got.([]interface{})
	if !ok || len(items) != 1 {
		t.Fatalf("applyResponseTransform = %#v", got)
	}
}

func TestApplyResponseTransformFailsOnErrors(t *testing.T) {
	_, err := applyResponseTransform(map[string]interface{}{
		"data":   []interface{}{},
		"errors": []interface{}{"boom"},
	}, ResponseTransformConfig{
		UnwrapPath:   "data",
		ErrorsPath:   "errors",
		FailOnErrors: true,
	})
	if err == nil {
		t.Fatal("expected response errors to fail transform")
	}
}

func TestApplyResponseTransformKeepsOriginalWhenNotConfigured(t *testing.T) {
	data := map[string]interface{}{"data": []interface{}{}}
	got, err := applyResponseTransform(data, ResponseTransformConfig{})
	if err != nil {
		t.Fatalf("applyResponseTransform returned error: %v", err)
	}
	gotMap, ok := got.(map[string]interface{})
	if !ok || len(gotMap) != 1 {
		t.Fatal("expected original response when transform is not configured")
	}
}
