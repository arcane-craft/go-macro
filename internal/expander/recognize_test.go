package expander_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/arcane-craft/go-macro/internal/expander"
	"github.com/arcane-craft/go-macro/macro"
)

const providerPath = "example.com/macprov"

func TestRecognizeExplicitImport(t *testing.T) {
	providerSrc := `package macprov

//macro: syntax-test

func MacroStub(int) int { panic("x") }
`
	fset := token.NewFileSet()
	pf, err := parser.ParseFile(fset, "stubs.go", providerSrc, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	reg := macro.NewRegistry()
	if err := reg.RegisterProvider(providerPath, []*ast.File{pf}, "syntax-test", func(macro.Context, *ast.CallExpr) (macro.ExpandResult, error) {
		return macro.ExpandResult{}, nil
	}); err != nil {
		t.Fatal(err)
	}

	src := `package u
import mp "example.com/macprov"
func f() int { return mp.MacroStub(1) }
`
	f, err := parser.ParseFile(fset, "u.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if sel, ok := c.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "MacroStub" {
				call = c
			}
		}
		return true
	})
	if call == nil {
		t.Fatal("no call")
	}
	sel := call.Fun.(*ast.SelectorExpr)
	providerPkg := types.NewPackage(providerPath, "macprov")
	params := types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.Int]))
	results := types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.Int]))
	sig := types.NewSignature(nil, params, results, false)
	stub := types.NewFunc(token.NoPos, providerPkg, "MacroStub", sig)
	importerPkg := types.NewPackage("u", "u")
	mpIdent := sel.X.(*ast.Ident)
	info := &types.Info{
		Uses: map[*ast.Ident]types.Object{
			mpIdent: types.NewPkgName(token.NoPos, importerPkg, mpIdent.Name, providerPkg),
			sel.Sel: stub,
		},
	}
	imports := map[string]string{"mp": providerPath}
	calls, err := expander.RecognizeMacroCalls(f, info, imports, reg)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Fatalf("got %d want 1", len(calls))
	}
}

func TestRecognizeShadowNotMacro(t *testing.T) {
	fset := token.NewFileSet()
	reg := macro.NewRegistry()
	params := types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.Int]))
	results := types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.Int]))
	stubSig := types.NewSignature(nil, params, results, false)
	_ = reg.RegisterProvider(providerPath, []*ast.File{mustParseProvider(t, fset)}, "syntax-test", nil)

	src := `package u
func MacroStub(int) int { return 0 }
func f() int { return MacroStub(1) }
`
	f, _ := parser.ParseFile(fset, "u.go", src, 0)
	var stubIdent *ast.Ident
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "MacroStub" {
				stubIdent = id
			}
		}
		return true
	})
	userPkg := types.NewPackage("u", "u")
	localStub := types.NewFunc(token.NoPos, userPkg, "MacroStub", stubSig)
	info := &types.Info{Uses: map[*ast.Ident]types.Object{stubIdent: localStub}}
	calls, err := expander.RecognizeMacroCalls(f, info, nil, reg)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 0 {
		t.Fatalf("shadow MacroStub must not be macro, got %d calls", len(calls))
	}
}

func TestRecognizeMethodCallNotMacro(t *testing.T) {
	fset := token.NewFileSet()
	reg := macro.NewRegistry()
	_ = reg.RegisterProvider(providerPath, []*ast.File{mustParseProvider(t, fset)}, "syntax-test", nil)

	src := `package u
type S struct{}
func (S) MacroStub(int) int { return 0 }
func f() int {
	var s S
	return s.MacroStub(1)
}
`
	f, _ := parser.ParseFile(fset, "u.go", src, 0)
	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
		Defs: make(map[*ast.Ident]types.Object),
	}
	cfg := &types.Config{Importer: importer.Default()}
	if _, err := cfg.Check("u", fset, []*ast.File{f}, info); err != nil {
		t.Fatal(err)
	}
	calls, err := expander.RecognizeMacroCalls(f, info, nil, reg)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 0 {
		t.Fatalf("method MacroStub must not be macro, got %d calls", len(calls))
	}
}

func mustParseProvider(t *testing.T, fset *token.FileSet) *ast.File {
	t.Helper()
	pf, err := parser.ParseFile(fset, "stubs.go", `package macprov
//macro: syntax-test
func MacroStub(int) int { panic("x") }
`, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	return pf
}
