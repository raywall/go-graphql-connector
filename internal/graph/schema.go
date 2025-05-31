package graph

import (
	"encoding/json"
	"log"

	"github.com/graphql-go/graphql"
)

type FieldConfig struct {
	Name   string      `json:"name"`
	Type   string      `json:"type"`
	OfType string      `json:"ofType,omitempty"`
	Args   []ArgConfig `json:"args,omitempty"`
}

type TypeConfig struct {
	Name   string        `json:"name"`
	Fields []FieldConfig `json:"fields"`
}

type ArgConfig struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	OfType string `json:"ofType,omitempty"`
}

type QueryConfig struct {
	Name   string        `json:"name"`
	Fields []FieldConfig `json:"fields"`
}

type SchemaConfig struct {
	Types []TypeConfig `json:"types"`
	Query QueryConfig  `json:"query"`
}

var typeMap map[string]*graphql.Object

func getGraphQLType(field FieldConfig, typeMap map[string]*graphql.Object) graphql.Output {
	switch field.Type {
	case "Int":
		return graphql.Int
	case "Boolean":
		return graphql.Boolean
	case "Float":
		return graphql.Float
	case "String":
		return graphql.String
	case "List":
		if field.OfType == "" {
			log.Fatalf("List type must specify ofType")
		}
		if basicType, exists := map[string]graphql.Type{
			"Int":    graphql.Int,
			"String": graphql.String,
		}[field.OfType]; exists {
			return graphql.NewList(basicType)
		}
		if objType, exists := typeMap[field.OfType]; exists {
			return graphql.NewList(objType)
		}
	case "Object":
		if field.OfType == "" {
			log.Fatalf("Object type must specify ofType")
		}
		if objType, exists := typeMap[field.OfType]; exists {
			return objType
		}
	case "NonNull":
		if field.OfType == "" {
			log.Fatalf("NonNull type must specify ofType")
		}
		if basicType, exists := map[string]graphql.Type{
			"Int":    graphql.Int,
			"String": graphql.String,
		}[field.OfType]; exists {
			return graphql.NewNonNull(basicType)
		}
		if objType, exists := typeMap[field.OfType]; exists {
			return graphql.NewNonNull(objType)
		}
	}
	log.Fatalf("Unknown type: %s with ofType: %s", field.Type, field.OfType)
	return nil
}

func CreateSchema(res Resolver, schemaConfig string) (*graphql.Schema, error) {
	var config SchemaConfig
	if err := json.Unmarshal([]byte(schemaConfig), &config); err != nil {
		return nil, err
	}

	// Inicializar o mapa de tipos
	typeMap = make(map[string]*graphql.Object)

	// Criar tipos GraphQL
	for _, typeDef := range config.Types {
		fields := graphql.Fields{}
		for _, field := range typeDef.Fields {
			fields[field.Name] = &graphql.Field{
				Type: getGraphQLType(field, typeMap),
			}
		}
		typeMap[typeDef.Name] = graphql.NewObject(graphql.ObjectConfig{
			Name:   typeDef.Name,
			Fields: fields,
		})
	}

	// Resolver dependências de tipos
	for _, typeDef := range config.Types {
		obj := typeMap[typeDef.Name]
		for _, field := range typeDef.Fields {
			obj.AddFieldConfig(field.Name, &graphql.Field{
				Type: getGraphQLType(field, typeMap),
			})
		}
	}

	// Criar o tipo Query
	queryFields := graphql.Fields{}
	for _, field := range config.Query.Fields {
		args := graphql.FieldConfigArgument{}
		for _, arg := range field.Args {
			args[arg.Name] = &graphql.ArgumentConfig{
				Type: getGraphQLType(FieldConfig{Type: arg.Type, OfType: arg.OfType}, typeMap),
			}
		}
		queryFields[field.Name] = &graphql.Field{
			Type: getGraphQLType(field, typeMap),
			Args: args,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				// Mapear o campo para o resolver correspondente
				if field.Name == "dataSources" {
					return res.ResolveDataSource(p)
				}
				return nil, nil
			},
		}
	}

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:   config.Query.Name,
		Fields: queryFields,
	})

	// Criar o schema
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
	if err != nil {
		return nil, err
	}

	return &schema, nil
}
