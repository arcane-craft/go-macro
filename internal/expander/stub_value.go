package expander

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/arcane-craft/go-macro/macro"
)

// ValidateStubValueUsage reports an error when a registered macro stub is used as a
// function value instead of a direct call.
func ValidateStubValueUsage(
	fset *token.FileSet,
	file *ast.File,
	info *types.Info,
	reg *macro.Registry,
) error {
	imports := BuildImportMap(file, "")
	parents := buildParentMap(file)
	var err error
	ast.Inspect(file, func(n ast.Node) bool {
		if err != nil {
			return false
		}
		switch e := n.(type) {
		case *ast.Ident:
			if isSelectorSel(e, parents) {
				return true
			}
			err = checkStubExpr(fset, e, info, reg, imports, parents)
		case *ast.SelectorExpr:
			err = checkStubExpr(fset, e, info, reg, imports, parents)
		}
		return err == nil
	})
	return err
}

func checkStubExpr(
	fset *token.FileSet,
	expr ast.Expr,
	info *types.Info,
	reg *macro.Registry,
	imports map[string]string,
	parents map[ast.Node]ast.Node,
) error {
	stub, pkgPath, ok := resolveStubExpr(expr, info)
	if !ok || !reg.HasStub(pkgPath, stub) {
		return nil
	}
	if isDirectMacroCallee(expr, parents) {
		return nil
	}
	qual := stubInvokeQualifier(expr, imports, stub, pkgPath)
	return macro.ErrorAt(
		fset,
		expr.Pos(),
		"macro stub %q must be invoked directly (e.g. %s), not used as a function value",
		stub,
		qual,
	)
}

func stubInvokeQualifier(expr ast.Expr, imports map[string]string, stub, importPath string) string {
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		if id, ok := sel.X.(*ast.Ident); ok {
			return id.Name + "." + stub + "(...)"
		}
	}
	if _, ok := expr.(*ast.Ident); ok {
		return stub + "(...)"
	}
	for local, path := range imports {
		if path == importPath && local != "." {
			return local + "." + stub + "(...)"
		}
	}
	return stub + "(...)"
}

type parentVisitor struct {
	parent ast.Node
	m      map[ast.Node]ast.Node
}

func (v *parentVisitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	if v.parent != nil {
		v.m[n] = v.parent
	}
	return &parentVisitor{parent: n, m: v.m}
}

func buildParentMap(file *ast.File) map[ast.Node]ast.Node {
	m := make(map[ast.Node]ast.Node)
	ast.Walk(&parentVisitor{m: m}, file)
	return m
}

func isSelectorSel(id *ast.Ident, parents map[ast.Node]ast.Node) bool {
	p := parents[id]
	sel, ok := p.(*ast.SelectorExpr)
	return ok && sel.Sel == id
}

func isDirectMacroCallee(expr ast.Expr, parents map[ast.Node]ast.Node) bool {
	for p := parents[expr]; p != nil; p = parents[p] {
		call, ok := p.(*ast.CallExpr)
		if !ok {
			continue
		}
		if unwrapParen(call.Fun) == expr {
			return true
		}
	}
	return false
}
