package graphql

import (
	"fmt"

	"github.com/raywall/cloud-easy-connector/pkg/cloud"
)

// Credentials represents the credentials that will be used to generate a new token
type Credentials struct {
	// ClientID indicates the registered client application id
	ClientID string `json:"client_id"`

	// ClientSecret indicates the password of the registered client application
	ClientSecret string `json:"client_secret"`
}

// Tokenservice contains the settings needed to generate a token STS
type TokenService struct {
	// TokenAuthorizationURL represents the URL for the Token service
	TokenAuthorizationURL string `json:"token_authorization_url"`

	// Credentials representa as credenciais que serão utilizadas para gerar um novo token
	Credentials Credentials
}

// Authorization contains the authorization settings to be used by the Graphql API connectors
type Authorization struct {
	// RequireTokenSTS indicates whether your API will use an STS token to generate authentication tokens
	// for API connectors
	RequireTokenSTS bool `json:"require_token_sts"`

	// TokenService contains the settings required to generate an STS token
	TokenService TokenService
}

// Config contains all the configuration required to create and instantiate a dynamic GraphQL API
type Config struct {
	// Schema is the content or path to retrieve the schema settings of the GraphQL API that will be
	// created dynamically
	Schema string `json:"schema"`

	// Connectors is the content or path to retrieve settings from the GraphQL API resolver that will
	// be created dynamically
	Connectors string `json:"connectors"`

	// Route represents the route that will be used by the GraphQL API (e.g. /graphql)
	Route string `json:"route"`

	// Authorization contains the authorization settings to be used by GraphQL API connectors
	Authorization Authorization

	// CloudContext is the cloud context that will be used to interact with available cloud resources
	CloudContext cloud.CloudContext
}

// GetSchemaValue is the method responsible for retrieving the schema settings from the GraphQL API
func (c *Config) GetSchemaValue() (string, error) {
	if c.Schema == "" {
		return "", fmt.Errorf("it's necessary to specify the GraphQL API schema")
	}

	if IsPath(c.Schema) {
		value, err := FromString(c.Schema).GetValue(c.CloudContext)
		if err != nil {
			return "", fmt.Errorf("failed to get the schema value: %v", err)
		}
		return value.(string), nil
	}

	return c.Schema, nil
}

// GetConnectorsValue is the method responsible for retrieving the configurations of the GraphQL API connectors
func (c *Config) GetConnectorsValue() (string, error) {
	if c.Connectors == "" {
		return "", fmt.Errorf("it's necessary to specify the GraphQL API connections")
	}

	if IsPath(c.Connectors) {
		value, err := FromString(c.Connectors).GetValue(c.CloudContext)
		if err != nil {
			return "", fmt.Errorf("failed to get the connectors value: %v", err)
		}
		return value.(string), nil
	}

	return c.Connectors, nil
}

func (c *Config) GetTokenServiceURL() (string, error) {
	authService := c.Authorization.TokenService.TokenAuthorizationURL
	if IsPath(authService) {
		value, err := FromString(authService).GetValue(c.CloudContext)
		if err != nil {
			return "", fmt.Errorf("failed to get the authorization service value: %v", err)
		}
		authService = value.(string)
	}

	return authService, nil
}

func (c *Config) GetCredentials() (string, string, error) {
	clientID := c.Authorization.TokenService.Credentials.ClientID
	if IsPath(clientID) {
		value, err := FromString(clientID).GetValue(c.CloudContext)
		if err != nil {
			return "", "", fmt.Errorf("failed to get the authorization client id: %v", err)
		}
		if id, ok := value.(string); ok {
			clientID = id
		}
		if obj, ok := value.(map[string]interface{}); ok {
			clientID = obj["client_id"].(string)
		}
	}

	clientSecret := c.Authorization.TokenService.Credentials.ClientSecret
	if IsPath(clientSecret) {
		value, err := FromString(clientSecret).GetValue(c.CloudContext)
		if err != nil {
			return "", "", fmt.Errorf("failed to get the authorization client secret: %v", err)
		}
		if id, ok := value.(string); ok {
			clientSecret = id
		}
		if obj, ok := value.(map[string]interface{}); ok {
			clientSecret = obj["client_secret"].(string)
		}
	}

	return clientID, clientSecret, nil
}
