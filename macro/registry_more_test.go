package macro_test

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func TestRegistryRegisterStubAndSyntax(t *testing.T) {
	r := macro.NewRegistry()
	expand := func(macro.Context, *ast.CallExpr) (macro.ExpandResult, error) {
		return macro.ExpandResult{Target: macro.SpliceReplaceCallExpr, Expr: ast.NewIdent("1")}, nil
	}
	r.RegisterSyntax("syntax-a", expand)
	r.RegisterImportExpander("example.com/p", expand)
	r.RegisterStub("example.com/p", "StubA", "syntax-a")
	sid, ex, ok := r.Lookup("example.com/p", "StubA")
	if !ok || sid != "syntax-a" || ex == nil {
		t.Fatalf("Lookup: ok=%v sid=%q", ok, sid)
	}
	stubs := r.ProviderStubs("missing")
	if stubs != nil {
		t.Fatalf("ProviderStubs missing: %v", stubs)
	}
}

func TestRegisterProviderErrors(t *testing.T) {
	r := macro.NewRegistry()
	fset := token.NewFileSet()
	bad, _ := macro.ParseProviderFiles(fset, map[string][]byte{
		"p.go": []byte(`package p
func X() {}
`),
	})
	if err := r.RegisterProvider("p", bad, func(macro.Context, *ast.CallExpr) (macro.ExpandResult, error) {
		return macro.ExpandResult{}, nil
	}); err == nil {
		t.Fatal("want missing directive error")
	}
	if err := r.RegisterProvider("p", bad, nil); err == nil {
		t.Fatal("want nil expander error")
	}
}

func TestParseProviderFilesError(t *testing.T) {
	fset := token.NewFileSet()
	_, err := macro.ParseProviderFiles(fset, map[string][]byte{"bad.go": []byte("not go")})
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestHasStubUnknownProvider(t *testing.T) {
	r := macro.NewRegistry()
	if r.HasStub("example.com/x", "Y") {
		t.Fatal("unexpected HasStub")
	}
}

func TestLookupDifferentProvidersSameStubName(t *testing.T) {
	fset := token.NewFileSet()
	makeFiles := func(pkg, stub string) []*ast.File {
		files, err := macro.ParseProviderFiles(fset, map[string][]byte{
			"s.go": []byte("package " + pkg + "\n//macro: syntax-test\nfunc " + stub + "() { panic(\"x\") }\n"),
			"e.go": []byte(`package ` + pkg + `
import ("go/ast"; "github.com/arcane-craft/go-macro/macro")
//macro: syntax-test
func XExpand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error) {
	return macro.ExpandResult{}, nil
}
`),
		})
		if err != nil {
			t.Fatal(err)
		}
		return files
	}
	r := macro.NewRegistry()
	expA := func(macro.Context, *ast.CallExpr) (macro.ExpandResult, error) {
		return macro.ExpandResult{Target: macro.SpliceReplaceCallExpr, Expr: ast.NewIdent("a")}, nil
	}
	expB := func(macro.Context, *ast.CallExpr) (macro.ExpandResult, error) {
		return macro.ExpandResult{Target: macro.SpliceReplaceCallExpr, Expr: ast.NewIdent("b")}, nil
	}
	if err := r.RegisterProvider("a.com/p", makeFiles("p", "Macro"), expA); err != nil {
		t.Fatal(err)
	}
	if err := r.RegisterProvider("b.com/p", makeFiles("p", "Macro"), expB); err != nil {
		t.Fatal(err)
	}
	_, exA, ok := r.Lookup("a.com/p", "Macro")
	if !ok || exA == nil {
		t.Fatal("a")
	}
	_, exB, ok := r.Lookup("b.com/p", "Macro")
	if !ok || exB == nil {
		t.Fatal("b")
	}
}
