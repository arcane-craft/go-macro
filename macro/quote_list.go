package macro

import (
	"fmt"
	"go/ast"
	"go/token"
)

// syntaxList wraps ellipsis captures for Quote #name ... injection.
type syntaxList struct {
	elems []Syntax
	pos   token.Pos
}

// WrapSyntaxList wraps a capture list for Quote ellipsis holes.
func WrapSyntaxList(elems []Syntax) Syntax {
	pos := token.NoPos
	if len(elems) > 0 {
		pos = elems[0].MacroPos()
	}
	return &syntaxList{elems: elems, pos: pos}
}

func (s *syntaxList) Match(string) (Bindings, error) {
	return nil, fmt.Errorf("macro: list Syntax cannot Match")
}

func (s *syntaxList) MacroPos() token.Pos { return s.pos }

func (s *syntaxList) Underlying() ast.Node { return nil }

func (s *syntaxList) ToExpr() (ast.Expr, error) {
	return nil, fmt.Errorf("macro: list Syntax is not Expr")
}

func (s *syntaxList) ToExprs() ([]ast.Expr, error) {
	return nil, fmt.Errorf("macro: list Syntax is not Exprs")
}

func (s *syntaxList) ToStmt() (ast.Stmt, error) {
	return nil, fmt.Errorf("macro: list Syntax is not Stmt")
}

func (s *syntaxList) ToStmts() ([]ast.Stmt, error) {
	return nil, fmt.Errorf("macro: list Syntax is not Stmts")
}

func (s *syntaxList) ToDecl() (ast.Decl, error) {
	return nil, fmt.Errorf("macro: list Syntax is not Decl")
}

func (s *syntaxList) ToDecls() ([]ast.Decl, error) {
	return nil, fmt.Errorf("macro: list Syntax is not Decls")
}

// QuoteElems returns underlying AST nodes for quote binding.
func QuoteElems(s Syntax) ([]ast.Node, bool) {
	if sl, ok := s.(*syntaxList); ok {
		out := make([]ast.Node, len(sl.elems))
		for i, e := range sl.elems {
			out[i] = e.Underlying()
		}
		return out, true
	}
	return nil, false
}
