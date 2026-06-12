package macro

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sync/atomic"
)

type implContext struct {
	fset        *token.FileSet
	info        *types.Info
	tempCounter *atomic.Uint64
}

// NewContext builds a Context for expanders.
func NewContext(fset *token.FileSet, info *types.Info) Context {
	return &implContext{
		fset:        fset,
		info:        info,
		tempCounter: &atomic.Uint64{},
	}
}

func (c *implContext) FileSet() *token.FileSet { return c.fset }
func (c *implContext) Types() *types.Info     { return c.info }

func (c *implContext) TempIdent(prefix string) *ast.Ident {
	n := c.tempCounter.Add(1)
	name := fmt.Sprintf("%s%d", prefix, n)
	return ast.NewIdent(name)
}
