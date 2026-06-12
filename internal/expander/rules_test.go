package expander

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func TestSyntaxRulesInline(t *testing.T) {
	exp := macro.SyntaxRules(macro.Clause{
		Pattern:  `Inline($v)`,
		Template: `#v`,
	})
	fset, f, call := parseInlineCall(t, `package p
func f() int { return Inline(9) }
`)
	site, err := ResolveSite(fset, f, call)
	if err != nil {
		t.Fatal(err)
	}
	ctx := macro.NewContext(fset, nil)
	out, err := exp(ctx, site)
	if err != nil {
		t.Fatal(err)
	}
	e, err := out.ToExpr()
	if err != nil {
		t.Fatal(err)
	}
	if lit, ok := e.(*ast.BasicLit); !ok || lit.Value != "9" {
		t.Fatalf("got %#v", e)
	}
}

func TestSyntaxCaseDeriveRemovesEmbed(t *testing.T) {
	exp := macro.SyntaxCase(macro.Clause{
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
	})
	fset := token.NewFileSet()
	const src = `package p
type Marker struct{}
type Item struct {
	Marker
	Name string
}
`
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	embed := findEmbedFieldPlan(t, f, "Marker")
	site, err := ResolveSite(fset, f, embed)
	if err != nil {
		t.Fatal(err)
	}
	ctx := macro.NewContext(fset, nil)
	out, err := exp(ctx, site)
	if err != nil {
		t.Fatal(err)
	}
	meta, ok := MatchMetaFromSite(site)
	if !ok {
		t.Fatal("no meta")
	}
	if err := ValidateSplice(out, meta); err != nil {
		t.Fatal(err)
	}
	if err := Apply(f, meta, out); err != nil {
		t.Fatal(err)
	}
	st := f.Decls[1].(*ast.GenDecl).Specs[0].(*ast.TypeSpec).Type.(*ast.StructType)
	if len(st.Fields.List) != 1 {
		t.Fatalf("fields=%d", len(st.Fields.List))
	}
}

func parseInlineCall(t *testing.T, src string) (*token.FileSet, *ast.File, *ast.CallExpr) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "Inline" {
				call = c
			}
		}
		return true
	})
	if call == nil {
		t.Fatal("no Inline call")
	}
	return fset, f, call
}
