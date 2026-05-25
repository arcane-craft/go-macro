package macro_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func TestContextAccessors(t *testing.T) {
	fset := token.NewFileSet()
	const src = `package p
func Macro(int) int { return 0 }
func f() int { return Macro(1) }
`
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "Macro" {
				call = c
			}
		}
		return true
	})
	fn := f.Decls[0].(*ast.FuncDecl)
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	cfg := &types.Config{Importer: importer.Default()}
	pkg, err := cfg.Check("p", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatal(err)
	}
	ctx, err := macro.NewContext(fset, info, pkg, call, "Macro", "syntax-x", macro.SiteReturn, fn)
	if err != nil {
		t.Fatal(err)
	}
	if ctx.FileSet() != fset || ctx.Call() != call || ctx.StubName() != "Macro" {
		t.Fatal("basic accessors")
	}
	if ctx.SyntaxID() != "syntax-x" || ctx.Site() != macro.SiteReturn || ctx.EnclosingFunc() != fn {
		t.Fatal("site/enclosing")
	}
	if !ctx.MacroPos().IsValid() {
		t.Fatal("MacroPos")
	}
	_ = ctx.Types()
	_ = ctx.Package()
	a := ctx.TempIdent("_v")
	b := ctx.TempIdent("_v")
	if a.Name != "_v1" || b.Name != "_v2" || a.Name == b.Name {
		t.Fatalf("TempIdent: %q %q", a.Name, b.Name)
	}
}

func TestNewContextInvalidEnclosing(t *testing.T) {
	fset := token.NewFileSet()
	_, err := macro.NewContext(fset, nil, nil, nil, "X", "s", macro.SiteExpr, &ast.BlockStmt{})
	if err == nil {
		t.Fatal("expected error for BlockStmt enclosing")
	}
}
