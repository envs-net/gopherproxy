package gopherproxy

import (
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	gopher "github.com/prologic/go-gopher"
	"github.com/temoto/robotstxt"
)

type trackingReadCloser struct {
	io.Reader
	closed atomic.Bool
}

func (r *trackingReadCloser) Close() error {
	r.closed.Store(true)
	return nil
}

func testTemplate(t *testing.T) *template.Template {
	t.Helper()
	return template.Must(template.New("test").Parse(`{{if .Gophermap}}{{range .Lines}}{{.Text}}{{end}}{{else}}{{.MdText}}{{end}}`))
}

func textResponse(body *trackingReadCloser) *gopher.Response {
	return &gopher.Response{
		Type: gopher.FILE,
		Body: body,
	}
}

func TestGopherHandlerRejectsExternalTargetWithoutFetching(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	handler := newGopherHandler(testTemplate(t), nil, "envs.net", handlerOptions{
		fetch: func(string) (*gopher.Response, error) {
			calls.Add(1)
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "http://proxy.test/example.org/0hello.txt", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if calls.Load() != 0 {
		t.Fatalf("fetch called %d times", calls.Load())
	}
}

func TestGopherHandlerBlocksCrawlerWhenRobotsDisallow(t *testing.T) {
	t.Parallel()

	robots, err := robotstxt.FromString("User-agent: *\nDisallow: /\n")
	if err != nil {
		t.Fatal(err)
	}

	var calls atomic.Int32
	handler := newGopherHandler(testTemplate(t), robots, "envs.net", handlerOptions{
		fetch: func(string) (*gopher.Response, error) {
			calls.Add(1)
			return nil, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "http://proxy.test/envs.net/0hello.txt", nil)
	req.Header.Set("User-Agent", "meta-externalagent/1.1")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	if calls.Load() != 0 {
		t.Fatalf("fetch called %d times", calls.Load())
	}
}

func TestGopherHandlerAllowsBrowserAndClosesBody(t *testing.T) {
	t.Parallel()

	robots, err := robotstxt.FromString("User-agent: *\nDisallow: /\n")
	if err != nil {
		t.Fatal(err)
	}

	body := &trackingReadCloser{Reader: strings.NewReader("hello")}
	var fetchedURI string
	handler := newGopherHandler(testTemplate(t), robots, "envs.net", handlerOptions{
		fetch: func(uri string) (*gopher.Response, error) {
			fetchedURI = uri
			return textResponse(body), nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "http://proxy.test/ENVS.NET:70/0hello.txt", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 Chrome/145.0.0.0 Safari/537.36")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if fetchedURI != "gopher://envs.net/0hello.txt" {
		t.Fatalf("fetched URI = %q", fetchedURI)
	}
	if !body.closed.Load() {
		t.Fatal("upstream response body was not closed")
	}
	if !strings.Contains(recorder.Body.String(), "hello") {
		t.Fatalf("response = %q", recorder.Body.String())
	}
}

func TestGopherHandlerCachesRenderedText(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	handler := newGopherHandler(testTemplate(t), nil, "envs.net", handlerOptions{
		fetch: func(string) (*gopher.Response, error) {
			calls.Add(1)
			return &gopher.Response{
				Type: gopher.FILE,
				Body: io.NopCloser(strings.NewReader("cached text")),
			}, nil
		},
		cache: newResponseCache(time.Minute, 16),
	})

	for index := 0; index < 2; index++ {
		req := httptest.NewRequest(http.MethodGet, "http://proxy.test/envs.net/0cached.txt", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0")
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Fatalf("request %d status = %d", index, recorder.Code)
		}
		wantCache := "MISS"
		if index == 1 {
			wantCache = "HIT"
		}
		if got := recorder.Header().Get("X-Gopherproxy-Cache"); got != wantCache {
			t.Fatalf("request %d cache header = %q, want %q", index, got, wantCache)
		}
	}

	if calls.Load() != 1 {
		t.Fatalf("fetch called %d times, want 1", calls.Load())
	}
}

func TestGopherHandlerRateLimits(t *testing.T) {
	t.Parallel()

	handler := newGopherHandler(testTemplate(t), nil, "envs.net", handlerOptions{
		fetch: func(string) (*gopher.Response, error) {
			return &gopher.Response{
				Type: gopher.FILE,
				Body: io.NopCloser(strings.NewReader("ok")),
			}, nil
		},
		limiter: newRateLimiter(1, time.Minute),
	})

	first := httptest.NewRequest(http.MethodGet, "http://proxy.test/envs.net/0file", nil)
	first.RemoteAddr = "192.0.2.10:1234"
	firstRecorder := httptest.NewRecorder()
	handler.ServeHTTP(firstRecorder, first)
	if firstRecorder.Code != http.StatusOK {
		t.Fatalf("first status = %d", firstRecorder.Code)
	}

	second := httptest.NewRequest(http.MethodGet, "http://proxy.test/envs.net/0file", nil)
	second.RemoteAddr = "192.0.2.10:5678"
	secondRecorder := httptest.NewRecorder()
	handler.ServeHTTP(secondRecorder, second)
	if secondRecorder.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", secondRecorder.Code, http.StatusTooManyRequests)
	}
}

func TestRenderDirectorySanitizesUnsafeURLSelector(t *testing.T) {
	t.Parallel()

	tpl := template.Must(template.New("links").Parse(`{{range .Lines}}<a href="{{.Link}}">{{.Text}}</a>{{end}}`))
	directory := gopher.Directory{Items: []*gopher.Item{
		{
			Type:        gopher.HTML,
			Description: "unsafe",
			Selector:    "URL:javascript:alert(1)",
			Host:        "envs.net",
			Port:        70,
		},
	}}

	var output strings.Builder
	if err := renderDirectory(&output, tpl, "envs.net", directory); err != nil {
		t.Fatal(err)
	}

	lower := strings.ToLower(output.String())
	if strings.Contains(lower, "javascript:") {
		t.Fatalf("unsafe URL survived template sanitization: %s", output.String())
	}
}

func TestRenderDirectoryDoesNotLinkExternalGopherTarget(t *testing.T) {
	t.Parallel()

	tpl := template.Must(template.New("links").Parse(`{{range .Lines}}{{if .Link}}<a href="{{.Link}}">{{.Text}}</a>{{else}}{{.Text}}{{end}}{{end}}`))
	directory := gopher.Directory{Items: []*gopher.Item{
		{
			Type:        gopher.FILE,
			Description: "external",
			Selector:    "hello.txt",
			Host:        "example.org",
			Port:        70,
		},
	}}

	var output strings.Builder
	if err := renderDirectory(&output, tpl, "envs.net", directory); err != nil {
		t.Fatal(err)
	}

	if strings.Contains(output.String(), "href=") {
		t.Fatalf("external Gopher target was rendered as proxy link: %s", output.String())
	}
}
