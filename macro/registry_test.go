package macro_test

import (
	"testing"
	"go/token"

	"github.com/arcane-craft/go-macro/macro"
)

func noopExpander(macro.Context, macro.Syntax) (macro.Syntax, error) {
	return nil, nil
}

func TestRegistryRegisterProviderSources(t *testing.T) {
	fset := token.NewFileSet()
	src := map[string][]byte{
		"stubs.go": []byte(`package p

//macro: syntax-test
func MacroStub(int) int { panic("macro") }
`),
		"expand.go": []byte(`package p

import "github.com/arcane-craft/go-macro/macro"

//macro: syntax-test
func TestExpand(ctx macro.Context, site macro.Syntax) (macro.Syntax, error) {
	return nil, nil
}
`),
	}
	files, err := macro.ParseProviderFiles(fset, src)
	if err != nil {
		t.Fatal(err)
	}
	r := macro.NewRegistry()
	if err := r.RegisterProviderSources("example.com/p", files); err != nil {
		t.Fatal(err)
	}
	r.RegisterExpander("syntax-test", noopExpander)
	sid, ok := r.SyntaxIDForStub("example.com/p", "MacroStub")
	if !ok || sid != "syntax-test" {
		t.Fatalf("SyntaxIDForStub MacroStub: ok=%v sid=%q", ok, sid)
	}
	exp, ok := r.LookupExpander(sid)
	if !ok || exp == nil {
		t.Fatalf("LookupExpander: ok=%v", ok)
	}
	if !r.HasStub("example.com/p", "MacroStub") {
		t.Fatal("expected HasStub")
	}
}

func TestScanProviderFiles(t *testing.T) {
	fset := token.NewFileSet()
	files, err := macro.ParseProviderFiles(fset, map[string][]byte{
		"stubs.go": []byte(`package p
//macro: syntax-x
func Stub() { panic("x") }
`),
		"expand.go": []byte(`package p
import "github.com/arcane-craft/go-macro/macro"
//macro: syntax-x
func XExpand(ctx macro.Context, site macro.Syntax) (macro.Syntax, error) {
	return nil, nil
}
`),
	})
	if err != nil {
		t.Fatal(err)
	}
	scan, err := macro.ScanProviderFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	if len(scan.Entries) != 1 {
		t.Fatalf("entries: %+v", scan.Entries)
	}
	e := scan.Entries[0]
	if e.SyntaxID != "syntax-x" || e.Expander != "XExpand" {
		t.Fatalf("entry: %+v", e)
	}
	if len(e.StubNames) != 1 || e.StubNames[0] != "Stub" {
		t.Fatalf("stubs: %v", e.StubNames)
	}
}
