package mactest

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"strings"
)

// FormatStmts prints statements the way generated Go would read.
func FormatStmts(fset *token.FileSet, stmts []ast.Stmt) string {
	var buf bytes.Buffer
	cfg := &printer.Config{Mode: printer.RawFormat, Tabwidth: 8}
	for _, s := range stmts {
		_ = cfg.Fprint(&buf, fset, s)
	}
	return buf.String()
}

// FormatExpr prints a single expression.
func FormatExpr(fset *token.FileSet, expr ast.Expr) string {
	var buf bytes.Buffer
	cfg := &printer.Config{Mode: printer.RawFormat, Tabwidth: 8}
	_ = cfg.Fprint(&buf, fset, expr)
	return buf.String()
}

// MustContainAll fails if s does not contain every substring.
func MustContainAll(t testingT, s string, subs ...string) {
	t.Helper()
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			t.Errorf("output missing %q\n--- full ---\n%s", sub, s)
		}
	}
}

type testingT interface {
	Helper()
	Errorf(string, ...any)
}

// IdentName returns the name of an *ast.Ident or "".
func IdentName(e ast.Expr) string {
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}
