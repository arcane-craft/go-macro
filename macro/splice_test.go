package macro_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func TestLegalSpliceTargetsAssign(t *testing.T) {
	f, call := parseSnippet(t, `func f() { x := M(1) }`)
	legal := macro.LegalSpliceTargetsForCall(f, call)
	assertTargets(t, legal, macro.SpliceReplaceAssignRHS, macro.SpliceReplaceAssignStmt)
}

func TestLegalSpliceTargetsReturn(t *testing.T) {
	f, call := parseSnippet(t, `func f() int { return M(1) }`)
	legal := macro.LegalSpliceTargetsForCall(f, call)
	assertTargets(t, legal, macro.SpliceReplaceReturnResults, macro.SpliceReplaceReturnStmt)
}

func TestLegalSpliceTargetsExpr(t *testing.T) {
	f, call := parseSnippet(t, `func f() int { return 1 + M(2) }`)
	legal := macro.LegalSpliceTargetsForCall(f, call)
	assertTargets(t, legal, macro.SpliceReplaceCallExpr)
}

func TestValidateCallExpandResultZeroTarget(t *testing.T) {
	f, call := parseSnippet(t, `func f() { M(1) }`)
	err := macro.ValidateCallExpandResultForCall(f, call, macro.CallExpandResult{})
	if err == nil || !strings.Contains(err.Error(), "Target is required") {
		t.Fatalf("got %v", err)
	}
}

func TestValidateCallExpandResultWrongTargetAtReturn(t *testing.T) {
	f, call := parseSnippet(t, `func f() int { return M(1) }`)
	err := macro.ValidateCallExpandResultForCall(f, call, macro.CallExpandResult{
		Target: macro.SpliceReplaceCallExpr,
		Expr:   ast.NewIdent("1"),
	})
	if err == nil || !strings.Contains(err.Error(), "legal targets") {
		t.Fatalf("got %v", err)
	}
}

func parseSnippet(t *testing.T, body string) (*ast.File, *ast.CallExpr) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "t.go", "package p\n"+body, 0)
	if err != nil {
		t.Fatal(err)
	}
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "M" {
				call = c
			}
		}
		return true
	})
	if call == nil {
		t.Fatal("no call M")
	}
	return f, call
}

func assertTargets(t *testing.T, got []macro.SpliceTarget, want ...macro.SpliceTarget) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}
}
