package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/raywall/go-graphql-integrator/internal/graph/connectors"

	"github.com/raywall/cloud-easy-connector/pkg/cloud"
	"github.com/raywall/cloud-easy-connector/pkg/local"
)

type Resolver interface {
	ResolveDataSource(p graphql.ResolveParams) (interface{}, error)
	AddCloudContext(ctx cloud.CloudContext) error
}

type resolver struct {
	dataConnectors map[string]connectors.Connector
	cloudContext   cloud.CloudContext
	mock           *mockResolver
	// apiClient      *adapters.APIClient
}

type mockResolver struct {
	Status bool                   `json:"status"`
	Values map[string]interface{} `json:"values"`
}

func NewResolver(connectorConfig string) (Resolver, error) {
	connectors, err := connectors.LoadConnectors(connectorConfig)
	if err != nil {
		return nil, err
	}

	return &resolver{
		dataConnectors: connectors,
		mock: &mockResolver{
			Status: false,
			Values: nil,
		},
		// apiClient:      adapters.NewAPIClient(),
	}, nil
}

func (r *resolver) AddCloudContext(ctx cloud.CloudContext) error {
	r.cloudContext = ctx

	if jsonMock, err := ctx.GetParameterValue(local.New().GetEnvOrDefault("SSM_MOCK_VALUE", "/graphql/dev/mock"), false); err != nil && jsonMock != nil {
		if err := json.Unmarshal([]byte(jsonMock.(string)), r.mock); err != nil {
			return fmt.Errorf("failed to deserialize mock value from ssm: \n\t%v", err)
		}
	}

	return nil
}

func (r *resolver) ResolveDataSource(p graphql.ResolveParams) (interface{}, error) {
	codigo, ok := p.Args["codigoConvenio"].(int)
	if !ok {
		return nil, errors.New("invalid codigoConvenio")
	}

	var (
		result          = make(map[string]interface{})
		requestedFields = getRequestedFields(p.Info)
		errChan         = make(chan error, len(requestedFields))
		wg              sync.WaitGroup
	)

	// Context for timeout/cancellation control
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, field := range requestedFields {
		conn, exists := r.dataConnectors[field]
		if !exists {
			errChan <- fmt.Errorf("no connector found for field: \n\t%s", field)
			continue
		}

		if r.mock.Status {
			if values, ok := r.mock.Values[field]; ok {
				result[field] = (values.(map[string]interface{}))[fmt.Sprintf("%d", codigo)]
			}
			continue
		}

		wg.Add(1)
		go func(field string, conn connectors.Connector) {
			defer wg.Done()
			data, err := conn.GetData(codigo)
			if err != nil {
				select {
				case errChan <- fmt.Errorf("error fetching %s: \n\t%w", field, err):
				case <-ctx.Done():
					return
				}
				return
			}
			result[field] = data
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
		}
	}

	return fields
}
