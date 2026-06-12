package expander

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/arcane-craft/go-macro/macro"
	"github.com/arcane-craft/go-macro/macro/pattern"
)

// siteSyntax is the engine site implementation with an internal meta slot.
type siteSyntax struct {
	file     *ast.File
	fset     *token.FileSet
	anchor   ast.Node
	macroPos token.Pos
	meta     *MatchMeta
}

func (s *siteSyntax) Match(pat string) (macro.Bindings, error) {
	parsed, err := pattern.Parse(pat)
	if err != nil {
		return nil, err
	}
	binds, meta, err := matchPattern(s, parsed)
	if err != nil {
		s.meta = nil
		return nil, err
	}
	s.meta = &meta
	return binds, nil
}

func (s *siteSyntax) MacroPos() token.Pos { return s.macroPos }

// ExpansionFile implements macro.FileCarrier.
func (s *siteSyntax) ExpansionFile() *ast.File { return s.file }

// ClearExpansionMeta implements macro.MetaSlot.
func (s *siteSyntax) ClearExpansionMeta() { s.meta = nil }

func (s *siteSyntax) Underlying() ast.Node { return s.anchor }

func (s *siteSyntax) ToExpr() (ast.Expr, error) {
	return nil, fmt.Errorf("macro: site Syntax has no expansion payload")
}

func (s *siteSyntax) ToExprs() ([]ast.Expr, error) {
	return nil, fmt.Errorf("macro: site Syntax has no expansion payload")
}

func (s *siteSyntax) ToStmt() (ast.Stmt, error) {
	return nil, fmt.Errorf("macro: site Syntax has no expansion payload")
}

func (s *siteSyntax) ToStmts() ([]ast.Stmt, error) {
	return nil, fmt.Errorf("macro: site Syntax has no expansion payload")
}

func (s *siteSyntax) ToDecl() (ast.Decl, error) {
	return nil, fmt.Errorf("macro: site Syntax has no expansion payload")
}

func (s *siteSyntax) ToDecls() ([]ast.Decl, error) {
	return nil, fmt.Errorf("macro: site Syntax has no expansion payload")
}
