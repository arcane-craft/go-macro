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

// Expand parses snippet, finds the first decl macro site, type-checks, and invokes expand.
func Expand(expand macro.Expander, syntaxID, snippet string) (macro.Syntax, error) {
	fset := token.NewFileSet()
	const filename = "snippet.go"
	src := "package mactest\n\n" + snippet
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	cfg := &types.Config{Importer: importer.Default()}
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Instances:  make(map[*ast.Ident]types.Instance),
		Scopes:     make(map[ast.Node]*types.Scope),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}
	if _, err := cfg.Check("mactest", fset, []*ast.File{f}, info); err != nil {
		return nil, fmt.Errorf("mactest: typecheck: %w", err)
	}
	reg := macro.NewRegistry()
	reg.RegisterExpander(syntaxID, expand)
	site, err := firstDeclSite(f, info, reg)
	if err != nil {
		return nil, err
	}
	reg.RegisterMarker(site.MarkerImportPath, site.MarkerTypeName, syntaxID)
	syn, err := expander.ResolveSite(fset, f, site.EmbedField)
	if err != nil {
		return nil, err
	}
	ctx := macro.NewContext(fset, info)
	return expand(ctx, syn)
}

type declSite struct {
	MarkerImportPath string
	MarkerTypeName   string
	EmbedField       *ast.Field
}

func firstDeclSite(file *ast.File, info *types.Info, reg *macro.Registry) (declSite, error) {
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			st, ok := ts.Type.(*ast.StructType)
			if !ok || st.Fields == nil {
				continue
			}
			for _, field := range st.Fields.List {
				if len(field.Names) > 0 {
					continue
				}
				importPath, baseName, _, ok := resolveEmbed(info, field.Type)
				if !ok {
					continue
				}
				reg.RegisterMarker(importPath, baseName, "pending")
				return declSite{
					MarkerImportPath: importPath,
					MarkerTypeName:   baseName,
					EmbedField:       field,
				}, nil
			}
		}
	}
	return declSite{}, fmt.Errorf("mactest: no anonymous embed in snippet")
}

func resolveEmbed(info *types.Info, typ ast.Expr) (importPath, baseName string, typeArgs []types.Type, ok bool) {
	tv := info.TypeOf(typ)
	if tv == nil {
		return "", "", nil, false
	}
	named, ok := types.Unalias(tv).(*types.Named)
	if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
		return "", "", nil, false
	}
	importPath = named.Obj().Pkg().Path()
	baseName = named.Obj().Name()
	if named.TypeArgs() != nil {
		for i := 0; i < named.TypeArgs().Len(); i++ {
			typeArgs = append(typeArgs, named.TypeArgs().At(i))
		}
	}
	return importPath, baseName, typeArgs, true
}
