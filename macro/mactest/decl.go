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

// ValidateDecl checks DeclExpandResult for the site in ctx.
func ValidateDecl(ctx macro.DeclContext, result macro.DeclExpandResult) error {
	return macro.ValidateDeclExpandResult(ctx, result)
}

// ExpandDecl parses snippet, finds the first decl macro site, type-checks, and invokes expand.
// The snippet must import the provider package that defines the embedded marker type.
func ExpandDecl(expand macro.DeclExpander, syntaxID, snippet string) (macro.DeclExpandResult, error) {
	fset := token.NewFileSet()
	const filename = "snippet.go"
	src := "package mactest\n\n" + snippet
	f, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return macro.DeclExpandResult{}, err
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
	pkg, err := cfg.Check("mactest", fset, []*ast.File{f}, info)
	if err != nil {
		return macro.DeclExpandResult{}, fmt.Errorf("mactest: typecheck: %w", err)
	}

	reg := macro.NewRegistry()
	reg.RegisterDeclSyntax(syntaxID, expand)

	site, err := firstDeclSite(f, info, reg)
	if err != nil {
		return macro.DeclExpandResult{}, err
	}
	reg.RegisterMarker(site.MarkerImportPath, site.MarkerTypeName, syntaxID)

	ctx := macro.NewDeclContext(fset, f, info, pkg, site, syntaxID)
	return expand(ctx, site)
}

func firstDeclSite(file *ast.File, info *types.Info, reg *macro.Registry) (macro.DeclSite, error) {
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
			targetType := info.TypeOf(ts.Name)
			for i, field := range st.Fields.List {
				if len(field.Names) > 0 {
					continue
				}
				importPath, baseName, typeArgs, ok := resolveEmbed(info, field.Type)
				if !ok {
					continue
				}
				reg.RegisterMarker(importPath, baseName, "pending")
				site := macro.DeclSite{
					Target:           ts,
					TargetType:       targetType,
					EmbedIndex:       i,
					EmbedField:       field,
					MarkerImportPath: importPath,
					MarkerTypeName:   baseName,
					MarkerTypeArgs:   typeArgs,
					MacroTag:         macro.ParseMacroTag(field.Tag),
				}
				return site, nil
			}
		}
	}
	return macro.DeclSite{}, fmt.Errorf("mactest: no anonymous embed in snippet")
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
