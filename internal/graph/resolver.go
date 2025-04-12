package graph

import (
	"errors"
	"fmt"
	"sync"

	"github.com/graphql-go/graphql"
	"github.com/raywall/go-graphql-integrator/internal/adapters"
	"github.com/raywall/go-graphql-integrator/internal/graph/models"
)

type Resolver interface {
	ResolveDataSource(p graphql.ResolveParams) (interface{}, error)
}

type resolver struct {
	redisAdapter adapters.RedisAdapter
	apiClient    *adapters.APIClient
}

type Result struct {
	Convenio            models.Convenio            `json:"convenio"`
	LimiteOperacional   models.LimiteOperacional   `json:"limiteOperacional"`
	TaxaFunding         models.TaxaFunding         `json:"taxaFunding"`
	ColaboradorConvenio models.ColaboradorConvenio `json:"colaboradorConvenio"`
}

func NewResolver(endpoint, pass string) Resolver {
	return &resolver{
		redisAdapter: adapters.NewRedisAdapter(endpoint, pass),
		apiClient:    adapters.NewAPIClient(),
	}
}

func (r *resolver) ResolveDataSource(p graphql.ResolveParams) (interface{}, error) {
	codigo, ok := p.Args["codigoConvenio"].(int)
	if !ok {
		return nil, errors.New("invalid codigoConvenio")
	}

	var (
		conv models.Convenio
		lim  models.LimiteOperacional
		tax  models.TaxaFunding
		col  models.ColaboradorConvenio

		errConv, errLim, errTax, errCol error
		wg                              sync.WaitGroup
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		conv, errConv = r.redisAdapter.GetConvenio(codigo)
	}()

	go func() {
		defer wg.Done()
		lim, errLim = r.redisAdapter.GetLimiteOperacional(codigo)
	}()

	go func() {
		defer wg.Done()
		tax, errTax = r.redisAdapter.GetTaxaFunding(codigo)
	}()

	// go func() {
	// 	defer wg.Done()
	// 	col, errCol = r.apiClient.GetColaboradorConvenio(codigo)
	// }()

	wg.Wait()

	if err := errors.Join(errConv, errLim, errTax, errCol); err != nil {
		return nil, fmt.Errorf("error fetching data: %w", err)
	}

	result := Result{
		Convenio:            conv,
		LimiteOperacional:   lim,
		TaxaFunding:         tax,
		ColaboradorConvenio: col,
	}

	return result, nil
}
