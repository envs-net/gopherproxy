# Gopher (RFC 1436) Web Proxy

[![Build Status](https://travis-ci.org/prologic/gopherproxy.svg)](https://travis-ci.org/prologic/gopherproxy)

This is a Gopher (RFC 1436) Web Proxy that acts as a gateway into Gopherspace
by proxying standard Web HTTP requests to Gopher requests of the target server.

gopherproxy is written in Go (#golang) using the
[go-gopher](https://github.com/prologic/go-gopher) library.

## Installation
  
  $ go install github.com/prologic/gopherproxy

## Usage

```#!bash
$ gopherproxy
```

Then simply visit: http://localhost/gopher.floodgap.com

## License

MIT
