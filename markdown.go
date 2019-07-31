package gopherproxy

import (
	"bytes"
	"html/template"
	"io"
	"strings"

	"github.com/gomarkdown/markdown"
)

func renderMd(b io.Reader) template.HTML {
	buf := new(bytes.Buffer)
	buf.ReadFrom(b)
	md := string(markdown.ToHTML(buf.Bytes(), nil, nil))
	return template.HTML(strings.Replace(md, "<img", "<img class=\"img-responsive\"", -1))
}
