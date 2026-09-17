package gopherproxy

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const maxRateLimitEntries = 8192

var crawlerUserAgentMarkers = []string{
	"bot",
	"crawler",
	"spider",
	"slurp",
	"meta-externalagent",
	"facebookexternalhit",
	"facebookcatalog",
	"bingpreview",
	"bytespider",
	"applebot",
	"discordbot",
	"twitterbot",
	"linkedinbot",
	"telegrambot",
	"semrush",
	"ahrefs",
	"yandex",
}

func isLikelyCrawler(userAgent string) bool {
	userAgent = strings.ToLower(strings.TrimSpace(userAgent))
	if userAgent == "" {
		return true
	}

	for _, marker := range crawlerUserAgentMarkers {
		if strings.Contains(userAgent, marker) {
			return true
		}
	}

	return false
}

func normalizeGopherTarget(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("empty Gopher target")
	}
	if strings.ContainsAny(raw, "/?#") {
		return "", fmt.Errorf("invalid Gopher target %q", raw)
	}

	u, err := url.Parse("gopher://" + raw)
	if err != nil {
		return "", fmt.Errorf("parse Gopher target: %w", err)
	}
	if u.User != nil || u.Hostname() == "" {
		return "", fmt.Errorf("invalid Gopher target %q", raw)
	}

	host := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	port := u.Port()
	if port == "" {
		port = "70"
	}

	portNumber, err := strconv.Atoi(port)
	if err != nil || portNumber < 1 || portNumber > 65535 {
		return "", fmt.Errorf("invalid Gopher port %q", port)
	}

	return net.JoinHostPort(host, port), nil
}

func sameGopherTarget(requested, configured string) bool {
	requestedTarget, err := normalizeGopherTarget(requested)
	if err != nil {
		return false
	}
	configuredTarget, err := normalizeGopherTarget(configured)
	if err != nil {
		return false
	}
	return requestedTarget == configuredTarget
}

func clientIP(req *http.Request, trustProxyHeaders bool) string {
	if trustProxyHeaders {
		if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
			parts := strings.Split(forwarded, ",")
			for index := len(parts) - 1; index >= 0; index-- {
				candidate := strings.TrimSpace(parts[index])
				if ip := net.ParseIP(candidate); ip != nil {
					return ip.String()
				}
			}
		}

		if candidate := strings.TrimSpace(req.Header.Get("X-Real-IP")); candidate != "" {
			if ip := net.ParseIP(candidate); ip != nil {
				return ip.String()
			}
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
		return host
	}

	if ip := net.ParseIP(req.RemoteAddr); ip != nil {
		return ip.String()
	}
	return req.RemoteAddr
}

type rateLimitEntry struct {
	windowStart time.Time
	count       int
	lastSeen    time.Time
}

type rateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	entries map[string]rateLimitEntry
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	if limit <= 0 || window <= 0 {
		return nil
	}
	return &rateLimiter{
		limit:   limit,
		window:  window,
		entries: make(map[string]rateLimitEntry),
	}
}

func (r *rateLimiter) allow(key string, now time.Time) bool {
	if r == nil {
		return true
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	entry := r.entries[key]
	if entry.windowStart.IsZero() || now.Sub(entry.windowStart) >= r.window {
		entry.windowStart = now
		entry.count = 0
	}

	entry.count++
	entry.lastSeen = now
	r.entries[key] = entry

	if len(r.entries) > maxRateLimitEntries/2 {
		cutoff := now.Add(-2 * r.window)
		for entryKey, candidate := range r.entries {
			if candidate.lastSeen.Before(cutoff) {
				delete(r.entries, entryKey)
			}
		}
	}

	if len(r.entries) > maxRateLimitEntries {
		var oldestKey string
		var oldestSeen time.Time
		for entryKey, candidate := range r.entries {
			if entryKey == key {
				continue
			}
			if oldestKey == "" || candidate.lastSeen.Before(oldestSeen) {
				oldestKey = entryKey
				oldestSeen = candidate.lastSeen
			}
		}
		if oldestKey != "" {
			delete(r.entries, oldestKey)
		}
	}

	return entry.count <= r.limit
}
