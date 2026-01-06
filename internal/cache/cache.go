package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stripsior/loc.api/internal/models"
)

type Cache interface {
	Get(key string) (*models.AnalyzeResponse, bool)
	Set(key string, result *models.AnalyzeResponse)
	Size() int
	Clear()
}

type CacheEntry struct {
	Result    *models.AnalyzeResponse
	ExpiresAt time.Time
}

type MemoryCache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

func NewMemoryCache(ttl time.Duration) *MemoryCache {
	cache := &MemoryCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}

	go cache.cleanupExpired()

	return cache
}

func (c *MemoryCache) Get(key string) (*models.AnalyzeResponse, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Result, true
}

func (c *MemoryCache) Set(key string, result *models.AnalyzeResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

func (c *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.ExpiresAt) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

const redisPrefix = "loc:cache:"

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
	ctx    context.Context
}

func NewRedisCache(url string, ttl time.Duration) (*RedisCache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)
	ctx := context.Background()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		ttl:    ttl,
		ctx:    ctx,
	}, nil
}

func (c *RedisCache) Get(key string) (*models.AnalyzeResponse, bool) {
	val, err := c.client.Get(c.ctx, redisPrefix+key).Bytes()
	if err != nil {
		return nil, false
	}

	var result models.AnalyzeResponse
	if err := json.Unmarshal(val, &result); err != nil {
		return nil, false
	}

	return &result, true
}

func (c *RedisCache) Set(key string, result *models.AnalyzeResponse) {
	val, err := json.Marshal(result)
	if err != nil {
		return
	}

	c.client.Set(c.ctx, redisPrefix+key, val, c.ttl)
}

func (c *RedisCache) Clear() {
	iter := c.client.Scan(c.ctx, 0, redisPrefix+"*", 0).Iterator()
	for iter.Next(c.ctx) {
		c.client.Del(c.ctx, iter.Val())
	}
}

func (c *RedisCache) Size() int {
	count := 0
	iter := c.client.Scan(c.ctx, 0, redisPrefix+"*", 0).Iterator()
	for iter.Next(c.ctx) {
		count++
	}
	return count
}

func GenerateCacheKey(owner, repo, branch string, includeAuthors bool, filters *models.FilterOptions) string {
	// Normalize filters to ensure consistent cache keys
	var normalizedFilters *models.FilterOptions
	if filters != nil {
		f := *filters
		// Sort slices
		if len(f.ExcludeExtensions) > 0 {
			sort.Strings(f.ExcludeExtensions)
		}
		if len(f.IncludeExtensions) > 0 {
			sort.Strings(f.IncludeExtensions)
		}
		if len(f.ExcludeDirectories) > 0 {
			sort.Strings(f.ExcludeDirectories)
		}
		normalizedFilters = &f
	}

	data := struct {
		Owner          string
		Repo           string
		Branch         string
		IncludeAuthors bool
		Filters        *models.FilterOptions
	}{
		Owner:          owner,
		Repo:           repo,
		Branch:         branch,
		IncludeAuthors: includeAuthors,
		Filters:        normalizedFilters,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash)
}
