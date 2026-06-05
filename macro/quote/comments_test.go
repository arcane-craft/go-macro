package quote

import (
	"strings"
	"testing"
)

func TestCommentPreservedInParsedFile(t *testing.T) {
	pt, err := parseTemplate(`@stmts{
		// hello
		x := 1
	}`)
	if err != nil {
		t.Fatal(err)
	}
	src, err := synthesize(pt.Kind, pt.Body)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := parseSynthesized(pt.Kind, src)
	if err != nil {
		t.Fatal(err)
	}
	out, err := formatParsedFile(parsed)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "hello") {
		t.Fatalf("comment missing:\n%s", out)
	}
}
