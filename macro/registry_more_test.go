package macro_test

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func TestRegistryRegisterStubAndSyntax(t *testing.T) {
	r := macro.NewRegistry()
	expand := noopExpander
	r.RegisterExpander("syntax-a", expand)
	r.RegisterStub("example.com/p", "StubA", "syntax-a")
	sid, ok := r.SyntaxIDForStub("example.com/p", "StubA")
	if !ok || sid != "syntax-a" {
		t.Fatalf("SyntaxIDForStub: ok=%v sid=%q", ok, sid)
	}
	ex, ok := r.LookupExpander(sid)
	if !ok || ex == nil {
		t.Fatalf("LookupExpander: ok=%v", ok)
	}
	stubs := r.ProviderStubs("missing")
	if stubs != nil {
		t.Fatalf("ProviderStubs missing: %v", stubs)
	}
}

func TestRegisterProviderSourcesErrors(t *testing.T) {
	r := macro.NewRegistry()
	fset := token.NewFileSet()
	bad, _ := macro.ParseProviderFiles(fset, map[string][]byte{
		"p.go": []byte(`package p
func X() {}
`),
	})
	if err := r.RegisterProviderSources("p", bad); err == nil {
		t.Fatal("want missing directive error")
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
import "github.com/arcane-craft/go-macro/macro"
//macro: syntax-test
func XExpand(ctx macro.Context, site macro.Syntax) (macro.Syntax, error) {
	return nil, nil
}
`),
		})
		if err != nil {
			t.Fatal(err)
		}
		return files
	}
	r := macro.NewRegistry()
	expA := noopExpander
	expB := func(macro.Context, macro.Syntax) (macro.Syntax, error) { return nil, nil }
	if err := r.RegisterProviderSources("a.com/p", makeFiles("p", "Macro")); err != nil {
		t.Fatal(err)
	}
	if err := r.RegisterProviderSources("b.com/p", makeFiles("p", "Macro")); err != nil {
		t.Fatal(err)
	}
	r.RegisterExpander("syntax-test", expA)
	sidA, ok := r.SyntaxIDForStub("a.com/p", "Macro")
	if !ok {
		t.Fatal("a sid")
	}
	_, ok = r.LookupExpander(sidA)
	if !ok {
		t.Fatal("a expander")
	}
	r.RegisterExpander("syntax-test", expB)
	sidB, ok := r.SyntaxIDForStub("b.com/p", "Macro")
	if !ok {
		t.Fatal("b sid")
	}
	_, ok = r.LookupExpander(sidB)
	if !ok {
		t.Fatal("b expander")
	}
}
