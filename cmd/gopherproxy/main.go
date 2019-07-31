package main

import (
	"flag"
	"log"

	"github.com/envs-net/gopherproxy"
)

var (
	// TODO: Allow config file and environment vars
	//       (opt -> env -> config -> default)
	bind       = flag.String("bind", "0.0.0.0:8000", "[int]:port to bind to")
	robotsfile = flag.String("robots-file", "robots.txt", "robots.txt file")
	uri        = flag.String("uri", "envs.net", "<host>:[port] to proxy to")
)

func main() {
	flag.Parse()

	// Use a config struct
	log.Fatal(gopherproxy.ListenAndServe(*bind, *robotsfile, *uri))
}
