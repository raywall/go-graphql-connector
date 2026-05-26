package graph

import (
	"testing"

	"github.com/graphql-go/graphql"
)

func TestCreateSchemaResolvesConfiguredFieldWithMock(t *testing.T) {
	resolver, err := NewResolver(`{"connectors":[]}`)
	if err != nil {
		t.Fatalf("NewResolver returned error: %v", err)
	}
	if err := resolver.AddMockConfig(`{
		"status": true,
		"values": {
			"10341": {
				"convenio": {
					"codigoConvenio": 10341,
					"nomeConvenio": "Convenio Teste"
				}
			}
		}
	}`); err != nil {
		t.Fatalf("AddMockConfig returned error: %v", err)
	}

	schema, err := CreateSchema(resolver, `{
		"types": [
			{
				"name": "Convenio",
				"fields": [
					{"name": "codigoConvenio", "type": "Int"},
					{"name": "nomeConvenio", "type": "String"}
				]
			},
			{
				"name": "CombinedData",
				"fields": [
					{"name": "convenio", "type": "Object", "ofType": "Convenio"}
				]
			}
		],
		"query": {
			"name": "Query",
			"fields": [
				{
					"name": "dataSources",
					"type": "Object",
					"ofType": "CombinedData",
					"args": [
						{"name": "codigoConvenio", "type": "NonNull", "ofType": "Int"}
					]
				}
			]
		}
	}`)
	if err != nil {
		t.Fatalf("CreateSchema returned error: %v", err)
	}

	result := graphql.Do(graphql.Params{
		Schema:        *schema,
		RequestString: `{ dataSources(codigoConvenio: 10341) { convenio { nomeConvenio } } }`,
	})
	if len(result.Errors) > 0 {
		t.Fatalf("graphql returned errors: %v", result.Errors)
	}

	data := result.Data.(map[string]interface{})
	dataSources := data["dataSources"].(map[string]interface{})
	convenio := dataSources["convenio"].(map[string]interface{})
	if convenio["nomeConvenio"] != "Convenio Teste" {
		t.Fatalf("nomeConvenio = %v", convenio["nomeConvenio"])
	}
}

func TestCreateSchemaReturnsValidationError(t *testing.T) {
	resolver, err := NewResolver(`{"connectors":[]}`)
	if err != nil {
		t.Fatalf("NewResolver returned error: %v", err)
	}
	_, err = CreateSchema(resolver, `{
		"types": [{"name": "Broken", "fields": [{"name": "items", "type": "List"}]}],
		"query": {"name": "Query", "fields": []}
	}`)
	if err == nil {
		t.Fatal("expected schema validation error")
	}
}
