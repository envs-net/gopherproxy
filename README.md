# gopherproxy

`gopherproxy` exposes a Gopher server through HTTP. This fork is configured as a
**fixed-target gateway**: the host encoded in an incoming URL must match the
server selected with `-uri`.

That restriction is intentional. Older versions accepted an arbitrary Gopher
host from the request path, which made a public deployment an open
HTTP-to-Gopher proxy. Crawlers could therefore make the proxy repeatedly fetch
resources from unrelated Gopher servers while all traffic appeared to originate
from the proxy host.

## Security model

The current implementation has the following safeguards:

- **Fixed upstream target.** `-uri envs.net` only permits requests for
  `envs.net` (with port `70` treated as the default). Requests for another host
  or port return HTTP 403 before any Gopher connection is opened.
- **robots.txt enforcement for crawlers.** Known crawler/bot user agents that
  violate the configured `robots.txt` are rejected instead of only being
  logged. Normal browsers are not blocked solely because `User-agent: *` is
  disallowed.
- **Short bounded cache.** Rendered Gopher directories and rendered `.txt` /`.md`
  responses can be cached briefly. Individual cache entries are limited to
  256 KiB and the cache is bounded to 128 entries (about 32 MiB maximum cached
  body data). Binary responses are streamed and are not buffered in the cache.
- **Optional rate limiting.** A fixed per-client request limit can be enabled
  with `-rate-limit`.
- **Bounded rendered text.** `.txt` and `.md` files larger than 4 MiB are not
  rendered into memory.
- **Safer Markdown and links.** Raw HTML and unsafe Markdown links are skipped,
  and Gopher `URL:` menu entries are left to `html/template` URL sanitization.
- **Connection cleanup and HTTP timeouts.** Upstream response bodies are closed,
  and the HTTP server uses header/idle timeouts.

The fixed-target check is the primary security boundary. User-Agent detection
is necessarily heuristic and must not be treated as one.

## Build

The project uses Go modules:

```sh
go mod tidy
go test ./...
go vet ./...
go build -trimpath -o gopherproxy ./cmd/gopherproxy
```

Or simply:

```sh
make test
make vet
make build
```

## Usage

```sh
./gopherproxy \
  -bind 127.0.0.1:8993 \
  -uri envs.net \
  -robots-file robots.txt \
  -cache-ttl 30s
```

Opening `/` redirects to the configured target, for example `/envs.net`.
Requests such as `/example.org/...` are rejected when `-uri envs.net` is in
use. Gopher menu entries pointing to another Gopher host are rendered as text
instead of as dead proxy links; ordinary `URL:` HTTP/HTTPS links remain usable.

### Options

```text
-bind <address>              HTTP listen address (default 0.0.0.0:8000)
-uri <host[:port]>           only Gopher upstream that may be fetched
-robots-file <path>          robots.txt file (default robots.txt)
-cache-ttl <duration>        rendered-response cache TTL (default 30s; 0 disables)
-rate-limit <n>              requests per client per minute (default 0; disabled)
-trust-proxy-headers         use X-Forwarded-For/X-Real-IP for rate limiting
```

## nginx deployment

For the envs.net deployment, keep gopherproxy bound to loopback and expose it
through nginx. If application-level rate limiting is enabled, proxy headers may
be trusted because nginx is the only process connecting to the loopback
listener:

```nginx
location / {
    proxy_pass http://127.0.0.1:8993;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
}
```

A suitable service command is then:

```sh
/opt/services/go/src/github.com/envs-net/gopherproxy/gopherproxy \
  -bind 127.0.0.1:8993 \
  -uri envs.net \
  -robots-file /opt/services/go/src/github.com/envs-net/gopherproxy/robots.txt \
  -cache-ttl 30s \
  -rate-limit 60 \
  -trust-proxy-headers
```

Only enable `-trust-proxy-headers` when the HTTP listener is reachable solely
through a reverse proxy you control. When `X-Forwarded-For` contains multiple
addresses, gopherproxy uses the rightmost valid address, matching a trusted
single nginx hop using `$proxy_add_x_forwarded_for`.

## robots.txt

The included file currently disallows crawling of the whole HTTP gateway:

```text
User-agent: *
Disallow: /
```

Compliant crawlers should stop after reading it. In addition, gopherproxy now
actively rejects recognizable crawlers that attempt to continue anyway.

## Docker

```sh
docker build -t gopherproxy .
docker run --rm -p 8000:8000 gopherproxy -uri envs.net
```

## License

MIT
