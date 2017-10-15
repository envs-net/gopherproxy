package gopherproxy

import (
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

type Item struct {
	Link template.URL
	Type string
	Text string
}

func renderDirectory(w http.ResponseWriter, tpl *template.Template, hostport string, d gopher.Directory) error {
	var title string

	out := make([]Item, len(d.Items))

	for i, x := range d.Items {
		if x.Type == gopher.INFO && x.Selector == "TITLE" {
			title = x.Description
			continue
		}

		tr := Item{
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
			path := url.PathEscape(x.Selector)
			path = strings.Replace(path, "%2F", "/", -1)
			tr.Link = template.URL(
				fmt.Sprintf(
					"/%s/%s%s",
					hostport,
					string(byte(x.Type)),
					path,
				),
			)
		}

		out[i] = tr
	}

	if title == "" {
		title = hostport
	}

	return tpl.Execute(w, struct {
		Title string
		Lines []Item
	}{title, out})
}

// Handler returns a Handler that proxies requests
// to the specified Gopher server as denoated by the first argument
// to the request path and renders the content using the provided template.
func Handler(tpl *template.Template, uri string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		parts := strings.Split(strings.TrimPrefix(req.URL.Path, "/"), "/")
		hostport := parts[0]

		if len(hostport) == 0 {
			http.Redirect(w, req, "/"+uri, http.StatusFound)
			return
		}

		var qs string

		path := strings.Join(parts[1:], "/")

		if req.URL.RawQuery != "" {
			qs = fmt.Sprintf("?%s", url.QueryEscape(req.URL.RawQuery))
		}

		uri, err := url.QueryUnescape(path)
		if err != nil {
			io.WriteString(w, fmt.Sprintf("<b>Error:</b><pre>%s</pre>", err))
			return
		}

		res, err := gopher.Get(
			fmt.Sprintf(
				"gopher://%s/%s%s",
				hostport,
				uri,
				qs,
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
}

// ListenAndServe creates a listening HTTP server bound to
// the interface specified by bind and sets up a Gopher to HTTP
// proxy proxying requests as requested and by default will prozy
// to a Gopher server address specified by uri if no servers is
// specified by the request.
func ListenAndServe(bind, uri string) error {
	var tpl *template.Template

	tpldata, err := ioutil.ReadFile(".template")
	if err == nil {
		tpltext = string(tpldata)
	}

	tpl, err = template.New("gophermenu").Parse(tpltext)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", Handler(tpl, uri))
	return http.ListenAndServe(bind, nil)
}
