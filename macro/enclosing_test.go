package macro_test

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

func TestEnclosingSignatureMatchesEnclosingFunc(t *testing.T) {
	fset := token.NewFileSet()
	const src = `package p
func f() (int, error) {
	x := g()
	return x, nil
}
func g() int { return 1 }
`
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "g" {
				call = c
			}
		}
		return true
	})
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	cfg := &types.Config{Importer: importer.Default()}
	_, err = cfg.Check("p", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatal(err)
	}

	site, err := expander.ResolveSite(fset, f, call)
	if err != nil {
		t.Fatal(err)
	}
	ctx := macro.NewContext(fset, info)
	sig, err := macro.EnclosingSignature(ctx, site)
	if err != nil {
		t.Fatal(err)
	}
	results, err := macro.EnclosingResults(ctx, site)
	if err != nil {
		t.Fatal(err)
	}
	if sig.Results().Len() != 2 {
		t.Fatalf("results len=%d", sig.Results().Len())
	}
	if results.Len() != 2 {
		t.Fatalf("tuple len=%d", results.Len())
	}

	fnObj, err := expander.FuncAtPos(f, info, site.MacroPos())
	if err != nil {
		t.Fatal(err)
	}
	if fnObj.Name() != "f" {
		t.Fatalf("FuncAtPos name=%q", fnObj.Name())
	}
}
