package macro

import (
	"go/ast"
	"go/token"
	"go/types"
)

// CallSiteKind describes where a macro call appears in the AST.
type CallSiteKind int

const (
	SiteAssign CallSiteKind = iota // lhs := Macro(...)
	SiteReturn                     // return Macro(...)
	SiteStmt                       // Macro(...); / Try0
	SiteExpr                       // expression position
)

// ExpandResult holds the expansion output. Set one primary field matching the call site.
type ExpandResult struct {
	Stmts []ast.Stmt // replaces AssignStmt, ReturnStmt, or ExprStmt
	Exprs []ast.Expr // replaces ReturnStmt.Results only (rare)
	Expr  ast.Expr   // replaces CallExpr only (expression macros)
}

// Expander expands a macro call.
type Expander func(ctx Context, call *ast.CallExpr) (ExpandResult, error)

// Context provides expansion-time information for macro authors.
type Context interface {
	FileSet() *token.FileSet
	Types() *types.Info
	Package() *types.Package
	Call() *ast.CallExpr
	StubName() string
	SyntaxID() string
	Site() CallSiteKind
	EnclosingFunc() ast.Node // *ast.FuncDecl or *ast.FuncLit
	TempIdent(prefix string) *ast.Ident
	MacroPos() token.Pos
}
