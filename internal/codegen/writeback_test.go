package codegen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestStmtLineFallsBackToSubtree(t *testing.T) {
	fset := token.NewFileSet()
	const src = `package p
func f() {
	x := g()
}
`
	f, err := parser.ParseFile(fset, "x.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	stmt := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt)
	// Simulate macro-generated assign: no TokPos, temp lhs, rhs keeps parse position.
	stmt.TokPos = 0
	stmt.Lhs = []ast.Expr{ast.NewIdent("_v")}
	if line := stmtLine(fset, stmt); line != 3 {
		t.Fatalf("stmtLine = %d, want 3 (from rhs g())", line)
	}
}
