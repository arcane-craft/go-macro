package expander

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/arcane-craft/go-macro/macro"
)

// ResolveSite constructs a fresh site Syntax for one expand round.
// Call anchor is *ast.CallExpr; Decl anchor is embed *ast.Field (design D18).
func ResolveSite(fset *token.FileSet, file *ast.File, anchor ast.Node) (macro.Syntax, error) {
	if file == nil || anchor == nil {
		return nil, fmt.Errorf("expander: ResolveSite requires file and anchor")
	}
	pos := anchor.Pos()
	switch a := anchor.(type) {
	case *ast.CallExpr:
		if a.Lparen.IsValid() {
			pos = a.Lparen
		}
	case *ast.Field:
		if a.Type != nil {
			pos = a.Type.Pos()
		}
	default:
		return nil, fmt.Errorf("expander: unsupported anchor type %T", anchor)
	}
	return &siteSyntax{
		file:     file,
		fset:     fset,
		anchor:   anchor,
		macroPos: pos,
		meta:     nil,
	}, nil
}
