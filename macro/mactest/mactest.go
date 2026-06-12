// Package mactest helps provider authors test Expander functions without full macro expand.
package mactest

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"

	"github.com/arcane-craft/go-macro/internal/expander"
	"github.com/arcane-craft/go-macro/macro"
)

// ExpandSyntax parses snippet, finds the first call to stubName, and invokes expand.
func ExpandSyntax(expand macro.Expander, stubName, syntaxID string, snippet string) (macro.Syntax, error) {
	fset := token.NewFileSet()
	const filename = "snippet.go"
	src := "package mactest\n\n" + snippet
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	call, err := firstCallNamed(f, stubName)
	if err != nil {
		return nil, err
	}
	cfg := &types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types:     make(map[ast.Expr]types.TypeAndValue),
		Defs:      make(map[*ast.Ident]types.Object),
		Uses:      make(map[*ast.Ident]types.Object),
		Instances: make(map[*ast.Ident]types.Instance),
		Scopes:    make(map[ast.Node]*types.Scope),
	}
	if _, err := cfg.Check("mactest", fset, []*ast.File{f}, info); err != nil {
		return nil, fmt.Errorf("mactest: typecheck: %w", err)
	}
	site, err := expander.ResolveSite(fset, f, call)
	if err != nil {
		return nil, err
	}
	ctx := macro.NewContext(fset, info)
	_ = syntaxID
	return expand(ctx, site)
}

func firstCallNamed(f *ast.File, stubName string) (*ast.CallExpr, error) {
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
	if call == nil {
		return nil, fmt.Errorf("mactest: no call to %s in snippet", stubName)
	}
	return call, nil
}
