package expander

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func TestValidateSpliceFailsBeforeApply(t *testing.T) {
	meta := MatchMeta{
		Plan: []SpliceStep{{
			Replace: &ReplaceInContainer{
				ContainerField: ContainerAssignRhs,
				Mode:           SpliceOneToOne,
			},
		}},
	}
	out := macro.WrapStmts([]ast.Stmt{&ast.ExprStmt{X: ast.NewIdent("x")}})
	if err := ValidateSplice(out, meta); err == nil {
		t.Fatal("expected validate error for stmt vs AssignRhs plan")
	}
}

func TestApplyAssignPlainViaPlan(t *testing.T) {
	const src = `package p
func f() {
	x, err = Try(helper())
}
func helper() error { return nil }
`
	fset, f := parsePlanFile(t, src)
	call := findTryCall(t, f)
	site, err := ResolveSite(fset, f, call)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := site.Match(`$lhs ... = Try($inner)`); err != nil {
		t.Fatal(err)
	}
	meta, ok := MatchMetaFromSite(site)
	if !ok {
		t.Fatal("no meta")
	}
	out := macro.WrapStmts([]ast.Stmt{
		&ast.AssignStmt{Tok: token.ASSIGN, Lhs: []ast.Expr{ast.NewIdent("x"), ast.NewIdent("err")}, Rhs: []ast.Expr{ast.NewIdent("v")}},
	})
	if err := ValidateSplice(out, meta); err != nil {
		t.Fatal(err)
	}
	if err := Apply(f, meta, out); err != nil {
		t.Fatal(err)
	}
}

func TestApplyVarViaPlan(t *testing.T) {
	const src = `package p
func f() {
	var x, err = Try(helper())
}
func helper() error { return nil }
`
	fset, f := parsePlanFile(t, src)
	call := findTryCall(t, f)
	site, _ := ResolveSite(fset, f, call)
	site.Match(`var $lhs ... = Try($inner)`)
	meta, _ := MatchMetaFromSite(site)
	out := macro.WrapStmts([]ast.Stmt{
		&ast.AssignStmt{Tok: token.DEFINE, Lhs: []ast.Expr{ast.NewIdent("x")}, Rhs: []ast.Expr{ast.NewIdent("v")}},
	})
	if err := ValidateSplice(out, meta); err != nil {
		t.Fatal(err)
	}
	if err := Apply(f, meta, out); err != nil {
		t.Fatal(err)
	}
}

func TestDerivePreservesUnmatchedMethods(t *testing.T) {
	const src = `package p
type Marker struct{}
type Item struct {
	Marker
	Name string
}
func (Item) Foo() {}
`
	fset, f := parsePlanFile(t, src)
	embed := findEmbedFieldPlan(t, f, "Marker")
	site, _ := ResolveSite(fset, f, embed)
	site.Match(`type $item struct { Marker $field ... }`)
	meta, _ := MatchMetaFromSite(site)
	itemTS, _ := meta.Bindings.Get("item")
	newTS := &ast.TypeSpec{
		Name: ast.NewIdent("Item"),
		Type: &ast.StructType{Fields: &ast.FieldList{List: []*ast.Field{{
			Names: []*ast.Ident{ast.NewIdent("Name")},
			Type:  ast.NewIdent("string"),
		}}}},
	}
	out := macro.WrapNode(newTS)
	if err := ValidateSplice(out, meta); err != nil {
		t.Fatal(err)
	}
	if err := Apply(f, meta, out); err != nil {
		t.Fatal(err)
	}
	var hasFoo bool
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if ok && fn.Name.Name == "Foo" {
			hasFoo = true
		}
	}
	if !hasFoo {
		t.Fatal("Foo method removed")
	}
	_ = itemTS
}

func parsePlanFile(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	return fset, f
}

func findTryCall(t *testing.T, f *ast.File) *ast.CallExpr {
	t.Helper()
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "Try" {
				call = c
			}
		}
		return true
	})
	if call == nil {
		t.Fatal("Try call not found")
	}
	return call
}

func findEmbedFieldPlan(t *testing.T, f *ast.File, marker string) *ast.Field {
	t.Helper()
	var field *ast.Field
	ast.Inspect(f, func(n ast.Node) bool {
		fl, ok := n.(*ast.Field)
		if !ok || fl.Names != nil {
			return true
		}
		if id, ok := fl.Type.(*ast.Ident); ok && id.Name == marker {
			field = fl
		}
		return true
	})
	if field == nil {
		t.Fatalf("embed %q not found", marker)
	}
	return field
}
