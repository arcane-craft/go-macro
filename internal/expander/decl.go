package expander

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/arcane-craft/go-macro/macro"
)

// ExpandDeclMacros expands all declaration macro sites in file.
func (e *Engine) ExpandDeclMacros(
	fset *token.FileSet,
	file *ast.File,
	info *types.Info,
	pkg *types.Package,
	imports map[string]string,
) error {
	for {
		sites, err := RecognizeDeclSites(file, info, imports, e.Registry)
		if err != nil {
			return err
		}
		if len(sites) == 0 {
			return nil
		}
		ds := sites[0]
		syntaxID, ok := e.lookupDeclSyntaxID(ds.MarkerImportPath, ds.MarkerTypeName)
		if !ok {
			return macro.ErrorAt(fset, ds.EmbedField.Pos(), "decl macro %q not linked", ds.MarkerTypeName)
		}
		expand, ok := e.lookupUnifiedExpanderForDecl(syntaxID, ds.MarkerImportPath, ds.MarkerTypeName)
		if !ok {
			return macro.ErrorAt(fset, ds.EmbedField.Pos(), "decl macro %q requires native Expander", ds.MarkerTypeName)
		}
		if err := e.expandOneDeclSite(fset, file, info, ds.EmbedField, expand); err != nil {
			return macro.ErrorAt(fset, ds.EmbedField.Pos(), "%s", err.Error())
		}
		_ = pkg
	}
}

// RecognizeDeclSites finds anonymous embedded registered markers in type specs.
func RecognizeDeclSites(
	file *ast.File,
	info *types.Info,
	imports map[string]string,
	reg *macro.Registry,
) ([]declSite, error) {
	var sites []declSite
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
			if targetType == nil {
				continue
			}
			for i, field := range st.Fields.List {
				if len(field.Names) > 0 {
					continue
				}
				importPath, baseName, typeArgs, err := embeddedMarker(info, field.Type, imports)
				if err != nil || baseName == "" {
					continue
				}
				if !reg.HasMarker(importPath, baseName) {
					continue
				}
				syntaxID, ok := reg.SyntaxIDForMarker(importPath, baseName)
				if !ok || syntaxID == "" {
					continue
				}
				_ = syntaxID
				sites = append(sites, declSite{
					Target:           ts,
					TargetType:       targetType,
					EmbedIndex:       i,
					EmbedField:       field,
					MarkerImportPath: importPath,
					MarkerTypeName:   baseName,
					MarkerTypeArgs:   typeArgs,
					MacroTag:         macroTagFromField(field),
				})
			}
		}
	}
	return sites, nil
}

// declSite describes one anonymous embed of a registered marker type.
type declSite struct {
	Target           *ast.TypeSpec
	TargetType       types.Type
	EmbedIndex       int
	EmbedField       *ast.Field
	MarkerImportPath string
	MarkerTypeName   string
	MarkerTypeArgs   []types.Type
	MacroTag         macro.MacroTag
}

func macroTagFromField(field *ast.Field) macro.MacroTag {
	if field.Tag == nil {
		return nil
	}
	return macro.ParseMacroTag(field.Tag)
}

func embeddedMarker(
	info *types.Info,
	typ ast.Expr,
	imports map[string]string,
) (importPath, baseName string, typeArgs []types.Type, err error) {
	tv := info.TypeOf(typ)
	if tv == nil {
		return "", "", nil, nil
	}
	named, ok := types.Unalias(tv).(*types.Named)
	if !ok {
		return "", "", nil, nil
	}
	obj := named.Obj()
	if obj == nil || obj.Pkg() == nil {
		return "", "", nil, nil
	}
	importPath = obj.Pkg().Path()
	baseName = obj.Name()
	if named.TypeParams() != nil && named.TypeArgs() != nil {
		for i := 0; i < named.TypeArgs().Len(); i++ {
			typeArgs = append(typeArgs, named.TypeArgs().At(i))
		}
	}
	_ = imports
	return importPath, baseName, typeArgs, nil
}
