# Gopher (RFC 1436) Web Proxy
[![Build Status](https://travis-ci.org/prologic/gopherproxy.svg)](https://travis-ci.org/prologic/gopherproxy)
[![GoDoc](https://godoc.org/github.com/prologic/gopherproxy?status.svg)](https://godoc.org/github.com/prologic/gopherproxy)
[![Wiki](https://img.shields.io/badge/docs-wiki-blue.svg)](https://github.com/prologic/gopherproxy/wiki)
[![Go Report Card](https://goreportcard.com/badge/github.com/prologic/gopherproxy)](https://goreportcard.com/report/github.com/prologic/gopherproxy)
[![Coverage](https://coveralls.io/repos/prologic/gopherproxy/badge.svg)](https://coveralls.io/r/prologic/gopherproxy)

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
