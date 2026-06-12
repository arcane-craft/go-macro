package macro

import (
	"fmt"
	"go/ast"
	"go/token"
)

// WrapNode wraps an AST node as Syntax. The node shape determines To* behavior.
func WrapNode(node ast.Node) Syntax {
	if node == nil {
		return &astSyntax{pos: token.NoPos, shape: shapeNode}
	}
	return &astSyntax{node: node, pos: node.Pos(), shape: shapeNode}
}

// WrapExpr wraps a single expression.
func WrapExpr(expr ast.Expr) Syntax {
	if expr == nil {
		return &astSyntax{pos: token.NoPos, shape: shapeNode}
	}
	return &astSyntax{node: expr, pos: expr.Pos(), shape: shapeNode}
}

// WrapExprs wraps an expression list.
func WrapExprs(exprs []ast.Expr) Syntax {
	return &astSyntax{exprs: exprs, shape: shapeExprs, pos: firstPosExprs(exprs)}
}

// WrapStmt wraps a single statement.
func WrapStmt(stmt ast.Stmt) Syntax {
	if stmt == nil {
		return &astSyntax{pos: token.NoPos, shape: shapeNode}
	}
	return &astSyntax{node: stmt, pos: stmt.Pos(), shape: shapeNode}
}

// WrapStmts wraps a statement list.
func WrapStmts(stmts []ast.Stmt) Syntax {
	return &astSyntax{stmts: stmts, shape: shapeStmts, pos: firstPosStmts(stmts)}
}

// WrapDecl wraps a single declaration.
func WrapDecl(decl ast.Decl) Syntax {
	if decl == nil {
		return &astSyntax{pos: token.NoPos, shape: shapeNode}
	}
	return &astSyntax{node: decl, pos: decl.Pos(), shape: shapeNode}
}

// WrapDecls wraps a declaration list.
func WrapDecls(decls []ast.Decl) Syntax {
	return &astSyntax{decls: decls, shape: shapeDecls, pos: firstPosDecls(decls)}
}

type syntaxShape int

const (
	shapeNode syntaxShape = iota
	shapeExprs
	shapeStmts
	shapeDecls
)

type astSyntax struct {
	node  ast.Node
	exprs []ast.Expr
	stmts []ast.Stmt
	decls []ast.Decl
	shape syntaxShape
	pos   token.Pos
}

func (s *astSyntax) Match(pattern string) (Bindings, error) {
	return nil, fmt.Errorf("macro: Match on wrapped Syntax requires engine site")
}

func (s *astSyntax) MacroPos() token.Pos { return s.pos }

func (s *astSyntax) Underlying() ast.Node {
	switch s.shape {
	case shapeExprs:
		if len(s.exprs) == 1 {
			return s.exprs[0]
		}
	case shapeStmts:
		if len(s.stmts) == 1 {
			return s.stmts[0]
		}
	case shapeDecls:
		if len(s.decls) == 1 {
			return s.decls[0]
		}
	}
	return s.node
}

func (s *astSyntax) ToExpr() (ast.Expr, error) {
	if e, ok := s.node.(ast.Expr); ok && s.shape == shapeNode {
		return e, nil
	}
	return nil, fmt.Errorf("macro: Syntax is not a single Expr")
}

func (s *astSyntax) ToExprs() ([]ast.Expr, error) {
	switch s.shape {
	case shapeExprs:
		if len(s.exprs) == 0 {
			return nil, fmt.Errorf("macro: Syntax expr list is empty")
		}
		return s.exprs, nil
	case shapeNode:
		if e, ok := s.node.(ast.Expr); ok {
			return []ast.Expr{e}, nil
		}
	}
	return nil, fmt.Errorf("macro: Syntax is not Expr or []Expr")
}

func (s *astSyntax) ToStmt() (ast.Stmt, error) {
	if st, ok := s.node.(ast.Stmt); ok && s.shape == shapeNode {
		return st, nil
	}
	return nil, fmt.Errorf("macro: Syntax is not a single Stmt")
}

func (s *astSyntax) ToStmts() ([]ast.Stmt, error) {
	switch s.shape {
	case shapeStmts:
		if len(s.stmts) == 0 {
			return nil, fmt.Errorf("macro: Syntax stmt list is empty")
		}
		return s.stmts, nil
	case shapeNode:
		if st, ok := s.node.(ast.Stmt); ok {
			return []ast.Stmt{st}, nil
		}
	}
	return nil, fmt.Errorf("macro: Syntax is not Stmt or []Stmt")
}

func (s *astSyntax) ToDecl() (ast.Decl, error) {
	if d, ok := s.node.(ast.Decl); ok && s.shape == shapeNode {
		return d, nil
	}
	return nil, fmt.Errorf("macro: Syntax is not a single Decl")
}

func (s *astSyntax) ToDecls() ([]ast.Decl, error) {
	switch s.shape {
	case shapeDecls:
		if len(s.decls) == 0 {
			return nil, fmt.Errorf("macro: Syntax decl list is empty")
		}
		return s.decls, nil
	case shapeNode:
		if d, ok := s.node.(ast.Decl); ok {
			return []ast.Decl{d}, nil
		}
	}
	return nil, fmt.Errorf("macro: Syntax is not Decl or []Decl")
}

type mapBindings struct {
	singles map[string]Syntax
	lists   map[string][]Syntax
}

func newMapBindings() *mapBindings {
	return &mapBindings{
		singles: make(map[string]Syntax),
		lists:   make(map[string][]Syntax),
	}
}

func (b *mapBindings) Get(name string) (Syntax, bool) {
	v, ok := b.singles[name]
	return v, ok
}

func (b *mapBindings) Elems(name string) ([]Syntax, bool) {
	v, ok := b.lists[name]
	if !ok {
		return nil, false
	}
	return v, true
}

func firstPosExprs(exprs []ast.Expr) token.Pos {
	for _, e := range exprs {
		if e != nil {
			return e.Pos()
		}
	}
	return token.NoPos
}

func firstPosStmts(stmts []ast.Stmt) token.Pos {
	for _, s := range stmts {
		if s != nil {
			return s.Pos()
		}
	}
	return token.NoPos
}

func firstPosDecls(decls []ast.Decl) token.Pos {
	for _, d := range decls {
		if d != nil {
			return d.Pos()
		}
	}
	return token.NoPos
}
