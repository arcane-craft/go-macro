package expander

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"sort"

	"github.com/arcane-craft/go-macro/macro"
)

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
	if err := ValidateStubValueUsage(fset, file, info, e.Registry); err != nil {
		return err
	}
	calls, err := RecognizeMacroCalls(file, info, imports, e.Registry)
	if err != nil {
		return err
	}
	sort.Slice(calls, func(i, j int) bool {
		return calls[i].Call.Pos() > calls[j].Call.Pos()
	})
	for _, mc := range calls {
		_, expand, ok := e.Registry.Lookup(mc.ImportPath, mc.StubName)
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

// RegisterLinked registers linked expanders for import paths that appear in filesByPath.
func (e *Engine) RegisterLinked(active map[string]macro.Expander, filesByPath map[string][]*ast.File) error {
	for importPath, expand := range active {
		files := filesByPath[importPath]
		if err := e.Registry.RegisterProvider(importPath, files, expand); err != nil {
			return fmt.Errorf("register %s: %w", importPath, err)
		}
	}
	return nil
}
