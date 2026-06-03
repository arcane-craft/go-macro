package macro

import (
	"go/ast"
	"go/token"
	"go/types"
)

// MacroTag holds optional key=value pairs from a struct field's `macro:"..."` tag.
type MacroTag map[string]string

// DeclExpandResult holds the output of a declaration macro expansion.
// On success, Fields and Methods MUST both be non-nil and express the full Target shape.
type DeclExpandResult struct {
	Fields  []*ast.Field
	Methods []*ast.FuncDecl
}

// DeclExpander expands a declaration macro site (anonymous embedded marker).
type DeclExpander func(ctx DeclContext, site DeclSite) (DeclExpandResult, error)

// DeclSite describes one anonymous embed of a registered marker type.
type DeclSite struct {
	Target           *ast.TypeSpec
	TargetType       types.Type
	EmbedIndex       int
	EmbedField       *ast.Field
	MarkerImportPath string
	MarkerTypeName   string
	MarkerTypeArgs   []types.Type
	MacroTag         MacroTag
}

// DeclContext provides expansion-time information for declaration macro authors.
type DeclContext interface {
	FileSet() *token.FileSet
	File() *ast.File
	Types() *types.Info
	Package() *types.Package
	Site() DeclSite
	SyntaxID() string
	MarkerTypeName() string
	TargetMethods() []*ast.FuncDecl
	TempIdent(prefix string) *ast.Ident
	MacroPos() token.Pos
}
