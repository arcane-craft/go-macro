package macro

import (
	"go/ast"
	"go/token"
	"go/types"
)

// CallSiteKind describes where a macro call appears in the AST (hint for expanders).
// Splicing uses ExpandResult.Target, not Site alone.
type CallSiteKind int

const (
	SiteAssign CallSiteKind = iota // lhs := Macro(...)
	SiteReturn                     // return Macro(...)
	SiteStmt                       // Macro(...); / Try0
	SiteExpr                       // expression position
)

// ExpandResult holds the expansion output. Target selects which AST node to replace.
//
// Payload by Target:
//
//	SpliceReplaceAssignStmt, SpliceReplaceReturnStmt, SpliceReplaceExprStmt → Stmts
//	SpliceReplaceAssignRHS, SpliceReplaceCallExpr → Expr
//	SpliceReplaceReturnResults → Exprs
type ExpandResult struct {
	Target SpliceTarget
	Stmts  []ast.Stmt
	Exprs  []ast.Expr
	Expr   ast.Expr
}

// Expander expands a macro call.
type Expander func(ctx Context, call *ast.CallExpr) (ExpandResult, error)

// Context provides expansion-time information for macro authors.
type Context interface {
	FileSet() *token.FileSet
	File() *ast.File
	Types() *types.Info
	Package() *types.Package
	Call() *ast.CallExpr
	StubName() string
	SyntaxID() string
	Site() CallSiteKind
	LegalSpliceTargets() []SpliceTarget
	EnclosingFunc() ast.Node // *ast.FuncDecl or *ast.FuncLit
	TempIdent(prefix string) *ast.Ident
	MacroPos() token.Pos
}
