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
		if err := e.expandOneCallSite(fset, file, info, pkg, mc); err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) expandOneCallSite(
	fset *token.FileSet,
	file *ast.File,
	info *types.Info,
	pkg *types.Package,
	mc MacroCall,
) error {
	expand, ok := e.lookupUnifiedExpander(mc.ImportPath, mc.StubName, mc.SyntaxID)
	if !ok {
		return macro.ErrorAt(fset, mc.Call.Pos(), "unknown macro stub %q", mc.StubName)
	}
	site, err := ResolveSite(fset, file, mc.Call)
	if err != nil {
		return macro.ErrorAt(fset, mc.Call.Pos(), "%s", err.Error())
	}
	ctx := macro.NewContext(fset, info)
	out, err := expand(ctx, site)
	if err != nil {
		return macro.ErrorAt(fset, mc.Call.Pos(), "%s", err.Error())
	}
	meta, ok := MatchMetaFromSite(site)
	if !ok {
		return macro.ErrorAt(fset, mc.Call.Pos(), "macro: missing match meta after expand")
	}
	if err := ValidateSplice(out, meta); err != nil {
		return macro.ErrorAt(fset, mc.Call.Pos(), "%s", err.Error())
	}
	if err := Apply(file, meta, out); err != nil {
		return macro.ErrorAt(fset, mc.Call.Pos(), "%s", err.Error())
	}
	if stmts, err := out.ToStmts(); err == nil {
		macro.StampStmtPos(site.MacroPos(), stmts)
	}
	_ = pkg
	return nil
}

func (e *Engine) expandOneDeclSite(
	fset *token.FileSet,
	file *ast.File,
	info *types.Info,
	embed *ast.Field,
	expand macro.Expander,
) error {
	site, err := ResolveSite(fset, file, embed)
	if err != nil {
		return err
	}
	ctx := macro.NewContext(fset, info)
	out, err := expand(ctx, site)
	if err != nil {
		return err
	}
	meta, ok := MatchMetaFromSite(site)
	if !ok {
		return fmt.Errorf("macro: missing match meta after expand")
	}
	if err := ValidateSplice(out, meta); err != nil {
		return err
	}
	if err := Apply(file, meta, out); err != nil {
		return err
	}
	if stmts, err := out.ToStmts(); err == nil {
		macro.StampStmtPos(site.MacroPos(), stmts)
	}
	return nil
}

func (e *Engine) lookupDeclSyntaxID(importPath, markerName string) (string, bool) {
	return e.Registry.SyntaxIDForMarker(importPath, markerName)
}

func (e *Engine) lookupUnifiedExpanderForDecl(syntaxID, _, _ string) (macro.Expander, bool) {
	return e.Registry.LookupExpander(syntaxID)
}

func (e *Engine) lookupUnifiedExpander(_, _, syntaxID string) (macro.Expander, bool) {
	return e.Registry.LookupExpander(syntaxID)
}

// RegisterLinked registers linked expanders for import paths (tests).
func (e *Engine) RegisterLinked(active map[string]macro.Expander, filesByPath map[string][]*ast.File) error {
	for importPath, expand := range active {
		files := filesByPath[importPath]
		if err := e.Registry.RegisterProviderSources(importPath, files); err != nil {
			return fmt.Errorf("register %s: %w", importPath, err)
		}
		scan, err := macro.ScanProviderFiles(files)
		if err != nil {
			return fmt.Errorf("register %s: %w", importPath, err)
		}
		for _, ent := range scan.Entries {
			if ent.Expander != "" {
				e.Registry.RegisterExpander(ent.SyntaxID, expand)
			}
		}
	}
	return nil
}
