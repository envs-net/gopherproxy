package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/prologic/go-gopher"
)

var (
	bind = flag.String("bind", ":80", "[int]:port to bind to")
	host = flag.String("host", "localhost", "host to proxy to")
	port = flag.Int("port", 70, "port to proxy to")
)

func proxy(res http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/")

	gr, err := gopher.Get(fmt.Sprintf("gopher://%s:%d/%s", *host, *port, path))
	if err != nil {
		io.WriteString(res, fmt.Sprintf("<b>Error:</b><pre>%s</pre>", err))
		return
	}

	if gr.Body != nil {
		io.Copy(res, gr.Body)
	} else {
		bytes, err := gr.Dir.ToText()
		if err != nil {
			io.WriteString(res, fmt.Sprintf("<b>Error:</b><pre>%s</pre>", err))
			return
		}

		io.WriteString(res, string(bytes))
	}
}

func main() {
	flag.Parse()

	http.HandleFunc("/", proxy)
	log.Fatal(http.ListenAndServe(*bind, nil))
}
