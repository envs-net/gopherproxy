package main

import (
	"flag"
	"log"

	"github.com/prologic/gopherproxy"
)

var (
	bind = flag.String("bind", ":80", "[int]:port to bind to")
	uri  = flag.String("uri", "127.0.0.1:70", "<host>:[port] to proxy to")
)

func main() {
	flag.Parse()

	log.Fatal(gopherproxy.ListenAndServe(*bind, *uri))
}
