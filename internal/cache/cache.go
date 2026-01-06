package cache

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/stripsior/loc.api/internal/models"
)

type CacheEntry struct {
	Result    *models.AnalyzeResponse
	ExpiresAt time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]*CacheEntry
	ttl     time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}

	go cache.cleanupExpired()

	return cache
}

func GenerateCacheKey(repository, branch string, hasToken bool, includeAuthors bool, filters *models.FilterOptions) string {
	data := struct {
		Repository     string
		Branch         string
		HasToken       bool
		IncludeAuthors bool
		Filters        *models.FilterOptions
	}{
		Repository:     repository,
		Branch:         branch,
		HasToken:       hasToken,
		IncludeAuthors: includeAuthors,
		Filters:        filters,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return fmt.Sprintf("%x", hash)
}

func (c *Cache) Get(key string) (*models.AnalyzeResponse, bool) {
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

func (c *Cache) Set(key string, result *models.AnalyzeResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Result:    result,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.entries)
}

func (c *Cache) cleanupExpired() {
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
