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
		return macro.ExpandResult{Expr: ast.NewIdent("1")}, nil
	}
	r.RegisterSyntax("syntax-a", expand)
	r.RegisterStub("StubA", "syntax-a")
	sid, ex, ok := r.Lookup("StubA")
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
	if err := r.RegisterProvider("p", bad, "syntax-x", func(macro.Context, *ast.CallExpr) (macro.ExpandResult, error) {
		return macro.ExpandResult{}, nil
	}); err == nil {
		t.Fatal("want missing directive error")
	}
	if err := r.RegisterProvider("p", bad, "syntax-x", nil); err == nil {
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
