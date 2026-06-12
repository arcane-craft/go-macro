package macro_test

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func formatExpr(t *testing.T, e ast.Expr) string {
	t.Helper()
	fset := token.NewFileSet()
	var buf bytes.Buffer
	_ = (&printer.Config{Mode: printer.RawFormat, Tabwidth: 8}).Fprint(&buf, fset, e)
	return buf.String()
}

func formatStmts(t *testing.T, stmts []ast.Stmt) string {
	t.Helper()
	fset := token.NewFileSet()
	var buf bytes.Buffer
	cfg := &printer.Config{Mode: printer.RawFormat, Tabwidth: 8}
	for _, s := range stmts {
		_ = cfg.Fprint(&buf, fset, s)
	}
	return buf.String()
}

func TestQuoteExpr(t *testing.T) {
	syn, err := macro.Quote(`1 + 2`, nil)
	if err != nil {
		t.Fatal(err)
	}
	e, err := syn.ToExpr()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(formatExpr(t, e), "1 + 2") {
		t.Fatalf("got %s", formatExpr(t, e))
	}
}

func TestQuoteExprs(t *testing.T) {
	syn, err := macro.Quote(`f(#a, #b)`, map[string]macro.Syntax{
		"a": macro.WrapExpr(&ast.BasicLit{Kind: token.INT, Value: "1"}),
		"b": macro.WrapExpr(&ast.BasicLit{Kind: token.INT, Value: "2"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	e, err := syn.ToExpr()
	if err != nil {
		t.Fatal(err)
	}
	call, ok := e.(*ast.CallExpr)
	if !ok || len(call.Args) != 2 {
		t.Fatalf("got %#v", e)
	}
}

func TestQuoteStmts(t *testing.T) {
	syn, err := macro.Quote(`x := 1`, nil)
	if err != nil {
		t.Fatal(err)
	}
	stmts, err := syn.ToStmts()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(formatStmts(t, stmts), "x := 1") {
		t.Fatalf("got %s", formatStmts(t, stmts))
	}
}

func TestQuoteDecls(t *testing.T) {
	syn, err := macro.Quote(`type T int`, nil)
	if err != nil {
		t.Fatal(err)
	}
	decls, err := syn.ToDecls()
	if err != nil || len(decls) != 1 {
		t.Fatalf("decls=%v err=%v", decls, err)
	}
}

func TestQuoteHoleBind(t *testing.T) {
	inner := macro.WrapExpr(ast.NewIdent("v"))
	syn, err := macro.Quote(`#inner + 1`, map[string]macro.Syntax{"inner": inner})
	if err != nil {
		t.Fatal(err)
	}
	e, err := syn.ToExpr()
	if err != nil {
		t.Fatal(err)
	}
	out := formatExpr(t, e)
	if !strings.Contains(out, "v") || !strings.Contains(out, "+ 1") {
		t.Fatalf("got %s", out)
	}
}
