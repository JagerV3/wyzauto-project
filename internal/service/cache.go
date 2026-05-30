package service

import (
	"strings"
	"sync"
	"time"

	"github.com/raymond/wyzauto-project/internal/domain"
)

type cacheEntry struct {
	translations domain.TranslationMap
	expiresAt    time.Time
}

type TranslationCache struct {
	mu      sync.RWMutex
	ttl     time.Duration
	entries map[string]cacheEntry
}

func NewTranslationCache(ttl time.Duration) *TranslationCache {
	return &TranslationCache{
		ttl:     ttl,
		entries: make(map[string]cacheEntry),
	}
}

func (c *TranslationCache) Get(entityType domain.EntityType, entityID string, locales []string) (domain.TranslationMap, bool) {
	if c.ttl <= 0 {
		return nil, false
	}

	key := cacheKey(entityType, entityID, locales)
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		if ok {
			c.mu.Lock()
			delete(c.entries, key)
			c.mu.Unlock()
		}
		return nil, false
	}

	return cloneTranslations(entry.translations), true
}

func (c *TranslationCache) Set(entityType domain.EntityType, entityID string, locales []string, translations domain.TranslationMap) {
	if c.ttl <= 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[cacheKey(entityType, entityID, locales)] = cacheEntry{
		translations: cloneTranslations(translations),
		expiresAt:    time.Now().Add(c.ttl),
	}
}

func (c *TranslationCache) Invalidate(entityType domain.EntityType, entityID string) {
	prefix := string(entityType) + ":" + entityID + ":"

	c.mu.Lock()
	defer c.mu.Unlock()
	for key := range c.entries {
		if strings.HasPrefix(key, prefix) {
			delete(c.entries, key)
		}
	}
}

func cacheKey(entityType domain.EntityType, entityID string, locales []string) string {
	return string(entityType) + ":" + entityID + ":" + strings.Join(locales, ",")
}

func cloneTranslations(src domain.TranslationMap) domain.TranslationMap {
	cloned := make(domain.TranslationMap, len(src))
	for key, value := range src {
		cloned[key] = value
	}
	return cloned
}
