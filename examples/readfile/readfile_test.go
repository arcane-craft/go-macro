package readfile_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileGenGolden(t *testing.T) {
	data, err := os.ReadFile("readfile_macro_gen.go")
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile(filepath.Join("testdata", "readfile_macro_gen.golden"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != strings.TrimSpace(string(golden)) {
		t.Fatal("readfile_macro_gen.go does not match testdata golden")
	}
}
