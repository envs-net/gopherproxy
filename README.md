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

## Installation
  
  $ go install github.com/prologic/gopherproxy/...

### Docker

```#!bash
$ docker build -t gopherproxy .
$ docker run -p 80:80 gopherproxy -uri floodgap.com
```

## Usage

```#!bash
$ gopherproxy
```

Then simply visit: http://localhost/gopher.floodgap.com

## License

MIT
