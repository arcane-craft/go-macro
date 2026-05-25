package expander

import (
	"go/ast"

	"github.com/arcane-craft/go-macro/macro"
)

// classifySiteInFile determines call site kind within file.
func classifySiteInFile(file *ast.File, call *ast.CallExpr) macro.CallSiteKind {
	site := macro.SiteExpr
	ast.Inspect(file, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, rhs := range stmt.Rhs {
				if rhs == call {
					site = macro.SiteAssign
				}
			}
		case *ast.ReturnStmt:
			for _, r := range stmt.Results {
				if r == call {
					site = macro.SiteReturn
				}
			}
		case *ast.ExprStmt:
			if stmt.X == call {
				site = macro.SiteStmt
			}
		}
		return true
	})
	return site
}

// enclosingFuncInFile returns the innermost func containing call.
func enclosingFuncInFile(file *ast.File, call *ast.CallExpr) ast.Node {
	var enc ast.Node
	bestSpan := -1
	ast.Inspect(file, func(n ast.Node) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			if fn.Body != nil && call.Pos() >= fn.Pos() && call.End() <= fn.End() {
				span := int(fn.End() - fn.Pos())
				if span > bestSpan {
					bestSpan = span
					enc = fn
				}
			}
		case *ast.FuncLit:
			if fn.Body != nil && call.Pos() >= fn.Pos() && call.End() <= fn.End() {
				span := int(fn.End() - fn.Pos())
				if span > bestSpan {
					bestSpan = span
					enc = fn
				}
			}
		}
		return true
	})
	return enc
}
