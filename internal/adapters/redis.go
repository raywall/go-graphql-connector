package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/raywall/go-graphql-integrator/internal/graph/models"
)

type RedisAdapter interface {
	GetConvenio(codigoConvenio int) (models.Convenio, error)
	GetLimiteOperacional(codigoConvenio int) (models.LimiteOperacional, error)
	GetTaxaFunding(codigoConvenio int) (models.TaxaFunding, error)
}

type redisAdapter struct {
	client *redis.Client
}

func NewRedisAdapter(endpoint, pass string) RedisAdapter {
	return &redisAdapter{
		client: redis.NewClient(
			&redis.Options{
				Addr:     endpoint,
				Password: pass,
				DB:       0,
			},
		),
	}
}

func (r *redisAdapter) GetConvenio(codigoConvenio int) (models.Convenio, error) {
	var convenio models.Convenio

	if codigoConvenio == 0 {
		return convenio, errors.New("codigoConvenio was not found")
	}

	data, err := r.client.Get(context.TODO(), fmt.Sprintf("CVN_%d", codigoConvenio)).Result()
	if err != nil {
		return convenio, err
	}

	if err := json.Unmarshal([]byte(data), &convenio); err != nil {
		return convenio, fmt.Errorf("error unmarshaling convenio: %v", err)
	}
	return convenio, nil
}

func (r *redisAdapter) GetLimiteOperacional(codigoConvenio int) (models.LimiteOperacional, error) {
	var limite models.LimiteOperacional

	if codigoConvenio == 0 {
		return limite, errors.New("codigoConvenio was not found")
	}

	data, err := r.client.Get(context.TODO(), fmt.Sprintf("LMT_%d", codigoConvenio)).Result()
	if err != nil {
		return limite, err
	}

	if err := json.Unmarshal([]byte(data), &limite); err != nil {
		return limite, fmt.Errorf("error unmarshaling limiteOperacional: %v", err)
	}
	return limite, nil
}

func (r *redisAdapter) GetTaxaFunding(codigoConvenio int) (models.TaxaFunding, error) {
	var funding models.TaxaFunding

	if codigoConvenio == 0 {
		return funding, errors.New("codigoConvenio was not found")
	}

	data, err := r.client.Get(context.TODO(), fmt.Sprintf("FDN_%d", codigoConvenio)).Result()
	if err != nil {
		return funding, err
	}

	if err := json.Unmarshal([]byte(data), &funding); err != nil {
		return funding, fmt.Errorf("error unmarshaling taxaFunding: %v", err)
	}
	return funding, nil
}
