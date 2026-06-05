package quote

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func validateBindings(holes []string, args map[string]any) error {
	if args == nil {
		args = map[string]any{}
	}
	for _, h := range holes {
		if _, ok := args[h]; !ok {
			return errMissingHole(h)
		}
	}
	return nil
}

func tryFastPath(kind Kind, body string, args map[string]any) (any, bool, error) {
	hole, ok := isHoleOnlyBody(body)
	if !ok {
		return nil, false, nil
	}
	val, ok := args[hole]
	if !ok {
		return nil, false, errMissingHole(hole)
	}
	switch kind {
	case KindStmts:
		if stmts, ok := val.([]ast.Stmt); ok {
			return cloneStmts(stmts), true, nil
		}
	case KindDecls:
		if decls, ok := val.([]ast.Decl); ok {
			return cloneDecls(decls), true, nil
		}
	case KindExprs:
		if exprs, ok := val.([]ast.Expr); ok {
			return cloneExprs(exprs), true, nil
		}
	case KindExpr:
		if e, ok := val.(ast.Expr); ok {
			return cloneExpr(e), true, nil
		}
	}
	return nil, false, nil
}

func bindAll(pt *parsedAST, args map[string]any) error {
	var bindErr error
	applyTree := func(node ast.Node) ast.Node {
		return astutil.Apply(node, func(c *astutil.Cursor) bool {
			if bindErr != nil {
				return false
			}
			ident, ok := c.Node().(*ast.Ident)
			if !ok || !strings.HasPrefix(ident.Name, "_q_") {
				return true
			}
			hole := ident.Name[len("_q_"):]
			val, ok := args[hole]
			if !ok {
				return true
			}
			repl, err := bindingToNode(val, pt.kind)
			if err != nil {
				bindErr = errBadBinding(hole, fmt.Sprintf("%s", reflect.TypeOf(val)))
				return false
			}
			c.Replace(repl)
			return true
		}, nil)
	}

	switch pt.kind {
	case KindExpr:
		pt.expr = applyTree(pt.expr).(ast.Expr)
	case KindExprs:
		for i, e := range pt.exprs {
			pt.exprs[i] = applyTree(e).(ast.Expr)
		}
	case KindStmts:
		for i, s := range pt.stmts {
			pt.stmts[i] = applyTree(s).(ast.Stmt)
		}
	case KindDecls:
		for i, d := range pt.decls {
			pt.decls[i] = applyTree(d).(ast.Decl)
		}
	}
	return bindErr
}

func bindingToNode(val any, kind Kind) (ast.Node, error) {
	switch v := val.(type) {
	case string:
		return &ast.Ident{Name: v}, nil
	case ast.Expr:
		return cloneExpr(v), nil
	case []ast.Expr:
		if len(v) == 0 {
			return nil, errBadBinding("", "empty []ast.Expr")
		}
		return cloneExpr(v[0]), nil
	case ast.Stmt:
		return cloneStmt(v), nil
	case []ast.Stmt:
		if len(v) == 0 {
			return nil, errBadBinding("", "empty []ast.Stmt")
		}
		return cloneStmt(v[0]), nil
	case ast.Decl:
		return cloneDecl(v), nil
	case []ast.Decl:
		if len(v) == 0 {
			return nil, errBadBinding("", "empty []ast.Decl")
		}
		return cloneDecl(v[0]), nil
	case []ast.Node:
		if len(v) == 0 {
			return nil, errBadBinding("", "empty []ast.Node")
		}
		return cloneNode(v[0])
	default:
		_ = kind
		return nil, errBadBinding("", fmt.Sprintf("%s", reflect.TypeOf(val)))
	}
}
