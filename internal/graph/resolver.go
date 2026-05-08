package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/raywall/go-graphql-connector/internal/graph/connectors"

	"github.com/raywall/cloud-easy-connector/pkg/cloud"
	"github.com/raywall/cloud-easy-connector/pkg/local"
)

type Resolver interface {
	ResolveDataSource(p graphql.ResolveParams) (interface{}, error)
	ResolveField(fieldName string, p graphql.ResolveParams) (interface{}, error)
	AddCloudContext(ctx cloud.CloudContext) error
	AddMockConfig(mockConfig string) error
}

type resolver struct {
	dataConnectors map[string]connectors.Connector
	cloudContext   cloud.CloudContext
	mock           *mockResolver
	allowPartial   bool
}

type mockResolver struct {
	Status bool                   `json:"status"`
	Values map[string]interface{} `json:"values"`
}

type ResolverOptions struct {
	AllowPartial bool
}

func NewResolver(connectorConfig string) (Resolver, error) {
	return NewResolverWithOptions(connectorConfig, ResolverOptions{})
}

func NewResolverWithOptions(connectorConfig string, options ResolverOptions) (Resolver, error) {
	connectors, err := connectors.LoadConnectors(connectorConfig)
	if err != nil {
		return nil, err
	}

	return &resolver{
		dataConnectors: connectors,
		allowPartial:   options.AllowPartial,
		mock: &mockResolver{
			Status: false,
			Values: nil,
		},
	}, nil
}

func (r *resolver) AddCloudContext(ctx cloud.CloudContext) error {
	r.cloudContext = ctx

	if ctx == nil {
		return nil
	}

	if jsonMock, err := ctx.GetParameterValue(local.New().GetEnvOrDefault("SSM_MOCK_VALUE", "/graphql/dev/mock"), false); err == nil && jsonMock != nil {
		return r.AddMockConfig(fmt.Sprintf("%v", jsonMock))
	}

	return nil
}

func (r *resolver) AddMockConfig(mockConfig string) error {
	if mockConfig == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(mockConfig), r.mock); err != nil {
		return fmt.Errorf("failed to deserialize mock value: \n\t%v", err)
	}
	return nil
}

func (r *resolver) ResolveDataSource(p graphql.ResolveParams) (interface{}, error) {
	return r.ResolveField("dataSources", p)
}

func (r *resolver) ResolveField(fieldName string, p graphql.ResolveParams) (interface{}, error) {
	_ = fieldName
	var (
		result          = make(map[string]interface{})
		requestedFields = getRequestedFields(p.Info)
		errChan         = make(chan error, len(requestedFields))
		wg              sync.WaitGroup
		mu              sync.Mutex
	)

	if len(requestedFields) == 0 {
		return result, nil
	}

	ctx := p.Context
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	args := normalizeArgs(p.Args)
	mockKey := primaryArgValue(args)

	for _, field := range requestedFields {
		if r.mock.Status {
			if data, ok := r.mockValue(mockKey, field); ok {
				result[field] = data
				continue
			}
		}

		conn, exists := r.dataConnectors[field]
		if !exists {
			errChan <- fmt.Errorf("no connector found for field: \n\t%s", field)
			continue
		}

		wg.Add(1)
		go func(field string, conn connectors.Connector) {
			defer wg.Done()
			start := time.Now()
			data, err := conn.GetData(ctx, args)
			if err != nil {
				if conn.Optional() || r.allowPartial {
					log.Printf("optional connector %s failed in %v: %v", field, time.Since(start), err)
					return
				}
				select {
				case errChan <- fmt.Errorf("error fetching %s: \n\t%w", field, err):
				case <-ctx.Done():
					return
				}
				return
			}
			mu.Lock()
			result[field] = data
			mu.Unlock()
			log.Printf("connector %s completed in %v", field, time.Since(start))
		}(field, conn)
	}

	wg.Wait()
	close(errChan)

	var combinedErr error
	for err := range errChan {
		combinedErr = errors.Join(combinedErr, err)
	}
	if combinedErr != nil {
		return nil, fmt.Errorf("error fetching data: \n\t%w", combinedErr)
	}

	return result, nil
}

func (r *resolver) mockValue(mockKey, field string) (interface{}, bool) {
	if r.mock == nil || r.mock.Values == nil || mockKey == "" {
		return nil, false
	}

	if valuesByKey, ok := r.mock.Values[mockKey].(map[string]interface{}); ok {
		value, ok := valuesByKey[field]
		return value, ok
	}
	if valuesByField, ok := r.mock.Values[field].(map[string]interface{}); ok {
		value, ok := valuesByField[mockKey]
		return value, ok
	}

	return nil, false
}

func normalizeArgs(args map[string]interface{}) map[string]interface{} {
	normalized := make(map[string]interface{}, len(args))
	for key, value := range args {
		normalized[key] = value
	}
	return normalized
}

func primaryArgValue(args map[string]interface{}) string {
	if value, ok := args["codigoConvenio"]; ok {
		return fmt.Sprintf("%v", value)
	}
	if len(args) == 0 {
		return ""
	}
	keys := make([]string, 0, len(args))
	for key := range args {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return fmt.Sprintf("%v", args[keys[0]])
}

func getRequestedFields(info graphql.ResolveInfo) []string {
	fields := make([]string, 0)

	if len(info.FieldASTs) == 0 {
		return fields
	}

	if info.FieldASTs[0].SelectionSet == nil {
		return fields
	}

	for _, selection := range info.FieldASTs[0].SelectionSet.Selections {
		switch sel := selection.(type) {
		case *ast.Field:
			fields = append(fields, sel.Name.Value)
		case *ast.InlineFragment:
			fields = append(fields, fieldsFromSelectionSet(sel.SelectionSet)...)
		case *ast.FragmentSpread:
			if fragment, ok := info.Fragments[sel.Name.Value]; ok {
				fields = append(fields, fieldsFromSelectionSet(fragment.GetSelectionSet())...)
			}
		}
	}

	return fields
}

func fieldsFromSelectionSet(selectionSet *ast.SelectionSet) []string {
	if selectionSet == nil {
		return nil
	}
	fields := make([]string, 0, len(selectionSet.Selections))
	for _, selection := range selectionSet.Selections {
		switch sel := selection.(type) {
		case *ast.Field:
			fields = append(fields, sel.Name.Value)
		case *ast.InlineFragment:
			fields = append(fields, fieldsFromSelectionSet(sel.SelectionSet)...)
		}
	}
	return fields
}
