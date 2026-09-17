package gopherproxy

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestSameGopherTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		requested  string
		configured string
		want       bool
	}{
		{name: "same host", requested: "envs.net", configured: "envs.net", want: true},
		{name: "default port", requested: "envs.net:70", configured: "envs.net", want: true},
		{name: "case insensitive", requested: "ENVS.NET", configured: "envs.net", want: true},
		{name: "trailing dot", requested: "envs.net.", configured: "envs.net", want: true},
		{name: "different host", requested: "example.org", configured: "envs.net", want: false},
		{name: "suffix trick", requested: "envs.net.example.org", configured: "envs.net", want: false},
		{name: "different port", requested: "envs.net:7070", configured: "envs.net", want: false},
		{name: "userinfo", requested: "envs.net@127.0.0.1", configured: "envs.net", want: false},
		{name: "path", requested: "envs.net/other", configured: "envs.net", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := sameGopherTarget(test.requested, test.configured); got != test.want {
				t.Fatalf("sameGopherTarget(%q, %q) = %v, want %v", test.requested, test.configured, got, test.want)
			}
		})
	}
}

func TestLikelyCrawler(t *testing.T) {
	t.Parallel()

	for _, userAgent := range []string{
		"meta-externalagent/1.1",
		"Mozilla/5.0 (compatible; Googlebot/2.1)",
		"Twitterbot/1.0",
		"",
	} {
		if !isLikelyCrawler(userAgent) {
			t.Errorf("expected %q to be detected as crawler", userAgent)
		}
	}

	for _, userAgent := range []string{
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/145 Safari/537.36",
		"curl/8.10.1",
	} {
		if isLikelyCrawler(userAgent) {
			t.Errorf("did not expect %q to be detected as crawler", userAgent)
		}
	}
}

func TestClientIP(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest("GET", "http://example.test/", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "198.51.100.99, 203.0.113.20")
	req.Header.Set("X-Real-IP", "203.0.113.21")

	if got := clientIP(req, false); got != "127.0.0.1" {
		t.Fatalf("untrusted proxy headers returned %q", got)
	}
	if got := clientIP(req, true); got != "203.0.113.20" {
		t.Fatalf("trusted proxy headers returned %q", got)
	}
}

func TestRateLimiter(t *testing.T) {
	t.Parallel()

	limiter := newRateLimiter(2, time.Minute)
	now := time.Unix(1_700_000_000, 0)

	if !limiter.allow("client", now) {
		t.Fatal("first request should be allowed")
	}
	if !limiter.allow("client", now.Add(time.Second)) {
		t.Fatal("second request should be allowed")
	}
	if limiter.allow("client", now.Add(2*time.Second)) {
		t.Fatal("third request should be rejected")
	}
	if !limiter.allow("client", now.Add(time.Minute)) {
		t.Fatal("request in new window should be allowed")
	}
}
