# Gopher (RFC 1436) Web Proxy

[![Build Status](https://cloud.drone.io/api/badges/prologic/gopherproxy/status.svg)](https://cloud.drone.io/prologic/gopherproxy)
[![CodeCov](https://codecov.io/gh/prologic/gopherproxy/branch/master/graph/badge.svg)](https://codecov.io/gh/prologic/gopherproxy)
[![Go Report Card](https://goreportcard.com/badge/prologic/gopherproxy)](https://goreportcard.com/report/prologic/gopherproxy)
[![GoDoc](https://godoc.org/github.com/prologic/gopherproxy?status.svg)](https://godoc.org/github.com/prologic/gopherproxy) 
[![Sourcegraph](https://sourcegraph.com/github.com/prologic/gopherproxy/-/badge.svg)](https://sourcegraph.com/github.com/prologic/gopherproxy?badge)

gopherproxy is a Gopher (RFC 1436) Web Proxy that acts as a gateway into Gopherspace
by proxying standard Web HTTP requests to Gopher requests of the target server.

gopherproxy is written in Go (#golang) using the
[go-gopher](https://github.com/prologic/go-gopher) library.

Demo: https://gopher.mills.io/

## Installation
  
  $ go install github.com/prologic/gopherproxy/...

### Docker

Run directly from a prebuild image from the [Docker Hub](https://hub.docker.com):

```#!bash
$ docker run -p 8000:8000 prologic/gopherproxy
```

Or build your own custom image from a source checkout:

```#!bash
$ docker build -t gopherproxy .
$ docker run -p 80:80 gopherproxy -uri floodgap.com
```

## Usage

```#!bash
$ gopherproxy
```

Then simply visit: http://localhost/gopher.floodgap.com

## Related

Related projects:

- [go-gopher](https://github.com/prologic/go-gopher)
  go-gopher is the Gopher client and server library used by gopherproxy

- [gopherclient](https://github.com/prologic/gopherclient)
  gopherclient is a cross-platform QT/QML GUI Gopher Client
  using the gopherproxy library as its backend.

## License

MIT
