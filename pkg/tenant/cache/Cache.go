package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// Cache TODO
type Cache struct {
	c *cache.Cache
}

// NewCache 创建一个cache
func NewCache(defaultExpiration, cleanupInterval time.Duration) *cache.Cache {
	return cache.New(defaultExpiration, cleanupInterval)
}
