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
func TestExpand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error) {
	return macro.ExpandResult{}, nil
}
`),
	}
	files, err := macro.ParseProviderFiles(fset, src)
	if err != nil {
		t.Fatal(err)
	}
	r := macro.NewRegistry()
	expand := func(macro.Context, *ast.CallExpr) (macro.ExpandResult, error) {
		return macro.ExpandResult{}, nil
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
	_, err := macro.NewContext(token.NewFileSet(), nil, nil, nil, nil, "X", "syntax-x", macro.SiteExpr, nil)
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
func XExpand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error) {
	return macro.ExpandResult{}, nil
}
`),
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := macro.ScanProviderFiles(files)
	if err != nil {
		t.Fatal(err)
	}
	if info.SyntaxID != "syntax-x" || info.ExpanderName != "XExpand" {
		t.Fatalf("info: %+v", info)
	}
	if len(info.StubNames) != 1 || info.StubNames[0] != "Stub" {
		t.Fatalf("stubs: %v", info.StubNames)
	}
}
