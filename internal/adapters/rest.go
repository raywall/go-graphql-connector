package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"time"

	"github.com/raywall/go-graphql-connector/internal/tracing"
)

type RestAdapter struct {
	client        *http.Client
	baseURL       string
	method        string
	headers       map[string]string
	body          string
	tokenProvider TokenProvider
}

type TokenProvider interface {
	GetToken() (string, error)
}

func NewRestAdapter(baseURL, method string, headers map[string]string, body string) (Adapter, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("rest baseUrl is required")
	}
	if method == "" {
		method = http.MethodGet
	}

	return &RestAdapter{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: baseURL,
		method:  method,
		headers: headers,
		body:    body,
	}, nil
}

func (r *RestAdapter) SetTokenProvider(provider TokenProvider) {
	if isNilTokenProvider(provider) {
		return
	}
	r.tokenProvider = provider
}

func isNilTokenProvider(provider TokenProvider) bool {
	if provider == nil {
		return true
	}
	value := reflect.ValueOf(provider)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (r *RestAdapter) GetData(ctx context.Context, key string) (map[string]interface{}, error) {
	url := r.baseURL + key
	var body io.Reader
	if r.body != "" {
		body = bytes.NewBufferString(r.body)
	}

	req, err := http.NewRequestWithContext(ctx, r.method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create REST request %s: %v", url, err)
	}
	for name, value := range r.headers {
		req.Header.Set(name, value)
	}
	tracing.Inject(req)
	if r.tokenProvider != nil && req.Header.Get("Authorization") == "" {
		token, err := r.tokenProvider.GetToken()
		if err != nil {
			return nil, fmt.Errorf("failed to get STS token for REST API %s: %v", url, err)
		}
		log.Printf("Using token from provider for REST API %s: %s", url, maskSecret(token))
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from REST API %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("REST API returned status %d for %s", resp.StatusCode, url)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read REST API response: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(responseBody, &data); err != nil {
		return nil, fmt.Errorf("failed to decode REST API response: %v", err)
	}
	return data, nil
}

func maskSecret(value string) string {
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-4:]
}
