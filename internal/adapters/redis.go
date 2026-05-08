package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Adapter interface {
	GetData(ctx context.Context, key string) (map[string]interface{}, error)
}

type RedisAdapter interface {
	Adapter
}

type redisAdapter struct {
	client *redis.Client
}

func NewRedisAdapter(endpoint, pass string) (RedisAdapter, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("redis endpoint is required")
	}

	return &redisAdapter{
		client: redis.NewClient(
			&redis.Options{
				Addr:         endpoint,
				Password:     pass,
				DB:           0,
				DialTimeout:  2 * time.Second,
				ReadTimeout:  2 * time.Second,
				WriteTimeout: 2 * time.Second,
			},
		),
	}, nil
}

func (r *redisAdapter) GetData(ctx context.Context, key string) (map[string]interface{}, error) {
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}
	return result, nil
}
