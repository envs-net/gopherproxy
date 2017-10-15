package main

import (
	"flag"
	"log"

	"github.com/prologic/gopherproxy"
)

var (
	bind = flag.String("bind", "0.0.0.0:8000", "[int]:port to bind to")
	uri  = flag.String("uri", "floodgap.com", "<host>:[port] to proxy to")
)

func main() {
	flag.Parse()

	log.Fatal(gopherproxy.ListenAndServe(*bind, *uri))
}
