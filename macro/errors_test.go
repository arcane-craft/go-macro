package macro_test

import (
	"go/token"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func TestErrorAt(t *testing.T) {
	fset := token.NewFileSet()
	file := fset.AddFile("a.go", fset.Base(), 100)
	pos := file.Pos(10)
	err := macro.ErrorAt(fset, pos, "boom %s", "x")
	if err == nil || err.Error() == "" {
		t.Fatal(err)
	}
	if err2 := macro.ErrorAt(fset, token.NoPos, "plain"); err2 == nil {
		t.Fatal("expected plain error")
	}
}
