package gopherproxy

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"time"
)

func TestResponseCacheExpires(t *testing.T) {
	t.Parallel()

	cache := newResponseCache(time.Second, 2)
	now := time.Unix(1_700_000_000, 0)
	cache.set("key", "text/plain", []byte("value"), now)

	entry, ok := cache.get("key", now.Add(500*time.Millisecond))
	if !ok || !bytes.Equal(entry.body, []byte("value")) {
		t.Fatalf("expected cached value, got %#v, %v", entry, ok)
	}

	if _, ok := cache.get("key", now.Add(2*time.Second)); ok {
		t.Fatal("expected cache entry to expire")
	}
}

func TestWriteCachedResponseHeadHasNoBody(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	writeCachedResponse(recorder, cacheEntry{
		body:        []byte("cached"),
		contentType: "text/plain",
	}, time.Minute, "HEAD")

	if recorder.Body.Len() != 0 {
		t.Fatalf("HEAD response body = %q", recorder.Body.String())
	}
	if got := recorder.Header().Get("X-Gopherproxy-Cache"); got != "HIT" {
		t.Fatalf("cache header = %q", got)
	}
}

func TestResponseCacheSkipsLargeBodies(t *testing.T) {
	t.Parallel()

	cache := newResponseCache(time.Minute, 2)
	body := make([]byte, maxCachedBodyBytes+1)
	cache.set("too-large", "application/octet-stream", body, time.Now())

	if _, ok := cache.get("too-large", time.Now()); ok {
		t.Fatal("oversized body should not be cached")
	}
}
