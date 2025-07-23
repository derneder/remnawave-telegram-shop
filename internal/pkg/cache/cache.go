package cache

import (
	"context"
	"sync"
	"time"
)

type Item struct {
	Value     int
	ExpiresAt time.Time
}

type Cache struct {
	data   map[int64]Item
	mutex  sync.RWMutex
	ttl    time.Duration
	cancel context.CancelFunc
}

func NewCache(ctx context.Context, ttl time.Duration) *Cache {
	ctx, cancel := context.WithCancel(ctx)
	c := &Cache{
		data:   make(map[int64]Item),
		ttl:    ttl,
		cancel: cancel,
	}
	go c.cleanupExpired(ctx)
	return c
}

func (c *Cache) Set(key int64, value int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[key] = Item{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache) Get(key int64) (int, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	item, found := c.data[key]
	if !found || time.Now().After(item.ExpiresAt) {
		return 0, false
	}
	return item.Value, true
}

func (c *Cache) Delete(key int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

// Close stops background cleanup goroutine.
func (c *Cache) Close() {
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Cache) cleanupExpired(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mutex.Lock()
			for k, v := range c.data {
				if now.After(v.ExpiresAt) {
					delete(c.data, k)
				}
			}
			c.mutex.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
