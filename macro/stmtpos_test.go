package macro_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func TestStampStmtPosAssign(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "x.go", `package p
func f() { x := g() }`, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	call := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.AssignStmt).Rhs[0].(*ast.CallExpr)
	macroPos := call.Pos()

	stmts := []ast.Stmt{
		&ast.AssignStmt{
			Tok: token.DEFINE,
			Lhs: []ast.Expr{ast.NewIdent("_v")},
			Rhs: []ast.Expr{call},
		},
	}
	macro.StampStmtPos(macroPos, stmts)
	if !stmts[0].Pos().IsValid() {
		t.Fatal("expected valid Pos on stamped AssignStmt")
	}
	if got := fset.Position(stmts[0].Pos()).Line; got != 2 {
		t.Fatalf("line = %d, want 2", got)
	}
}
