package utils

import (
	"12305/model"
	"context"
	"errors"
	"sync"
	"time"
)

// 本地cache       //map在并发情况下不安全，需要使用sync.Map
type Cache struct {
	data sync.Map
	mu   sync.RWMutex
}

type LocalCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, expiration time.Duration) error
	Del(ctx context.Context, key string) error
	Len(ctx context.Context, key string) (int, error)
}

func GetCache() *Cache {
	return &Cache{
		data: sync.Map{},
		mu:   sync.RWMutex{},
	}
}

func (c *Cache) Get(ctx context.Context, key string) ([]*model.Ticket, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.data.Load(key)
	if !ok {
		return nil, errors.New("key not found")
	}
	return value.([]*model.Ticket), nil
}

func (c *Cache) Set(ctx context.Context, key string, value []*model.Ticket, expiration time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data.Store(key, value)
	return nil
}

func (c *Cache) Del(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data.Delete(key)
	return nil
}

func (c *Cache) Len(ctx context.Context, key string) (int, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, ok := c.data.Load(key)
	if !ok {
		return 0, errors.New("key not found")
	}
	return len(value.([]*model.Ticket)), nil
}
