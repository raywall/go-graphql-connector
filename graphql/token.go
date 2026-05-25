package graphql

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type managedToken struct {
	url          string
	clientID     string
	clientSecret string
	headers      map[string]string
	client       *http.Client

	mu        sync.Mutex
	token     string
	expiresAt time.Time
}

func newManagedToken(tokenURL, clientID, clientSecret string, headers map[string]string) *managedToken {
	return &managedToken{
		url:          tokenURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		headers:      headers,
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSNextProto: map[string]func(string, *tls.Conn) http.RoundTripper{},
			},
		},
	}
}

func (t *managedToken) GetToken() (string, error) {
	if t == nil {
		return "", fmt.Errorf("token manager is not configured")
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.token != "" && time.Now().Add(30*time.Second).Before(t.expiresAt) {
		return t.token, nil
	}
	return t.refresh(context.Background())
}

func (t *managedToken) refresh(ctx context.Context) (string, error) {
	form := "grant_type=client_credentials" +
		"&client_id=" + url.QueryEscape(t.clientID) +
		"&client_secret=" + url.QueryEscape(t.clientSecret)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, t.url, strings.NewReader(form))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for name, value := range t.headers {
		req.Header[name] = []string{value}
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("token service returned status %s", resp.Status)
	}

	var body struct {
		AccessToken string `json:"access_token"`
		Token       string `json:"token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", err
	}

	token := body.AccessToken
	if token == "" {
		token = body.Token
	}
	if token == "" {
		return "", fmt.Errorf("token service response does not contain access_token")
	}
	expiresIn := body.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 300
	}

	t.token = token
	t.expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	return t.token, nil
}
