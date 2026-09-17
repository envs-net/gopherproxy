package gopherproxy

import (
	"strings"
	"testing"
)

func TestRenderMdSkipsRawHTMLAndUnsafeLinks(t *testing.T) {
	t.Parallel()

	rendered := string(renderMd(strings.NewReader(`# Test

<script>alert("xss")</script>

[unsafe](javascript:alert(1))
`)))

	lower := strings.ToLower(rendered)
	if strings.Contains(lower, "<script") {
		t.Fatalf("raw script survived markdown rendering: %s", rendered)
	}
	if strings.Contains(lower, "javascript:") {
		t.Fatalf("unsafe link survived markdown rendering: %s", rendered)
	}
}
