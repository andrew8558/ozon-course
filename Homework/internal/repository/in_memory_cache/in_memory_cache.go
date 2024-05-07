package in_memory_cache

import (
	"Homework/internal/repository"
	"errors"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"sync"
	"time"
)

var (
	pickupPointNotFound = errors.New("cant find pickup point by id")
)

type InMemoryCache struct {
	cache *expirable.LRU[int64, repository.PickupPoint]
	mx    sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	newCache := expirable.NewLRU[int64, repository.PickupPoint](100, nil, 10*time.Millisecond)
	inMemoryCache := &InMemoryCache{
		cache: newCache,
		mx:    sync.RWMutex{},
	}

	return inMemoryCache
}

func (c *InMemoryCache) SetPickupPoint(id int64, pickupPoint repository.PickupPoint) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.cache.Add(id, pickupPoint)
}

func (c *InMemoryCache) GetPickupPoint(id int64) (repository.PickupPoint, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()
	pickupPoint, ok := c.cache.Get(id)
	if !ok {
		return repository.PickupPoint{}, pickupPointNotFound
	}
	return pickupPoint, nil
}

func (c *InMemoryCache) DeletePickupPoint(id int64) {
	c.mx.Lock()
	defer c.mx.Unlock()
	c.cache.Remove(id)
}
