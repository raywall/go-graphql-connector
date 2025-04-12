package adapters

import (
	"time"

	"github.com/raywall/go-graphql-integrator/internal/graph/models"
)

type APIClient struct {
	// Add actual HTTP client configuration here
}

func NewAPIClient() *APIClient {
	return &APIClient{}
}

func (c *APIClient) GetColaboradorConvenio(codigoConvenio int) (*models.ColaboradorConvenio, error) {
	time.Sleep(60 * time.Millisecond)
	return &models.ColaboradorConvenio{
		CodigoIdentificacaoPessoa: "COL_123",
		// ... other fields
	}, nil
}
