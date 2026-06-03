package macro

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sync/atomic"
)

// implCallContext is the concrete CallContext implementation used by the expander.
type implCallContext struct {
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

// NewCallContext builds a CallContext for expanders. file is the macro-tagged source file
// containing call. EnclosingFunc must be *ast.FuncDecl or *ast.FuncLit.
func NewCallContext(
	fset *token.FileSet,
	file *ast.File,
	info *types.Info,
	pkg *types.Package,
	call *ast.CallExpr,
	stubName, syntaxID string,
	site CallSiteKind,
	enclosing ast.Node,
) (CallContext, error) {
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
	return &implCallContext{
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

func (c *implCallContext) FileSet() *token.FileSet { return c.fset }
func (c *implCallContext) File() *ast.File         { return c.file }
func (c *implCallContext) Types() *types.Info      { return c.info }
func (c *implCallContext) Package() *types.Package { return c.pkg }
func (c *implCallContext) Call() *ast.CallExpr     { return c.call }
func (c *implCallContext) StubName() string        { return c.stubName }
func (c *implCallContext) SyntaxID() string        { return c.syntaxID }
func (c *implCallContext) Site() CallSiteKind      { return c.site }
func (c *implCallContext) LegalSpliceTargets() []SpliceTarget {
	return LegalSpliceTargetsForCall(c.file, c.call)
}
func (c *implCallContext) EnclosingFunc() ast.Node { return c.enclosing }
func (c *implCallContext) MacroPos() token.Pos     { return c.macroPos }

func (c *implCallContext) TempIdent(prefix string) *ast.Ident {
	n := c.tempCounter.Add(1)
	name := fmt.Sprintf("%s%d", prefix, n)
	return ast.NewIdent(name)
}
