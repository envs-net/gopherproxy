package main

import (
	"flag"
	"log"
	"time"

	"github.com/envs-net/gopherproxy"
)

var (
	bind = flag.String("bind", "0.0.0.0:8000", "[int]:port to bind to")

	robotsfile = flag.String("robots-file", "robots.txt", "robots.txt file")
	uri        = flag.String("uri", "envs.net", "<host>:[port] to proxy to; requests for any other Gopher target are rejected")

	cacheTTL          = flag.Duration("cache-ttl", 30*time.Second, "cache lifetime for rendered directories and text responses; 0 disables caching")
	rateLimit         = flag.Int("rate-limit", 0, "maximum HTTP requests per client per minute; 0 disables rate limiting")
	trustProxyHeaders = flag.Bool(
		"trust-proxy-headers",
		false,
		"trust X-Forwarded-For/X-Real-IP for rate limiting; enable only behind a trusted reverse proxy",
	)
)

func main() {
	flag.Parse()

	log.Fatal(gopherproxy.ListenAndServeConfig(gopherproxy.ServerConfig{
		Bind:               *bind,
		RobotsFile:         *robotsfile,
		URI:                *uri,
		CacheTTL:           *cacheTTL,
		RateLimitPerMinute: *rateLimit,
		TrustProxyHeaders:  *trustProxyHeaders,
	}))
}
