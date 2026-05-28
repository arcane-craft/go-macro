package macro

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sync/atomic"
)

// implContext is the concrete Context implementation used by the expander.
type implContext struct {
	fset        *token.FileSet
	file        *ast.File
	info        *types.Info
	pkg         *types.Package
	call        *ast.CallExpr
	stubName    string
	syntaxID    string
	site        CallSiteKind
	enclosing   ast.Node
	tempCounter *atomic.Uint64
	macroPos    token.Pos
}

// NewContext builds a Context for expanders. file is the macro-tagged source file
// containing call. EnclosingFunc must be *ast.FuncDecl or *ast.FuncLit.
func NewContext(
	fset *token.FileSet,
	file *ast.File,
	info *types.Info,
	pkg *types.Package,
	call *ast.CallExpr,
	stubName, syntaxID string,
	site CallSiteKind,
	enclosing ast.Node,
) (Context, error) {
	if enclosing == nil {
		return nil, fmt.Errorf("macro: EnclosingFunc is required")
	}
	switch enclosing.(type) {
	case *ast.FuncDecl, *ast.FuncLit:
	default:
		return nil, fmt.Errorf("macro: EnclosingFunc must be *ast.FuncDecl or *ast.FuncLit, got %T", enclosing)
	}
	pos := call.Pos()
	if call != nil && call.Lparen.IsValid() {
		pos = call.Lparen
	}
	return &implContext{
		fset:        fset,
		file:        file,
		info:        info,
		pkg:         pkg,
		call:        call,
		stubName:    stubName,
		syntaxID:    syntaxID,
		site:        site,
		enclosing:   enclosing,
		tempCounter: &atomic.Uint64{},
		macroPos:    pos,
	}, nil
}

func (c *implContext) FileSet() *token.FileSet { return c.fset }
func (c *implContext) File() *ast.File         { return c.file }
func (c *implContext) Types() *types.Info      { return c.info }
func (c *implContext) Package() *types.Package { return c.pkg }
func (c *implContext) Call() *ast.CallExpr     { return c.call }
func (c *implContext) StubName() string        { return c.stubName }
func (c *implContext) SyntaxID() string        { return c.syntaxID }
func (c *implContext) Site() CallSiteKind      { return c.site }
func (c *implContext) LegalSpliceTargets() []SpliceTarget {
	return LegalSpliceTargetsForCall(c.file, c.call)
}
func (c *implContext) EnclosingFunc() ast.Node { return c.enclosing }
func (c *implContext) MacroPos() token.Pos     { return c.macroPos }

func (c *implContext) TempIdent(prefix string) *ast.Ident {
	n := c.tempCounter.Add(1)
	name := fmt.Sprintf("%s%d", prefix, n)
	return ast.NewIdent(name)
}
