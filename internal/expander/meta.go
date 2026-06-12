package expander

import (
	"go/ast"

	"github.com/arcane-craft/go-macro/macro"
)

// MatchRoot classifies which top-level pattern form matched.
type MatchRoot int

const (
	MatchRootCall MatchRoot = iota
	MatchRootStmt
	MatchRootDecl
)

// ContainerField names a parent AST slot targeted by ReplaceInContainer.
type ContainerField int

const (
	ContainerBlockStmts ContainerField = iota
	ContainerAssignRhs
	ContainerReturnResults
	ContainerGenDeclSpecs
	ContainerExprSlot
)

// SpliceMode describes how many nodes replace a container slot.
type SpliceMode int

const (
	SpliceOneToOne SpliceMode = iota
	SpliceOneToMany
	SpliceReplaceAll
)

// ReplaceInContainer replaces nodes in a parent container field.
type ReplaceInContainer struct {
	Parent         ast.Node
	ContainerField ContainerField
	Index          int  // ignored when Mode is SpliceReplaceAll
	Mode           SpliceMode
}

// InsertAfterInFileDecls appends decls after a GenDecl in file.Decls.
type InsertAfterInFileDecls struct {
	After *ast.GenDecl
}

// SpliceStep is one splice operation in a Match-derived plan.
type SpliceStep struct {
	Replace  *ReplaceInContainer
	InsertAfter *InsertAfterInFileDecls
}

// MatchMeta holds match results written to the site internal slot.
type MatchMeta struct {
	Bindings    macro.Bindings
	MatchedSpan ast.Node
	Plan        []SpliceStep
	MatchRoot   MatchRoot
}

// SetMatchMeta writes meta into a siteSyntax internal slot.
func SetMatchMeta(site macro.Syntax, meta MatchMeta) bool {
	s, ok := site.(*siteSyntax)
	if !ok {
		return false
	}
	s.meta = &meta
	return true
}

// ClearMatchMeta empties the site meta slot.
func ClearMatchMeta(site macro.Syntax) {
	if s, ok := site.(*siteSyntax); ok {
		s.meta = nil
	}
}

// MatchMetaFromSite reads meta from a siteSyntax internal slot.
func MatchMetaFromSite(site macro.Syntax) (MatchMeta, bool) {
	s, ok := site.(*siteSyntax)
	if !ok || s.meta == nil {
		return MatchMeta{}, false
	}
	return *s.meta, true
}
