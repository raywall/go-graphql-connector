package adapters

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

type Adapter interface {
	GetData(key string) (map[string]interface{}, error)
}

type RedisAdapter interface {
	Adapter
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

func (r *redisAdapter) GetData(key string) (map[string]interface{}, error) {
	data, err := r.client.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}
	return result, nil
}
