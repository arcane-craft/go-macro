package macro

import (
	"go/ast"
	"go/token"
	"go/types"
)

// Syntax is the unified type for macro match, quote, and splice payloads.
type Syntax interface {
	Match(pattern string) (Bindings, error)
	ToExpr() (ast.Expr, error)
	ToExprs() ([]ast.Expr, error)
	ToStmt() (ast.Stmt, error)
	ToStmts() ([]ast.Stmt, error)
	ToDecl() (ast.Decl, error)
	ToDecls() ([]ast.Decl, error)
	Underlying() ast.Node
	MacroPos() token.Pos
}

// Bindings holds pattern capture results from a successful Match.
type Bindings interface {
	Get(name string) (Syntax, bool)
	Elems(name string) ([]Syntax, bool)
}

// Expander is the unified macro expansion signature.
type Expander func(ctx Context, site Syntax) (Syntax, error)

// Context provides minimal expansion-time services for macro authors.
type Context interface {
	FileSet() *token.FileSet
	Types() *types.Info
	TempIdent(prefix string) *ast.Ident
}
