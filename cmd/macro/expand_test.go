package main_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arcane-craft/go-macro/expander"
)

func TestExpandIdempotent(t *testing.T) {
	dir := filepath.Join("..", "..", "examples", "readfile")
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	gen := "readfile_macro_gen.go"
	if _, err := os.Stat(gen); os.IsNotExist(err) {
		if err := expander.ExpandPackages([]string{"."}, nil); err != nil {
			t.Fatal(err)
		}
	}
	before, err := os.ReadFile(gen)
	if err != nil {
		t.Fatal(err)
	}
	if err := expander.ExpandPackages([]string{"."}, nil); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(gen)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("expand is not idempotent")
	}
}
