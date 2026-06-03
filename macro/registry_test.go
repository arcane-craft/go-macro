package macro_test

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func TestRegistryRegisterProvider(t *testing.T) {
	fset := token.NewFileSet()
	src := map[string][]byte{
		"stubs.go": []byte(`package p

//macro: syntax-test
func MacroStub(int) int { panic("macro") }
`),
		"expand.go": []byte(`package p

import (
	"go/ast"
	"github.com/arcane-craft/go-macro/macro"
)

//macro: syntax-test
func TestExpand(ctx macro.CallContext, call *ast.CallExpr) (macro.CallExpandResult, error) {
	return macro.CallExpandResult{}, nil
}
`),
	}
	files, err := macro.ParseProviderFiles(fset, src)
	if err != nil {
		t.Fatal(err)
	}
	r := macro.NewRegistry()
	expand := func(macro.CallContext, *ast.CallExpr) (macro.CallExpandResult, error) {
		return macro.CallExpandResult{}, nil
	}
	if err := r.RegisterProvider("example.com/p", files, expand); err != nil {
		t.Fatal(err)
	}
	sid, ex, ok := r.Lookup("example.com/p", "MacroStub")
	if !ok || sid != "syntax-test" || ex == nil {
		t.Fatalf("Lookup MacroStub: ok=%v sid=%q", ok, sid)
	}
	if !r.HasStub("example.com/p", "MacroStub") {
		t.Fatal("expected HasStub")
	}
}

func TestContextRequiresEnclosingFunc(t *testing.T) {
	_, err := macro.NewCallContext(token.NewFileSet(), nil, nil, nil, nil, "X", "syntax-x", macro.SiteExpr, nil)
	if err == nil {
		t.Fatal("expected error without enclosing func")
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
import ("go/ast"; "github.com/arcane-craft/go-macro/macro")
//macro: syntax-x
func XExpand(ctx macro.CallContext, call *ast.CallExpr) (macro.CallExpandResult, error) {
	return macro.CallExpandResult{}, nil
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
	if e.SyntaxID != "syntax-x" || e.CallExpander != "XExpand" {
		t.Fatalf("entry: %+v", e)
	}
	if len(e.StubNames) != 1 || e.StubNames[0] != "Stub" {
		t.Fatalf("stubs: %v", e.StubNames)
	}
}
