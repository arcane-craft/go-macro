package quote_test

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/arcane-craft/go-macro/macro/quote"
)

func TestStringBindingIdent(t *testing.T) {
	t.Parallel()
	e, err := quote.Expr(`#x`, map[string]any{"x": "file"})
	if err != nil {
		t.Fatal(err)
	}
	id, ok := e.(*ast.Ident)
	if !ok || id.Name != "file" {
		t.Fatalf("got %#v", e)
	}
}

func TestStmtListFastPath(t *testing.T) {
	t.Parallel()
	stmts := []ast.Stmt{
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("a")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "1"}},
		},
		&ast.AssignStmt{
			Lhs: []ast.Expr{ast.NewIdent("b")},
			Tok: token.DEFINE,
			Rhs: []ast.Expr{&ast.BasicLit{Kind: token.INT, Value: "2"}},
		},
	}
	out, err := quote.Stmts(`#block`, map[string]any{"block": stmts})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("len=%d", len(out))
	}
}

func TestNestedQuote(t *testing.T) {
	t.Parallel()
	inner, err := quote.Expr(`1`, nil)
	if err != nil {
		t.Fatal(err)
	}
	e, err := quote.Expr(`#inner`, map[string]any{"inner": inner})
	if err != nil {
		t.Fatal(err)
	}
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Value != "1" {
		t.Fatalf("got %#v", e)
	}
}

func TestExprsWithHole(t *testing.T) {
	t.Parallel()
	v := ast.NewIdent("v")
	exprs, err := quote.Exprs(`#v, nil`, map[string]any{"v": v})
	if err != nil {
		t.Fatal(err)
	}
	if len(exprs) != 2 {
		t.Fatalf("len=%d", len(exprs))
	}
	if id, ok := exprs[0].(*ast.Ident); !ok || id.Name != "v" {
		t.Fatalf("first=%#v", exprs[0])
	}
}
