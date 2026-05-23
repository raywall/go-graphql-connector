package graph

import (
	"encoding/json"
	"fmt"

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

func getGraphQLOutputType(field FieldConfig, typeMap map[string]*graphql.Object) (graphql.Output, error) {
	basicTypes := map[string]graphql.Type{
		"Int":     graphql.Int,
		"Boolean": graphql.Boolean,
		"Float":   graphql.Float,
		"String":  graphql.String,
	}
	switch field.Type {
	case "Int":
		return graphql.Int, nil
	case "Boolean":
		return graphql.Boolean, nil
	case "Float":
		return graphql.Float, nil
	case "String":
		return graphql.String, nil
	case "List":
		if field.OfType == "" {
			return nil, fmt.Errorf("list type must specify ofType")
		}
		if basicType, exists := basicTypes[field.OfType]; exists {
			return graphql.NewList(basicType), nil
		}
		if objType, exists := typeMap[field.OfType]; exists {
			return graphql.NewList(objType), nil
		}
	case "Object":
		if field.OfType == "" {
			return nil, fmt.Errorf("object type must specify ofType")
		}
		if objType, exists := typeMap[field.OfType]; exists {
			return objType, nil
		}
	case "NonNull":
		if field.OfType == "" {
			return nil, fmt.Errorf("nonnull type must specify ofType")
		}
		if basicType, exists := basicTypes[field.OfType]; exists {
			return graphql.NewNonNull(basicType), nil
		}
		if objType, exists := typeMap[field.OfType]; exists {
			return graphql.NewNonNull(objType), nil
		}
	}
	return nil, fmt.Errorf("unknown type: %s with ofType: %s", field.Type, field.OfType)
}

func getGraphQLInputType(arg ArgConfig) (graphql.Input, error) {
	switch arg.Type {
	case "Int":
		return graphql.Int, nil
	case "Boolean":
		return graphql.Boolean, nil
	case "Float":
		return graphql.Float, nil
	case "String":
		return graphql.String, nil
	case "NonNull":
		switch arg.OfType {
		case "Int":
			return graphql.NewNonNull(graphql.Int), nil
		case "Boolean":
			return graphql.NewNonNull(graphql.Boolean), nil
		case "Float":
			return graphql.NewNonNull(graphql.Float), nil
		case "String":
			return graphql.NewNonNull(graphql.String), nil
		default:
			return nil, fmt.Errorf("unsupported non-null input type: %s", arg.OfType)
		}
	default:
		return nil, fmt.Errorf("unsupported input type: %s", arg.Type)
	}
}

func CreateSchema(res Resolver, schemaConfig string) (*graphql.Schema, error) {
	var config SchemaConfig
	if err := json.Unmarshal([]byte(schemaConfig), &config); err != nil {
		return nil, err
	}

	typeMap := make(map[string]*graphql.Object)

	for _, typeDef := range config.Types {
		if typeDef.Name == "" {
			return nil, fmt.Errorf("schema type name is required")
		}
		typeMap[typeDef.Name] = graphql.NewObject(graphql.ObjectConfig{
			Name:   typeDef.Name,
			Fields: graphql.Fields{},
		})
	}

	for _, typeDef := range config.Types {
		obj := typeMap[typeDef.Name]
		for _, field := range typeDef.Fields {
			fieldType, err := getGraphQLOutputType(field, typeMap)
			if err != nil {
				return nil, fmt.Errorf("invalid field %s.%s: %w", typeDef.Name, field.Name, err)
			}
			obj.AddFieldConfig(field.Name, &graphql.Field{
				Type: fieldType,
			})
		}
	}

	queryFields := graphql.Fields{}
	for _, field := range config.Query.Fields {
		args := graphql.FieldConfigArgument{}
		for _, arg := range field.Args {
			argType, err := getGraphQLInputType(arg)
			if err != nil {
				return nil, fmt.Errorf("invalid argument %s.%s: %w", field.Name, arg.Name, err)
			}
			args[arg.Name] = &graphql.ArgumentConfig{
				Type: argType,
			}
		}
		fieldType, err := getGraphQLOutputType(field, typeMap)
		if err != nil {
			return nil, fmt.Errorf("invalid query field %s: %w", field.Name, err)
		}
		queryField := field
		queryFields[field.Name] = &graphql.Field{
			Type: fieldType,
			Args: args,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return res.ResolveField(queryField.Name, p)
			},
		}
	}

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name:   config.Query.Name,
		Fields: queryFields,
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: queryType,
	})
	if err != nil {
		return nil, err
	}

	return &schema, nil
}
