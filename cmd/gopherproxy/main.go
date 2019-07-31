package main

import (
	"flag"
	"log"

	"tildegit.org/tildeverse/gopherproxy"
)

var (
	// TODO: Allow config file and environment vars
	//       (opt -> env -> config -> default)
	bind       = flag.String("bind", "0.0.0.0:8000", "[int]:port to bind to")
	robotsfile = flag.String("robots-file", "robots.txt", "robots.txt file")
	uri        = flag.String("uri", "tilde.team", "<host>:[port] to proxy to")
)

func main() {
	flag.Parse()

	// Use a config struct
	log.Fatal(gopherproxy.ListenAndServe(*bind, *robotsfile, *uri))
}
