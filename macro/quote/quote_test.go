package quote_test

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
	"testing"

	"github.com/arcane-craft/go-macro/macro/quote"
)

func formatStmts(stmts []ast.Stmt) string {
	fset := token.NewFileSet()
	var buf bytes.Buffer
	cfg := &printer.Config{Mode: printer.RawFormat, Tabwidth: 8}
	for _, s := range stmts {
		_ = cfg.Fprint(&buf, fset, s)
	}
	return buf.String()
}

func formatExpr(e ast.Expr) string {
	fset := token.NewFileSet()
	var buf bytes.Buffer
	cfg := &printer.Config{Mode: printer.RawFormat, Tabwidth: 8}
	_ = cfg.Fprint(&buf, fset, e)
	return buf.String()
}

func TestGoldenExpr(t *testing.T) {
	t.Parallel()
	e, err := quote.Expr(`1 + 2`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(formatExpr(e), "1 + 2") {
		t.Fatalf("got %s", formatExpr(e))
	}
}

func TestGoldenExprs(t *testing.T) {
	t.Parallel()
	exprs, err := quote.Exprs(`1, 2`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(exprs) != 2 {
		t.Fatalf("len=%d", len(exprs))
	}
}

func TestGoldenStmts(t *testing.T) {
	t.Parallel()
	stmts, err := quote.Stmts(`x := 1`, nil)
	if err != nil {
		t.Fatal(err)
	}
	out := formatStmts(stmts)
	if !strings.Contains(out, "x := 1") {
		t.Fatalf("got %s", out)
	}
}

func TestGoldenDecls(t *testing.T) {
	t.Parallel()
	decls, err := quote.Decls(`type T int`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(decls) != 1 {
		t.Fatalf("len=%d", len(decls))
	}
}

func TestQuoteGeneric(t *testing.T) {
	t.Parallel()
	nodes, err := quote.Quote(`@stmts{ y := 2 }`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 1 {
		t.Fatalf("len=%d", len(nodes))
	}
}
