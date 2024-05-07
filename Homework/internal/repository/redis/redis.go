package redis

import (
	"Homework/internal/repository"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(opt *redis.Options) *Redis {
	return &Redis{
		redis.NewClient(opt),
	}
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}) error {
	return r.client.Set(ctx, key, value, time.Minute*10).Err()
}

func (r *Redis) Get(ctx context.Context, key string) (repository.PickupPoint, error) {
	res := r.client.Get(ctx, key)
	if res.Err() != nil {
		return repository.PickupPoint{}, res.Err()
	}

	var pickupPoint repository.PickupPoint
	err := res.Scan(&pickupPoint)
	if err != nil {
		return repository.PickupPoint{}, err
	}

	return pickupPoint, nil
}
