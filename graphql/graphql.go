package graphql

import (
	"fmt"

	gp "github.com/graphql-go/graphql"
	"github.com/raywall/cloud-easy-connector/pkg/cloud"
	"github.com/raywall/go-graphql-connector/internal/adapters"
	"github.com/raywall/go-graphql-connector/internal/graph"
	"github.com/raywall/go-graphql-connector/internal/graph/connectors"
)

// route is the API route name that will be used by default
const route string = "/graphql"

type GraphQL struct {
	tokenManager *managedToken   `json:"-"`
	AccessToken  *string         `json:"token"`
	Config       Config          `json:"config"`
	Resolver     *graph.Resolver `json:"resolver"`
	Schema       *gp.Schema      `json:"schema"`
}

func New(config *Config, resources *cloud.CloudContextList, region, endpoint string) (*GraphQL, error) {
	var (
		err error
		api = GraphQL{}
	)

	// route
	if config.Route == "" {
		config.Route = route
	}
	if !config.Pretty {
		config.Pretty = true
	}
	api.Config = *config

	// cloud context
	if region == "" {
		return nil, fmt.Errorf("it's necessary to inform the AWS region you want to use")
	}

	config.CloudContext, err = cloud.NewAwsCloudContext(region, endpoint, resources)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new AWS Cloud Context: %v", err)
	}

	// token
	if config.Authorization.RequireTokenSTS {
		// auth_service_url
		authServiceUrl, err := config.GetTokenServiceURL()
		if err != nil {
			return nil, err
		}

		// credentials
		clientID, clientSecret, err := config.GetCredentials()
		if err != nil {
			return nil, err
		}

		api.tokenManager = newManagedToken(
			authServiceUrl,
			clientID,
			clientSecret,
			config.GetTokenServiceHeaders(),
			config.Authorization.SkipTLSVerify)

		if token, err := api.tokenManager.GetToken(); err != nil {
			return nil, fmt.Errorf("failed to start STS token manager: %v", err)
		} else {
			api.AccessToken = &token
		}
	}

	// connections
	connectionsConfig, err := config.GetConnectorsValue()
	if err != nil {
		return nil, fmt.Errorf("failed to get the connections config: %v", err)
	}

	var tokenProvider adapters.TokenProvider
	if api.tokenManager != nil {
		tokenProvider = api.tokenManager
	}

	res, err := graph.NewResolverWithOptions(connectionsConfig, graph.ResolverOptions{
		AllowPartial:  config.AllowPartial,
		TokenProvider: tokenProvider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create a resolver: %v", err)
	}
	if err := res.AddCloudContext(config.CloudContext); err != nil {
		return nil, fmt.Errorf("failed to add cloud context to resolver: %v", err)
	}
	if mockConfig, err := config.GetMockValue(); err != nil {
		return nil, err
	} else if mockConfig != "" {
		if err := res.AddMockConfig(mockConfig); err != nil {
			return nil, err
		}
	}
	api.Resolver = &res

	// schema
	schemaConfig, err := config.GetSchemaValue()
	if err != nil {
		return nil, fmt.Errorf("failed to get the schema config: %v", err)
	}

	api.Schema, err = graph.CreateSchema(res, schemaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create a schema: %v", err)
	}

	return &api, nil
}

// ConnectorCircuitStates returns the current diagnostic state of each connector
// circuit breaker.
func (g *GraphQL) ConnectorCircuitStates() map[string]connectors.CircuitState {
	if g == nil || g.Resolver == nil {
		return map[string]connectors.CircuitState{}
	}
	return (*g.Resolver).ConnectorCircuitStates()
}
