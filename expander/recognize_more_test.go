package expander_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"testing"

	"github.com/arcane-craft/go-macro/expander"
)

func TestBuildImportMap(t *testing.T) {
	fset := token.NewFileSet()
	src := `package u
import (
	"fmt"
	mp "example.com/macprov"
	. "example.com/dot"
	_ "example.com/blank"
)
`
	f, err := parser.ParseFile(fset, "u.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	m := expander.BuildImportMap(f, "example.com/u")
	if m["fmt"] != "fmt" || m["mp"] != "example.com/macprov" || m["."] != "example.com/dot" {
		t.Fatalf("imports: %v", m)
	}
	if _, ok := m["_"]; ok {
		t.Fatal("blank import should be skipped")
	}
}

func TestBuildImportMapDotAndAlias(t *testing.T) {
	fset := token.NewFileSet()
	src := `package u
import (
	mp "example.com/macprov"
	. "example.com/dot"
)
`
	f, err := parser.ParseFile(fset, "u.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	m := expander.BuildImportMap(f, "example.com/u")
	if m["mp"] != "example.com/macprov" || m["."] != "example.com/dot" {
		t.Fatalf("imports: %v", m)
	}
}

func TestValidateRecognize(t *testing.T) {
	if err := expander.ValidateRecognize(nil); err != nil {
		t.Fatal(err)
	}
	err := expander.ValidateRecognize(fmt.Errorf("boom"))
	if err == nil || err.Error() == "" {
		t.Fatal("expected wrapped error")
	}
}

func TestRecognizeDotImportBareIdent(t *testing.T) {
	providerPath := "example.com/macprov"
	fset := token.NewFileSet()
	reg := setupProviderReg(t, fset, providerPath)

	src := `package u
import . "example.com/macprov"
func f() int { return MacroStub(1) }
`
	f, _ := parser.ParseFile(fset, "u.go", src, 0)
	info := typecheckWithProvider(t, fset, f, providerPath, "MacroStub")
	calls, err := expander.RecognizeMacroCalls(f, info, expander.BuildImportMap(f, "u"), reg)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("dot import: got %d calls", len(calls))
	}
}

func TestRecognizeParenWrappedCall(t *testing.T) {
	providerPath := "example.com/macprov"
	fset := token.NewFileSet()
	reg := setupProviderReg(t, fset, providerPath)

	src := `package u
import mp "example.com/macprov"
func f() int { return (mp.MacroStub)(1) }
`
	f, _ := parser.ParseFile(fset, "u.go", src, 0)
	info := typecheckWithProvider(t, fset, f, providerPath, "MacroStub")
	imports := expander.BuildImportMap(f, "u")
	calls, err := expander.RecognizeMacroCalls(f, info, imports, reg)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("paren call: got %d", len(calls))
	}
}
