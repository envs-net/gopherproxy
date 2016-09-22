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

By default gopherproxy will proxy requests to a locally running Gopher server
on gopher://localhost:70/ -- To change where to proxy to:

```#!bash
$ gopherproxy -host gopher.floodgap.com
```

## License

MIT
