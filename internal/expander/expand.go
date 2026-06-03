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

// ExpandFile expands decl macros then call macros in file and mutates AST in place.
func (e *Engine) ExpandFile(
	fset *token.FileSet,
	file *ast.File,
	info *types.Info,
	pkg *types.Package,
	imports map[string]string,
) error {
	if err := e.ExpandDeclMacros(fset, file, info, pkg, imports); err != nil {
		return err
	}
	return e.expandCallMacros(fset, file, info, pkg, imports)
}

func (e *Engine) expandCallMacros(
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
		ctx, err := macro.NewCallContext(fset, file, info, pkg, mc.Call, mc.StubName, mc.SyntaxID, site, enc)
		if err != nil {
			return err
		}
		result, err := expand(ctx, mc.Call)
		if err != nil {
			return err
		}
		if err := macro.ValidateCallExpandResult(ctx, result); err != nil {
			return macro.ErrorAt(fset, mc.Call.Pos(), "%s", err.Error())
		}
		if err := ApplyExpandResult(file, mc.Call, result); err != nil {
			return macro.ErrorAt(fset, mc.Call.Pos(), "%s", err.Error())
		}
	}
	return nil
}

// RegisterLinked registers linked call expanders for import paths (legacy tests).
func (e *Engine) RegisterLinked(active map[string]macro.CallExpander, filesByPath map[string][]*ast.File) error {
	for importPath, expand := range active {
		files := filesByPath[importPath]
		if err := e.Registry.RegisterProvider(importPath, files, expand); err != nil {
			return fmt.Errorf("register %s: %w", importPath, err)
		}
	}
	return nil
}
