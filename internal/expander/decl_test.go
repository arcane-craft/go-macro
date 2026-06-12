package expander_test

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/arcane-craft/go-macro/internal/expander"
	"github.com/arcane-craft/go-macro/macro"
)

func TestExpandDeclMacrosRemovesEmbed(t *testing.T) {
	fset := token.NewFileSet()
	const src = `package p

type Marker struct{}

type Item struct {
	Marker
	Name string
}
`
	f, err := parser.ParseFile(fset, "f.go", src, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}
	cfg := &types.Config{Importer: importer.Default()}
	pkg, err := cfg.Check("p", fset, []*ast.File{f}, info)
	if err != nil {
		t.Fatal(err)
	}
	reg := macro.NewRegistry()
	reg.RegisterMarker("p", "Marker", "test-syntax")
	reg.RegisterExpander("test-syntax", macro.SyntaxCase(macro.Clause{
		Pattern: `type $item struct { Marker $field ... }`,
		Transform: func(ctx macro.Context, site macro.Syntax, binds macro.Bindings) (macro.Syntax, error) {
			fields, _ := binds.Elems("field")
			var list []*ast.Field
			for _, f := range fields {
				list = append(list, f.Underlying().(*ast.Field))
			}
			itemTS, _ := binds.Get("item")
			ts := itemTS.Underlying().(*ast.TypeSpec)
			return macro.WrapNode(&ast.TypeSpec{
				Name: ts.Name,
				Type: &ast.StructType{Fields: &ast.FieldList{List: list}},
			}), nil
		},
	}))
	engine := &expander.Engine{Registry: reg}
	if err := engine.ExpandDeclMacros(fset, f, info, pkg, map[string]string{}); err != nil {
		t.Fatal(err)
	}
	var ts *ast.TypeSpec
	for _, d := range f.Decls {
		gen, ok := d.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			if t, ok := spec.(*ast.TypeSpec); ok && t.Name.Name == "Item" {
				ts = t
				break
			}
		}
	}
	if ts == nil {
		t.Fatal("Item type not found")
	}
	st := ts.Type.(*ast.StructType)
	if len(st.Fields.List) != 1 || st.Fields.List[0].Names[0].Name != "Name" {
		t.Fatalf("fields after expand: %+v", st.Fields.List)
	}
}
