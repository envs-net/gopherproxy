package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
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

type tplRow struct {
	Link template.URL
	Type string
	Text string
}

func renderDirectory(w http.ResponseWriter, tpl *template.Template, d gopher.Directory) error {
	out := make([]tplRow, len(d))

	for i, x := range d {
		tr := tplRow{
			Text: x.Description,
			Type: x.Type.String(),
		}

		if x.Type == gopher.INFO {
			out[i] = tr
			continue
		}

		if strings.HasPrefix(x.Selector, "URL:") {
			tr.Link = template.URL(x.Selector[4:])
		} else {
			tr.Link = template.URL(
				fmt.Sprintf(
					"%s%s", string(byte(x.Type)), x.Selector,
				),
			)
		}

		out[i] = tr
	}

	return tpl.Execute(w, struct {
		Title string
		Lines []tplRow
	}{"XXX", out})
}

func proxy(w http.ResponseWriter, req *http.Request) {
	path := strings.TrimPrefix(req.URL.Path, "/")

	res, err := gopher.Get(fmt.Sprintf("gopher://%s:%d/%s", *host, *port, path))
	if err != nil {
		io.WriteString(w, fmt.Sprintf("<b>Error:</b><pre>%s</pre>", err))
		return
	}

	if res.Body != nil {
		io.Copy(w, res.Body)
	} else {
		if err := renderDirectory(w, tpl, res.Dir); err != nil {
			io.WriteString(w, fmt.Sprintf("<b>Error:</b><pre>%s</pre>", err))
			return
		}
	}
}

var tpl *template.Template

func main() {
	flag.Parse()

	tpldata, err := ioutil.ReadFile(".template")
	if err == nil {
		tpltext = string(tpldata)
	}

	tpl, err = template.New("gophermenu").Parse(tpltext)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", proxy)
	log.Fatal(http.ListenAndServe(*bind, nil))
}
