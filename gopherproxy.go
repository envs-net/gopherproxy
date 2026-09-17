package gopherproxy

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	gopher "github.com/prologic/go-gopher"
	"github.com/temoto/robotstxt"
)

const maxRenderedTextBytes = 4 << 20 // 4 MiB

type Item struct {
	Link string
	Type string
	Text string
}

type gopherFetchFunc func(string) (*gopher.Response, error)

type handlerOptions struct {
	fetch             gopherFetchFunc
	cache             *responseCache
	limiter           *rateLimiter
	trustProxyHeaders bool
}

// ServerConfig controls the HTTP-facing Gopher proxy.
type ServerConfig struct {
	Bind               string
	RobotsFile         string
	URI                string
	CacheTTL           time.Duration
	RateLimitPerMinute int
	TrustProxyHeaders  bool
}

func renderDirectory(w io.Writer, tpl *template.Template, hostport string, d gopher.Directory) error {
	var title string

	out := make([]Item, 0, len(d.Items))

	for _, x := range d.Items {
		if x.Type == gopher.INFO && x.Selector == "TITLE" {
			title = x.Description
			continue
		}

		tr := Item{
			Text: x.Description,
			Type: x.Type.String(),
		}

		if x.Type == gopher.INFO {
			out = append(out, tr)
			continue
		}

		if strings.HasPrefix(x.Selector, "URL:") {
			tr.Link = x.Selector[4:]
		} else {
			itemHostport := x.Host
			if x.Port != 70 {
				itemHostport = fmt.Sprintf("%s:%d", x.Host, x.Port)
			}

			// Do not advertise proxy URLs that this fixed-target gateway will
			// reject. This also keeps crawlers from repeatedly probing foreign
			// Gopher hosts through the HTTP endpoint.
			if !sameGopherTarget(itemHostport, hostport) {
				out = append(out, tr)
				continue
			}

			path := url.PathEscape(x.Selector)
			path = strings.ReplaceAll(path, "%2F", "/")
			tr.Link = fmt.Sprintf(
				"/%s/%s%s",
				itemHostport,
				string(byte(x.Type)),
				path,
			)
		}

		out = append(out, tr)
	}

	if title == "" {
		title = hostport
	}

	return tpl.Execute(w, struct {
		Title     string
		Lines     []Item
		Gophermap bool
	}{title, out, true})
}

// GopherHandler returns a handler that proxies only to uri. The host encoded in
// the request path must resolve to the same Gopher host and port as uri. This
// prevents the HTTP endpoint from becoming an open HTTP-to-Gopher proxy.
//
// If robotsdata disallows a request and its User-Agent looks like a crawler,
// the request is rejected instead of merely being logged.
func GopherHandler(tpl *template.Template, robotsdata *robotstxt.RobotsData, uri string) http.HandlerFunc {
	return newGopherHandler(tpl, robotsdata, uri, handlerOptions{fetch: gopher.Get})
}

func newGopherHandler(
	tpl *template.Template,
	robotsdata *robotstxt.RobotsData,
	configuredURI string,
	options handlerOptions,
) http.HandlerFunc {
	fetch := options.fetch
	if fetch == nil {
		fetch = gopher.Get
	}

	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodGet && req.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if options.limiter != nil {
			requester := clientIP(req, options.trustProxyHeaders)
			if !options.limiter.allow(requester, time.Now()) {
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				log.Printf("rate limit exceeded for %s", requester)
				return
			}
		}

		escapedPath := strings.TrimPrefix(req.URL.EscapedPath(), "/")
		if escapedPath == "" {
			http.Redirect(w, req, "/"+configuredURI, http.StatusFound)
			return
		}

		parts := strings.SplitN(escapedPath, "/", 2)
		hostport, err := url.PathUnescape(parts[0])
		if err != nil || hostport == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		if !sameGopherTarget(hostport, configuredURI) {
			log.Printf("blocked proxy request for disallowed Gopher target %q", hostport)
			http.Error(w, "Forbidden Gopher target", http.StatusForbidden)
			return
		}

		agent := req.UserAgent()
		if robotsdata != nil && isLikelyCrawler(agent) && !robotsdata.TestAgent(req.URL.Path, agent) {
			log.Printf("blocked UserAgent %q by robots.txt", agent)
			http.Error(w, "Forbidden by robots.txt", http.StatusForbidden)
			return
		}

		selector := ""
		if len(parts) == 2 {
			selector, err = url.PathUnescape(parts[1])
			if err != nil {
				http.Error(w, "Bad Request", http.StatusBadRequest)
				return
			}
		}

		var querySuffix string
		if req.URL.RawQuery != "" {
			querySuffix = fmt.Sprintf("?%s", url.QueryEscape(req.URL.RawQuery))
		}

		normalizedTarget, _ := normalizeGopherTarget(configuredURI)
		cacheKey := normalizedTarget + "\x00" + selector + "\x00" + req.URL.RawQuery
		if entry, ok := options.cache.get(cacheKey, time.Now()); ok {
			writeCachedResponse(w, entry, options.cache.ttl, req.Method)
			return
		}

		res, err := fetch(fmt.Sprintf("gopher://%s/%s%s", configuredURI, selector, querySuffix))
		if err != nil {
			http.Error(w, "Gopher upstream request failed", http.StatusBadGateway)
			log.Printf("Gopher request to %q failed: %v", configuredURI, err)
			return
		}

		if res.Body != nil {
			defer res.Body.Close()

			switch {
			case strings.HasSuffix(strings.ToLower(selector), ".md"):
				body, err := readRenderedText(res.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadGateway)
					return
				}

				var rendered bytes.Buffer
				if err := tpl.Execute(&rendered, struct {
					Title     string
					MdText    template.HTML
					Gophermap bool
					Pre       bool
				}{selector, renderMd(bytes.NewReader(body)), false, false}); err != nil {
					http.Error(w, "Template rendering failed", http.StatusInternalServerError)
					return
				}
				writeCacheableResponse(w, options.cache, cacheKey, "text/html; charset=utf-8", rendered.Bytes(), req.Method)
				return

			case strings.HasSuffix(strings.ToLower(selector), ".txt"):
				body, err := readRenderedText(res.Body)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadGateway)
					return
				}

				var rendered bytes.Buffer
				if err := tpl.Execute(&rendered, struct {
					Title     string
					MdText    string
					Gophermap bool
					Pre       bool
				}{selector, string(body), false, true}); err != nil {
					http.Error(w, "Template rendering failed", http.StatusInternalServerError)
					return
				}
				writeCacheableResponse(w, options.cache, cacheKey, "text/html; charset=utf-8", rendered.Bytes(), req.Method)
				return
			}

			w.Header().Set("X-Gopherproxy-Cache", "BYPASS")
			if req.Method != http.MethodHead {
				if _, err := io.Copy(w, res.Body); err != nil {
					log.Printf("streaming Gopher response failed: %v", err)
				}
			}
			return
		}

		var rendered bytes.Buffer
		if err := renderDirectory(&rendered, tpl, hostport, res.Dir); err != nil {
			http.Error(w, "Template rendering failed", http.StatusInternalServerError)
			return
		}
		writeCacheableResponse(w, options.cache, cacheKey, "text/html; charset=utf-8", rendered.Bytes(), req.Method)
	}
}

