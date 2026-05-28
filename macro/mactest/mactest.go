// Package mactest helps provider authors test Expander functions without full macro expand.
//
// After Expand, call Validate to check Target and payload against the call site:
//
//	result, err := mactest.Expand(exp, "Stub", "syntax-id", snippet)
//	if err != nil { ... }
//	if err := mactest.Validate(ctx, result); err != nil { ... }
package mactest

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/arcane-craft/go-macro/macro"
)

// Validate checks ExpandResult Target and payload for the call site in ctx.
func Validate(ctx macro.Context, result macro.ExpandResult) error {
	return macro.ValidateExpandResult(ctx, result)
}

// Expand parses snippet as a package body, finds the first macro CallExpr named stubName,
// type-checks, and invokes expand.
func Expand(expand macro.Expander, stubName, syntaxID string, snippet string) (macro.ExpandResult, error) {
	fset := token.NewFileSet()
	const filename = "snippet.go"
	src := "package mactest\n\n" + snippet
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return macro.ExpandResult{}, err
	}

	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if call == nil {
			if c, ok := n.(*ast.CallExpr); ok {
				if id, ok := c.Fun.(*ast.Ident); ok && id.Name == stubName {
					call = c
				}
			}
		}
		return true
	})
	var enclosing ast.Node
	if call != nil {
		best := -1
		ast.Inspect(f, func(n ast.Node) bool {
			switch fn := n.(type) {
			case *ast.FuncDecl:
				if fn.Body != nil && call.Pos() >= fn.Pos() && call.End() <= fn.End() {
					span := int(fn.End() - fn.Pos())
					if span > best {
						best = span
						enclosing = fn
					}
				}
			}
			return true
		})
	}
	if call == nil {
		return macro.ExpandResult{}, fmt.Errorf("mactest: no call to %s in snippet", stubName)
	}
	if enclosing == nil {
		return macro.ExpandResult{}, fmt.Errorf("mactest: snippet must contain a function")
	}

	cfg := &types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types:     make(map[ast.Expr]types.TypeAndValue),
		Defs:      make(map[*ast.Ident]types.Object),
		Uses:      make(map[*ast.Ident]types.Object),
		Instances: make(map[*ast.Ident]types.Instance),
		Scopes:    make(map[ast.Node]*types.Scope),
	}
	pkg, err := cfg.Check("mactest", fset, []*ast.File{f}, info)
	if err != nil {
		return macro.ExpandResult{}, fmt.Errorf("mactest: typecheck: %w", err)
	}

	site := classifySite(f, call)
	ctx, err := macro.NewContext(fset, f, info, pkg, call, stubName, syntaxID, site, enclosing)
	if err != nil {
		return macro.ExpandResult{}, err
	}
	return expand(ctx, call)
}

func classifySite(f *ast.File, call *ast.CallExpr) macro.CallSiteKind {
	site := macro.SiteExpr
	ast.Inspect(f, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			for _, rhs := range stmt.Rhs {
				if rhs == call {
					site = macro.SiteAssign
				}
			}
		case *ast.ReturnStmt:
			for _, r := range stmt.Results {
				if r == call {
					site = macro.SiteReturn
				}
			}
		case *ast.ExprStmt:
			if stmt.X == call {
				site = macro.SiteStmt
			}
		}
		return true
	})
	return site
}
