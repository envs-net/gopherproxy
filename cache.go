package gopherproxy

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	defaultCacheEntries = 128
	maxCachedBodyBytes  = 256 << 10 // 256 KiB; at most ~32 MiB across 128 entries
)

type cacheEntry struct {
	body        []byte
	contentType string
	expiresAt   time.Time
}

type responseCache struct {
	mu         sync.Mutex
	ttl        time.Duration
	maxEntries int
	entries    map[string]cacheEntry
}

func newResponseCache(ttl time.Duration, maxEntries int) *responseCache {
	if ttl <= 0 || maxEntries <= 0 {
		return nil
	}
	return &responseCache{
		ttl:        ttl,
		maxEntries: maxEntries,
		entries:    make(map[string]cacheEntry),
	}
}

func (c *responseCache) get(key string, now time.Time) (cacheEntry, bool) {
	if c == nil {
		return cacheEntry{}, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return cacheEntry{}, false
	}
	if !entry.expiresAt.After(now) {
		delete(c.entries, key)
		return cacheEntry{}, false
	}

	entry.body = append([]byte(nil), entry.body...)
	return entry, true
}

func (c *responseCache) set(key, contentType string, body []byte, now time.Time) {
	if c == nil || len(body) > maxCachedBodyBytes {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.entries) >= c.maxEntries {
		for entryKey, entry := range c.entries {
			if !entry.expiresAt.After(now) {
				delete(c.entries, entryKey)
			}
		}
	}

	if len(c.entries) >= c.maxEntries {
		var oldestKey string
		var oldestExpiry time.Time
		for entryKey, entry := range c.entries {
			if oldestKey == "" || entry.expiresAt.Before(oldestExpiry) {
				oldestKey = entryKey
				oldestExpiry = entry.expiresAt
			}
		}
		if oldestKey != "" {
			delete(c.entries, oldestKey)
		}
	}

	c.entries[key] = cacheEntry{
		body:        append([]byte(nil), body...),
		contentType: contentType,
		expiresAt:   now.Add(c.ttl),
	}
}

func writeCachedResponse(w http.ResponseWriter, entry cacheEntry, ttl time.Duration, method string) {
	if entry.contentType != "" {
		w.Header().Set("Content-Type", entry.contentType)
	}
	w.Header().Set("Cache-Control", cacheControlValue(ttl))
	w.Header().Set("X-Gopherproxy-Cache", "HIT")
	if method != http.MethodHead {
		_, _ = w.Write(entry.body)
	}
}

func cacheControlValue(ttl time.Duration) string {
	seconds := int(ttl / time.Second)
	if seconds < 0 {
		seconds = 0
	}
	return "public, max-age=" + strconv.Itoa(seconds)
}
