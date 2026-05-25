package expander

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sort"

	"github.com/arcane-craft/go-macro/macro"
)

// Provider describes a macro provider package loaded into the registry.
type Provider struct {
	ImportPath string
	SyntaxID   string
	Expand     macro.Expander
}

// Engine expands macros in macro-tagged source files.
type Engine struct {
	Registry *macro.Registry
}

// ExpandFile expands all macro calls in file and mutates file AST in place.
// Calls are collected once, then expanded from back to front so types.Info stays valid.
func (e *Engine) ExpandFile(
	fset *token.FileSet,
	file *ast.File,
	info *types.Info,
	pkg *types.Package,
	imports map[string]string,
) error {
	calls, err := RecognizeMacroCalls(file, info, imports, e.Registry)
	if err != nil {
		return err
	}
	sort.Slice(calls, func(i, j int) bool {
		return calls[i].Call.Pos() > calls[j].Call.Pos()
	})
	for _, mc := range calls {
		_, expand, ok := e.Registry.Lookup(mc.StubName)
		if !ok {
			return macro.ErrorAt(fset, mc.Call.Pos(), "unknown macro stub %q", mc.StubName)
		}
		site := classifySiteInFile(file, mc.Call)
		enc := enclosingFuncInFile(file, mc.Call)
		if enc == nil {
			return macro.ErrorAt(fset, mc.Call.Pos(), "macro call must appear inside a function")
		}
		ctx, err := macro.NewContext(fset, info, pkg, mc.Call, mc.StubName, mc.SyntaxID, site, enc)
		if err != nil {
			return err
		}
		result, err := expand(ctx, mc.Call)
		if err != nil {
			return err
		}
		if err := ApplyExpandResult(file, mc.Call, site, result); err != nil {
			return macro.ErrorAt(fset, mc.Call.Pos(), "%s", err.Error())
		}
	}
	return nil
}

// RegisterProviders registers all providers with the engine registry.
func (e *Engine) RegisterProviders(providers []Provider, filesByPath map[string][]*ast.File) error {
	for _, p := range providers {
		files := filesByPath[p.ImportPath]
		if err := e.Registry.RegisterProvider(p.ImportPath, files, p.SyntaxID, p.Expand); err != nil {
			return fmt.Errorf("register %s: %w", p.ImportPath, err)
		}
	}
	return nil
}
