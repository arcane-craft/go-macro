package expander

import (
	"fmt"
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
		site := sites[0]
		syntaxID, expand, ok := e.Registry.LookupDeclMarker(site.MarkerImportPath, site.MarkerTypeName)
		if !ok {
			return macro.ErrorAt(fset, site.EmbedField.Pos(), "decl macro %q not linked", site.MarkerTypeName)
		}
		ctx := macro.NewDeclContext(fset, file, info, pkg, site, syntaxID)
		result, err := expand(ctx, site)
		if err != nil {
			return err
		}
		if err := macro.ValidateDeclExpandResult(ctx, result); err != nil {
			return macro.ErrorAt(fset, site.EmbedField.Pos(), "%s", err.Error())
		}
		if err := ApplyDeclExpandResult(file, site.Target, result); err != nil {
			return macro.ErrorAt(fset, site.EmbedField.Pos(), "%s", err.Error())
		}
	}
}

// RecognizeDeclSites finds anonymous embedded registered markers in type specs.
func RecognizeDeclSites(
	file *ast.File,
	info *types.Info,
	imports map[string]string,
	reg *macro.Registry,
) ([]macro.DeclSite, error) {
	var sites []macro.DeclSite
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
				syntaxID, _, ok := reg.LookupDeclMarker(importPath, baseName)
				if !ok || syntaxID == "" {
					continue
				}
				_ = syntaxID
				sites = append(sites, macro.DeclSite{
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

// ApplyDeclExpandResult replaces struct fields and Target methods in file.
func ApplyDeclExpandResult(file *ast.File, target *ast.TypeSpec, result macro.DeclExpandResult) error {
	st, ok := target.Type.(*ast.StructType)
	if !ok {
		return fmt.Errorf("macro: decl target %s is not a struct", target.Name.Name)
	}
	st.Fields = &ast.FieldList{List: result.Fields}
	removeTargetMethods(file, target.Name.Name)
	insertTargetMethods(file, target, result.Methods)
	return nil
}

func removeTargetMethods(file *ast.File, typeName string) {
	var kept []ast.Decl
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && fn.Recv != nil && len(fn.Recv.List) > 0 {
			if recvNameForDecl(fn.Recv.List[0].Type) == typeName {
				continue
			}
		}
		kept = append(kept, decl)
	}
	file.Decls = kept
}

func insertTargetMethods(file *ast.File, target *ast.TypeSpec, methods []*ast.FuncDecl) {
	idx := -1
	for i, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			if spec == target {
				idx = i
				break
			}
		}
	}
	if idx < 0 {
		file.Decls = append(file.Decls, declsFromMethods(methods)...)
		return
	}
	var out []ast.Decl
	out = append(out, file.Decls[:idx+1]...)
	out = append(out, declsFromMethods(methods)...)
	out = append(out, file.Decls[idx+1:]...)
	file.Decls = out
}

func declsFromMethods(methods []*ast.FuncDecl) []ast.Decl {
	out := make([]ast.Decl, len(methods))
	for i, m := range methods {
		out[i] = m
	}
	return out
}

func recvNameForDecl(t ast.Expr) string {
	switch r := t.(type) {
	case *ast.Ident:
		return r.Name
	case *ast.StarExpr:
		if id, ok := r.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.IndexExpr:
		return recvNameForDecl(r.X)
	case *ast.IndexListExpr:
		return recvNameForDecl(r.X)
	}
	return ""
}
