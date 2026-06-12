// Package quote implements template-based AST construction for macro.Quote.
//
// Typed APIs (Expr, Exprs, Stmts, Decls) take the template body directly, e.g.
// quote.Stmts("x := 1", nil). An explicit @kind{ } wrapper is optional on those
// entry points. Quote requires @expr{ }, @exprs{ }, @stmts{ }, or @decls{ }.
// Holes use #name in the body, filled from a map[string]any binding table.
package quote

import (
	"go/ast"
)

func evalTemplate(tpl string, args map[string]any, expect Kind) (any, error) {
	pt, err := resolveTemplate(tpl, expect)
	if err != nil {
		return nil, err
	}
	if err := validateBindings(pt.Holes, args); err != nil {
		return nil, err
	}
	if fast, ok, err := tryFastPath(pt.Kind, pt.Body, args); err != nil {
		return nil, err
	} else if ok {
		return fast, nil
	}
	src, err := synthesize(pt.Kind, pt.Body)
	if err != nil {
		return nil, err
	}
	parsed, err := parseSynthesized(pt.Kind, src)
	if err != nil {
		return nil, err
	}
	if err := bindAll(parsed, args); err != nil {
		return nil, err
	}
	switch pt.Kind {
	case KindExpr:
		return parsed.expr, nil
	case KindExprs:
		return parsed.exprs, nil
	case KindStmts:
		return parsed.stmts, nil
	case KindDecls:
		return parsed.decls, nil
	default:
		return nil, errf("unknown kind %v", pt.Kind)
	}
}

// Quote parses tpl and returns nodes according to the root @kind.
func Quote(tpl string, args map[string]any) ([]ast.Node, error) {
	pt, err := parseTemplate(tpl)
	if err != nil {
		return nil, err
	}
	if err := validateBindings(pt.Holes, args); err != nil {
		return nil, err
	}
	if fast, ok, err := tryFastPath(pt.Kind, pt.Body, args); err != nil {
		return nil, err
	} else if ok {
		return valueToNodes(pt.Kind, fast)
	}
	src, err := synthesize(pt.Kind, pt.Body)
	if err != nil {
		return nil, err
	}
	parsed, err := parseSynthesized(pt.Kind, src)
	if err != nil {
		return nil, err
	}
	if err := bindAll(parsed, args); err != nil {
		return nil, err
	}
	return parsedToNodes(parsed)
}

// Expr expands a template body to a single ast.Expr.
func Expr(tpl string, args map[string]any) (ast.Expr, error) {
	v, err := evalTemplate(tpl, args, KindExpr)
	if err != nil {
		return nil, err
	}
	e, ok := v.(ast.Expr)
	if !ok {
		return nil, errf("internal: expected ast.Expr")
	}
	return e, nil
}

// Exprs expands a template body to a []ast.Expr.
func Exprs(tpl string, args map[string]any) ([]ast.Expr, error) {
	v, err := evalTemplate(tpl, args, KindExprs)
	if err != nil {
		return nil, err
	}
	exprs, ok := v.([]ast.Expr)
	if !ok {
		return nil, errf("internal: expected []ast.Expr")
	}
	return exprs, nil
}

// Stmts expands a template body to a []ast.Stmt.
func Stmts(tpl string, args map[string]any) ([]ast.Stmt, error) {
	v, err := evalTemplate(tpl, args, KindStmts)
	if err != nil {
		return nil, err
	}
	stmts, ok := v.([]ast.Stmt)
	if !ok {
		return nil, errf("internal: expected []ast.Stmt")
	}
	return stmts, nil
}

// Decls expands a template body to a []ast.Decl.
func Decls(tpl string, args map[string]any) ([]ast.Decl, error) {
	v, err := evalTemplate(tpl, args, KindDecls)
	if err != nil {
		return nil, err
	}
	decls, ok := v.([]ast.Decl)
	if !ok {
		return nil, errf("internal: expected []ast.Decl")
	}
	return decls, nil
}

func valueToNodes(kind Kind, v any) ([]ast.Node, error) {
	switch kind {
	case KindExpr:
		e, ok := v.(ast.Expr)
		if !ok {
			return nil, errf("internal: expected ast.Expr")
		}
		return []ast.Node{e}, nil
	case KindExprs:
		exprs, ok := v.([]ast.Expr)
		if !ok {
			return nil, errf("internal: expected []ast.Expr")
		}
		out := make([]ast.Node, len(exprs))
		for i, e := range exprs {
			out[i] = e
		}
		return out, nil
	case KindStmts:
		stmts, ok := v.([]ast.Stmt)
		if !ok {
			return nil, errf("internal: expected []ast.Stmt")
		}
		out := make([]ast.Node, len(stmts))
		for i, s := range stmts {
			out[i] = s
		}
		return out, nil
	case KindDecls:
		decls, ok := v.([]ast.Decl)
		if !ok {
			return nil, errf("internal: expected []ast.Decl")
		}
		out := make([]ast.Node, len(decls))
		for i, d := range decls {
			out[i] = d
		}
		return out, nil
	default:
		return nil, errf("unknown kind %v", kind)
	}
}

func parsedToNodes(pt *parsedAST) ([]ast.Node, error) {
	switch pt.kind {
	case KindExpr:
		return []ast.Node{pt.expr}, nil
	case KindExprs:
		out := make([]ast.Node, len(pt.exprs))
		for i, e := range pt.exprs {
			out[i] = e
		}
		return out, nil
	case KindStmts:
		out := make([]ast.Node, len(pt.stmts))
		for i, s := range pt.stmts {
			out[i] = s
		}
		return out, nil
	case KindDecls:
		out := make([]ast.Node, len(pt.decls))
		for i, d := range pt.decls {
			out[i] = d
		}
		return out, nil
	default:
		return nil, errf("unknown kind %v", pt.kind)
	}
}