func readRenderedText(reader io.Reader) ([]byte, error) {
	limited := io.LimitReader(reader, maxRenderedTextBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read Gopher response: %w", err)
	}
	if len(body) > maxRenderedTextBytes {
		return nil, fmt.Errorf("Gopher text response exceeds %d bytes", maxRenderedTextBytes)
	}
	return body, nil
}

func writeCacheableResponse(
	w http.ResponseWriter,
	cache *responseCache,
	cacheKey string,
	contentType string,
	body []byte,
	method string,
) {
	w.Header().Set("Content-Type", contentType)
	if cache != nil {
		cache.set(cacheKey, contentType, body, time.Now())
		w.Header().Set("Cache-Control", cacheControlValue(cache.ttl))
		w.Header().Set("X-Gopherproxy-Cache", "MISS")
	} else {
		w.Header().Set("X-Gopherproxy-Cache", "DISABLED")
	}
	if method != http.MethodHead {
		_, _ = w.Write(body)
	}
}

// RobotsTxtHandler returns the contents of the robots.txt file if configured
// and valid.
func RobotsTxtHandler(robotstxtdata []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if robotstxtdata == nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if req.Method != http.MethodHead {
			_, _ = w.Write(robotstxtdata)
		}
	}
}

// ListenAndServe preserves the historic public API while using secure defaults:
// only uri is allowed as an upstream target and small rendered responses are
// cached briefly. Per-client rate limiting remains opt-in because deployments
// behind a reverse proxy need an explicit proxy-header trust decision.
func ListenAndServe(bind, robotsfile, uri string) error {
	return ListenAndServeConfig(ServerConfig{
		Bind:       bind,
		RobotsFile: robotsfile,
		URI:        uri,
		CacheTTL:   30 * time.Second,
	})
}

// ListenAndServeConfig starts the HTTP server using cfg.
func ListenAndServeConfig(cfg ServerConfig) error {
	if cfg.Bind == "" {
		cfg.Bind = "0.0.0.0:8000"
	}
	if cfg.URI == "" {
		return fmt.Errorf("Gopher target URI must not be empty")
	}
	if _, err := normalizeGopherTarget(cfg.URI); err != nil {
		return err
	}

	var robotsdata *robotstxt.RobotsData
	var robotstxtdata []byte
	if cfg.RobotsFile != "" {
		data, err := os.ReadFile(cfg.RobotsFile)
		if err != nil {
			log.Printf("error reading robots.txt: %s", err)
		} else {
			parsed, parseErr := robotstxt.FromBytes(data)
			if parseErr != nil {
				log.Printf("error parsing robots.txt: %s", parseErr)
			} else {
				robotstxtdata = data
				robotsdata = parsed
			}
		}
	}

	if tpldata, err := os.ReadFile(".template"); err == nil {
		tpltext = string(tpldata)
	}

	tpl, err := template.New("gophermenu").Parse(tpltext)
	if err != nil {
		return fmt.Errorf("parse template: %w", err)
	}

	cache := newResponseCache(cfg.CacheTTL, defaultCacheEntries)
	limiter := newRateLimiter(cfg.RateLimitPerMinute, time.Minute)

	mux := http.NewServeMux()
	mux.HandleFunc("/robots.txt", RobotsTxtHandler(robotstxtdata))
	mux.HandleFunc("/", newGopherHandler(tpl, robotsdata, cfg.URI, handlerOptions{
		fetch:             gopher.Get,
		cache:             cache,
		limiter:           limiter,
		trustProxyHeaders: cfg.TrustProxyHeaders,
	}))

	server := &http.Server{
		Addr:              cfg.Bind,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	return server.ListenAndServe()
}
