package cache

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto"
	"github.com/miekg/dns"
	"github.com/thedeepak12/hermes-dns/internal/config"
)

type Cache struct {
	cache *ristretto.Cache
	ttl   time.Duration
}

func New(cfg *config.CacheConfig) (*Cache, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(cfg.MaxSize * 10),
		MaxCost:     int64(cfg.MaxSize),
		BufferItems: 64,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to create cache: %w", err)
	}

	return &Cache{
		cache: cache,
		ttl:   time.Duration(cfg.DefaultTTL) * time.Second,
	}, nil
}

func (c *Cache) Get(key string) (*dns.Msg, bool) {
	if c == nil {
		return nil, false
	}

	value, found := c.cache.Get(key)
	if !found {
		return nil, false
	}

	msg, ok := value.(*dns.Msg)
	if !ok {
		return nil, false
	}

	return msg, true
}

func (c *Cache) Set(key string, msg *dns.Msg) {
	if c == nil {
		return
	}

	ttl := c.ttl
	if len(msg.Answer) > 0 {
		ttl = time.Duration(msg.Answer[0].Header().Ttl) * time.Second
		if ttl == 0 {
			ttl = c.ttl
		}
	}

	c.cache.SetWithTTL(key, msg, 1, ttl)
	c.cache.Wait()
}

func (c *Cache) Close() {
	if c != nil {
		c.cache.Close()
	}
}
