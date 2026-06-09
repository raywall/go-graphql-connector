package adapters

import "testing"

type nilTokenProvider struct{}

func (*nilTokenProvider) GetToken() (string, error) {
	return "token", nil
}

func TestRestAdapterIgnoresTypedNilTokenProvider(t *testing.T) {
	adapter, err := NewRestAdapter("https://example.com", "GET", nil, "", false)
	if err != nil {
		t.Fatalf("NewRestAdapter returned error: %v", err)
	}

	var provider *nilTokenProvider
	rest := adapter.(*RestAdapter)
	rest.SetTokenProvider(provider)

	if rest.tokenProvider != nil {
		t.Fatal("expected typed nil token provider to be ignored")
	}
}
