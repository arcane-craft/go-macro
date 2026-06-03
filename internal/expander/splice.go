package expander

import (
	"fmt"
	"go/ast"

	"github.com/arcane-craft/go-macro/macro"
)

// ApplyExpandResult splices expand result into the file AST.
func ApplyExpandResult(file *ast.File, call *ast.CallExpr, result macro.CallExpandResult) error {
	if err := macro.ValidateCallExpandResultForCall(file, call, result); err != nil {
		return err
	}
	switch result.Target {
	case macro.SpliceReplaceAssignStmt:
		block, stmt, ok := findEnclosingBlockStmt(file, call)
		if !ok {
			return fmt.Errorf("macro: cannot find AssignStmt for call")
		}
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok {
			return fmt.Errorf("macro: expected AssignStmt, got %T", stmt)
		}
		replaceStmtInBlock(block, assign, result.Stmts)
	case macro.SpliceReplaceAssignRHS:
		if err := replaceAssignRHS(file, call, result.Expr); err != nil {
			return err
		}
	case macro.SpliceReplaceReturnStmt:
		block, stmt, ok := findEnclosingBlockStmt(file, call)
		if !ok {
			return fmt.Errorf("macro: cannot find ReturnStmt for call")
		}
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			return fmt.Errorf("macro: expected ReturnStmt, got %T", stmt)
		}
		replaceStmtInBlock(block, ret, result.Stmts)
	case macro.SpliceReplaceReturnResults:
		block, stmt, ok := findEnclosingBlockStmt(file, call)
		if !ok {
			return fmt.Errorf("macro: cannot find ReturnStmt for call")
		}
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok {
			return fmt.Errorf("macro: expected ReturnStmt, got %T", stmt)
		}
		ret.Results = result.Exprs
		_ = block
	case macro.SpliceReplaceExprStmt:
		block, stmt, ok := findEnclosingBlockStmt(file, call)
		if !ok {
			return fmt.Errorf("macro: cannot find ExprStmt for call")
		}
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			return fmt.Errorf("macro: expected ExprStmt, got %T", stmt)
		}
		replaceStmtInBlock(block, exprStmt, result.Stmts)
	case macro.SpliceReplaceCallExpr:
		if !replaceCallExpr(file, call, result.Expr) {
			return fmt.Errorf("macro: cannot replace CallExpr")
		}
	default:
		return fmt.Errorf("macro: unknown splice target %v", result.Target)
	}
	return nil
}

func replaceAssignRHS(file *ast.File, call *ast.CallExpr, expr ast.Expr) error {
	var assign *ast.AssignStmt
	ast.Inspect(file, func(n ast.Node) bool {
		a, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, rhs := range a.Rhs {
			if rhs == call || unwrapParenExpr(rhs) == call {
				assign = a
				return false
			}
		}
		return true
	})
	if assign == nil {
		return fmt.Errorf("macro: cannot find AssignStmt RHS slot for call")
	}
	for i, rhs := range assign.Rhs {
		if rhs == call {
			assign.Rhs[i] = expr
			return nil
		}
		if unwrapParenExpr(rhs) == call {
			assign.Rhs[i] = expr
			return nil
		}
	}
	return fmt.Errorf("macro: cannot replace macro call in AssignStmt.Rhs")
}

func unwrapParenExpr(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}

func findEnclosingBlockStmt(file *ast.File, call *ast.CallExpr) (*ast.BlockStmt, ast.Stmt, bool) {
	var best ast.Stmt
	var bestBlock *ast.BlockStmt
	bestLen := -1
	ast.Inspect(file, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for _, s := range block.List {
			if call.Pos() >= s.Pos() && call.End() <= s.End() {
				span := int(s.End() - s.Pos())
				if span > bestLen {
					bestLen = span
					best = s
					bestBlock = block
				}
			}
		}
		return true
	})
	return bestBlock, best, best != nil && bestBlock != nil
}

func replaceStmtInBlock(block *ast.BlockStmt, old ast.Stmt, newStmts []ast.Stmt) {
	for i, s := range block.List {
		if s == old {
			prefix := append([]ast.Stmt(nil), block.List[:i]...)
			block.List = append(append(prefix, newStmts...), block.List[i+1:]...)
			return
		}
	}
}

func replaceCallExpr(file *ast.File, call *ast.CallExpr, expr ast.Expr) bool {
	replaced := false
	ast.Inspect(file, func(n ast.Node) bool {
		if replaced {
			return false
		}
		switch parent := n.(type) {
		case *ast.AssignStmt:
			for i, rhs := range parent.Rhs {
				if rhs == call || unwrapParenExpr(rhs) == call {
					parent.Rhs[i] = expr
					replaced = true
					return false
				}
			}
		case *ast.ReturnStmt:
			for i, r := range parent.Results {
				if r == call || unwrapParenExpr(r) == call {
					parent.Results[i] = expr
					replaced = true
					return false
				}
			}
		case *ast.ExprStmt:
			if parent.X == call || unwrapParenExpr(parent.X) == call {
				parent.X = expr
				replaced = true
				return false
			}
		case *ast.BinaryExpr:
			if parent.X == call || unwrapParenExpr(parent.X) == call {
				parent.X = expr
				replaced = true
				return false
			}
			if parent.Y == call || unwrapParenExpr(parent.Y) == call {
				parent.Y = expr
				replaced = true
				return false
			}
		case *ast.UnaryExpr:
			if parent.X == call || unwrapParenExpr(parent.X) == call {
				parent.X = expr
				replaced = true
				return false
			}
		case *ast.CallExpr:
			for i, arg := range parent.Args {
				if arg == call || unwrapParenExpr(arg) == call {
					parent.Args[i] = expr
					replaced = true
					return false
				}
			}
		case *ast.CompositeLit:
			for i, elt := range parent.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok && (kv.Value == call || unwrapParenExpr(kv.Value) == call) {
					kv.Value = expr
					replaced = true
					return false
				}
				if elt == call || unwrapParenExpr(elt) == call {
					parent.Elts[i] = expr
					replaced = true
					return false
				}
			}
		}
		return true
	})
	return replaced
}
