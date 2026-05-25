package macro

import (
	"go/ast"
	"go/token"
)

// StampStmtPos sets token positions on macro-generated statements so generated
// //line directives map to the macro call site (ctx.MacroPos()).
// Nodes that already carry a valid position are left unchanged.
func StampStmtPos(pos token.Pos, stmts []ast.Stmt) {
	if !pos.IsValid() {
		return
	}
	for _, s := range stmts {
		stampStmt(pos, s)
	}
}

func stampStmt(pos token.Pos, s ast.Stmt) {
	switch s := s.(type) {
	case *ast.AssignStmt:
		if !s.TokPos.IsValid() {
			s.TokPos = pos
		}
		// go/ast: AssignStmt.Pos reports Lhs[0].Pos(), not TokPos.
		if len(s.Lhs) > 0 {
			stampExprPos(pos, s.Lhs[0])
		}
	case *ast.IfStmt:
		if !s.If.IsValid() {
			s.If = pos
		}
	case *ast.ReturnStmt:
		if !s.Return.IsValid() {
			s.Return = pos
		}
	case *ast.ExprStmt:
		if p := s.X.Pos(); !p.IsValid() {
			stampExprPos(pos, s.X)
		}
	}
}

func stampExprPos(pos token.Pos, e ast.Expr) {
	switch e := e.(type) {
	case *ast.Ident:
		if !e.NamePos.IsValid() {
			e.NamePos = pos
		}
	case *ast.CallExpr:
		if !e.Lparen.IsValid() {
			e.Lparen = pos
		}
	}
}
