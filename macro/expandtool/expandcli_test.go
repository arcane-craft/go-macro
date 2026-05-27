package expandtool_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arcane-craft/go-macro/internal/expander"
)

func TestDiscoverProviderLinksReadfile(t *testing.T) {
	root, err := expander.ModuleRoot([]string{"./examples/readfile/..."})
	if err != nil {
		t.Skip("examples module:", err)
	}
	readfileDir := filepath.Join(root, "readfile")
	if _, err := os.Stat(readfileDir); err != nil {
		t.Skip("examples/readfile not found")
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(readfileDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	links, err := expander.DiscoverProviderLinks([]string{"."})
	if err != nil {
		t.Fatal(err)
	}
	if len(links) != 1 || links[0].ExpanderName != "TryExpand" {
		t.Fatalf("links: %+v", links)
	}
}
