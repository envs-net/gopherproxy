package gopherproxy

import (
	"bytes"
	"html/template"
	"io"
	"strings"

	"github.com/gomarkdown/markdown"
	markdownhtml "github.com/gomarkdown/markdown/html"
)

func renderMd(reader io.Reader) template.HTML {
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(reader)

	renderer := markdownhtml.NewRenderer(markdownhtml.RendererOptions{
		Flags: markdownhtml.SkipHTML |
			markdownhtml.Safelink |
			markdownhtml.NofollowLinks |
			markdownhtml.NoreferrerLinks |
			markdownhtml.NoopenerLinks |
			markdownhtml.LazyLoadImages,
	})

	rendered := string(markdown.ToHTML(buf.Bytes(), nil, renderer))
	rendered = strings.ReplaceAll(rendered, "<img", "<img class=\"img-responsive\"")
	return template.HTML(rendered)
}
