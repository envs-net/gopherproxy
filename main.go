package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/prologic/go-gopher"
)

var (
	bind = flag.String("bind", ":80", "[int]:port to bind to")
	uri  = flag.String("uri", "127.0.0.1:70", "<host>:[port] to proxy to")

	tpl *template.Template
)

type tplRow struct {
	Link template.URL
	Type string
	Text string
}

func renderDirectory(w http.ResponseWriter, tpl *template.Template, hostport string, d gopher.Directory) error {
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
			var hostport string
			if x.Port == 70 {
				hostport = x.Host
			} else {
				hostport = fmt.Sprintf("%s:%d", x.Host, x.Port)
			}
			tr.Link = template.URL(
				fmt.Sprintf(
					"/%s/%s%s",
					hostport,
					string(byte(x.Type)),
					url.QueryEscape(x.Selector),
				),
			)
		}

		out[i] = tr
	}

	return tpl.Execute(w, struct {
		Title string
		Lines []tplRow
	}{hostport, out})
}

func proxy(w http.ResponseWriter, req *http.Request) {
	parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/"), "/")
	hostport := parts[0]
	path := strings.Join(parts[1:], "/")

	if len(hostport) == 0 {
		http.Redirect(w, req, "/"+*uri, http.StatusFound)
		return
	}

	uri, err := url.QueryUnescape(path)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("<b>Error:</b><pre>%s</pre>", err))
		return
	}
	res, err := gopher.Get(
		fmt.Sprintf(
			"gopher://%s/%s",
			hostport,
			uri,
		),
	)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("<b>Error:</b><pre>%s</pre>", err))
		return
	}

	if res.Body != nil {
		io.Copy(w, res.Body)
	} else {
		if err := renderDirectory(w, tpl, hostport, res.Dir); err != nil {
			io.WriteString(w, fmt.Sprintf("<b>Error:</b><pre>%s</pre>", err))
			return
		}
	}
}

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
