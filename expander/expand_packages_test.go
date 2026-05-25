package expander

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arcane-craft/go-macro/inline"
	"github.com/arcane-craft/go-macro/try"
)

func TestExpandPackagesReadfileGen(t *testing.T) {
	root, _ := filepath.Abs("..")
	genPath := filepath.Join(root, "examples", "readfile", "readfile_macro_gen.go")
	_ = os.Remove(genPath)
	providers := []Provider{
		{ImportPath: "github.com/arcane-craft/go-macro/inline", SyntaxID: "syntax-inline", Expand: inline.InlineExpand},
		{ImportPath: "github.com/arcane-craft/go-macro/try", SyntaxID: "syntax-try", Expand: try.TryExpand},
	}
	if err := ExpandPackages([]string{"github.com/arcane-craft/go-macro/examples/readfile"}, providers); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(genPath)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if strings.Count(s, "_err1 :=") > 1 || strings.Count(s, "file := _v2") > 1 {
		t.Fatalf("duplicate expansion in gen:\n%s", s)
	}
	if !strings.Contains(s, "defer file.Close()") {
		t.Fatalf("missing defer in gen:\n%s", s)
	}
	if strings.Contains(s, "go-macro/try") {
		t.Fatalf("gen must not import try after expand:\n%s", s)
	}

	goldenPath := filepath.Join(root, "examples", "readfile", "testdata", "readfile_macro_gen.golden")
	golden, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(golden)) != strings.TrimSpace(s) {
		t.Fatalf("gen does not match golden\n--- got ---\n%s\n--- want ---\n%s", s, golden)
	}
}
