package expander

import (
	"go/ast"
	"go/types"
)

func resolveStubCall(call *ast.CallExpr, info *types.Info, imports map[string]string) (stubName, pkgPath string, ok bool) {
	_ = imports
	return resolveStubExpr(call.Fun, info)
}

func resolveStubExpr(expr ast.Expr, info *types.Info) (stubName, pkgPath string, ok bool) {
	switch e := unwrapParen(expr).(type) {
	case *ast.Ident:
		obj := info.Uses[e]
		if obj == nil {
			return "", "", false
		}
		return objectStub(obj)
	case *ast.SelectorExpr:
		if !isPackageSelector(e, info) {
			return "", "", false
		}
		obj := info.Uses[e.Sel]
		if obj == nil {
			return "", "", false
		}
		return objectStub(obj)
	default:
		return "", "", false
	}
}

func objectStub(obj types.Object) (stubName, pkgPath string, ok bool) {
	fn, ok := obj.(*types.Func)
	if !ok {
		return "", "", false
	}
	typ := fn.Type()
	sig, ok := typ.(*types.Signature)
	if !ok || sig == nil {
		return "", "", false
	}
	if sig.Recv() != nil {
		return "", "", false
	}
	pkg := fn.Pkg()
	if pkg == nil {
		return "", "", false
	}
	return fn.Name(), pkg.Path(), true
}

func isPackageSelector(sel *ast.SelectorExpr, info *types.Info) bool {
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	obj := info.Uses[id]
	if obj == nil {
		return false
	}
	_, ok = obj.(*types.PkgName)
	return ok
}

func unwrapParen(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}
