package expander

import (
	"fmt"
	"go/ast"

	"github.com/arcane-craft/go-macro/macro"
)

// ApplyExpandResult splices expand result into the file AST.
func ApplyExpandResult(file *ast.File, call *ast.CallExpr, site macro.CallSiteKind, result macro.ExpandResult) error {
	switch site {
	case macro.SiteAssign:
		if len(result.Stmts) == 0 {
			return fmt.Errorf("macro: SiteAssign requires Stmts")
		}
		block, stmt, ok := findEnclosingBlockStmt(file, call)
		if !ok {
			return fmt.Errorf("macro: cannot find AssignStmt for call")
		}
		assign, ok := stmt.(*ast.AssignStmt)
		if !ok {
			return fmt.Errorf("macro: expected AssignStmt, got %T", stmt)
		}
		replaceStmtInBlock(block, assign, result.Stmts)
	case macro.SiteReturn:
		if len(result.Stmts) > 0 {
			block, stmt, ok := findEnclosingBlockStmt(file, call)
			if !ok {
				return fmt.Errorf("macro: cannot find ReturnStmt for call")
			}
			ret, ok := stmt.(*ast.ReturnStmt)
			if !ok {
				return fmt.Errorf("macro: expected ReturnStmt, got %T", stmt)
			}
			replaceStmtInBlock(block, ret, result.Stmts)
		} else if len(result.Exprs) > 0 {
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
		} else {
			return fmt.Errorf("macro: SiteReturn requires Stmts or Exprs")
		}
	case macro.SiteStmt:
		if len(result.Stmts) == 0 {
			return fmt.Errorf("macro: SiteStmt requires Stmts")
		}
		block, stmt, ok := findEnclosingBlockStmt(file, call)
		if !ok {
			return fmt.Errorf("macro: cannot find ExprStmt for call")
		}
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			return fmt.Errorf("macro: expected ExprStmt, got %T", stmt)
		}
		replaceStmtInBlock(block, exprStmt, result.Stmts)
	case macro.SiteExpr:
		if result.Expr == nil {
			return fmt.Errorf("macro: SiteExpr requires Expr")
		}
		if !replaceCallExpr(file, call, result.Expr) {
			return fmt.Errorf("macro: cannot replace CallExpr")
		}
	default:
		return fmt.Errorf("macro: unknown site %d", site)
	}
	return nil
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
			// Do not append into block.List[:i] — it shares the backing array with block.List
			// and would overwrite trailing statements (e.g. defer/return after the macro call).
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
				if rhs == call {
					parent.Rhs[i] = expr
					replaced = true
					return false
				}
			}
		case *ast.ReturnStmt:
			for i, r := range parent.Results {
				if r == call {
					parent.Results[i] = expr
					replaced = true
					return false
				}
			}
		case *ast.ExprStmt:
			if parent.X == call {
				parent.X = expr
				replaced = true
				return false
			}
		case *ast.BinaryExpr:
			if parent.X == call {
				parent.X = expr
				replaced = true
				return false
			}
			if parent.Y == call {
				parent.Y = expr
				replaced = true
				return false
			}
		case *ast.UnaryExpr:
			if parent.X == call {
				parent.X = expr
				replaced = true
				return false
			}
		case *ast.CallExpr:
			for i, arg := range parent.Args {
				if arg == call {
					parent.Args[i] = expr
					replaced = true
					return false
				}
			}
		case *ast.CompositeLit:
			for i, elt := range parent.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok && kv.Value == call {
					kv.Value = expr
					replaced = true
					return false
				}
				if elt == call {
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
