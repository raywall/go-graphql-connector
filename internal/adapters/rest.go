package adapters

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type RestAdapter struct {
	client   *http.Client
	baseUrl  string
	endpoint string
}

func NewRestAdapter(baseUrl, endpoint string) Adapter {
	return &RestAdapter{
		client:   &http.Client{},
		baseUrl:  baseUrl,
		endpoint: endpoint,
	}
}

func (r *RestAdapter) GetData(key string) (map[string]interface{}, error) {
	url := r.baseUrl + strings.Replace(r.endpoint, "{codigoConvenio}", key, 1)
	resp, err := r.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from REST API %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("REST API returned status %d for %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read REST API response: %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to decode REST API response: %v", err)
	}
	return data, nil
}
