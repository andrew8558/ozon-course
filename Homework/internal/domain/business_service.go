package domain

import (
	"Homework/internal/repository"
	"Homework/internal/repository/in_memory_cache"
	"Homework/internal/repository/postgresql"
	"Homework/internal/repository/redis"
	"context"
	"log"
	"strconv"
)

type BusinessService struct {
	Repo          *postgresql.PickupPointRepo
	InMemoryCache *in_memory_cache.InMemoryCache
	Redis         *redis.Redis
}

func (b *BusinessService) Add(ctx context.Context, pickupPoint repository.PickupPoint) (int64, error) {
	id, err := b.Repo.Add(ctx, pickupPoint)
	if err != nil {
		return 0, err
	}
	b.InMemoryCache.SetPickupPoint(id, pickupPoint)
	err = b.Redis.Set(ctx, strconv.Itoa(int(id)), pickupPoint)
	if err != nil {
		log.Println(err)
	}
	return id, nil
}

func (b *BusinessService) GetByID(ctx context.Context, id int64) (repository.PickupPoint, error) {
	pickupPoint, err := b.Redis.Get(ctx, strconv.Itoa(int(id)))
	if err == nil {
		return pickupPoint, nil
	} else {
		log.Println(err)
	}
	pickupPoint, err = b.Repo.GetByID(ctx, id)
	if err != nil {
		return repository.PickupPoint{}, err
	}

	err = b.Redis.Set(ctx, strconv.Itoa(int(id)), pickupPoint)
	if err != nil {
		return repository.PickupPoint{}, err
	}

	return pickupPoint, nil
}

func (b *BusinessService) Delete(ctx context.Context, id int64) error {
	b.InMemoryCache.DeletePickupPoint(id)
	return b.Repo.Delete(ctx, id)
}

func (b *BusinessService) List(ctx context.Context) ([]repository.PickupPoint, error) {
	pickupPoints, err := b.Repo.List(ctx)
	if err != nil {
		return []repository.PickupPoint{}, err
	}
	return pickupPoints, nil
}

func (b *BusinessService) Update(ctx context.Context, pickupPoint repository.PickupPoint) error {
	err := b.Repo.Update(ctx, pickupPoint)
	if err != nil {
		return err
	}
	b.InMemoryCache.SetPickupPoint(pickupPoint.ID, pickupPoint)
	return nil
}
