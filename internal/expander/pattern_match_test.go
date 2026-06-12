package expander

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func parseMatchFile(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	return fset, f
}

func findCall(t *testing.T, f *ast.File, name string) *ast.CallExpr {
	t.Helper()
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if invokedName(c.Fun) == name {
			if call == nil {
				call = c
			}
		}
		return true
	})
	if call == nil {
		t.Fatalf("call %q not found", name)
	}
	return call
}

func findEmbedField(t *testing.T, f *ast.File, marker string) *ast.Field {
	t.Helper()
	var field *ast.Field
	ast.Inspect(f, func(n ast.Node) bool {
		fl, ok := n.(*ast.Field)
		if !ok || fl.Names != nil {
			return true
		}
		if ie, ok := fl.Type.(*ast.IndexExpr); ok && typeExprInvokedName(ie.X) == marker {
			field = fl
		}
		return true
	})
	if field == nil {
		t.Fatalf("embed %q not found", marker)
	}
	return field
}

func TestMatchAssignDefine(t *testing.T) {
	src := `package p
func f() {
	x, err := Try(helper())
}
func helper() error { return nil }
`
	fset, f := parseMatchFile(t, src)
	call := findCall(t, f, "Try")
	site, err := ResolveSite(fset, f, call)
	if err != nil {
		t.Fatal(err)
	}
	binds, err := site.Match(`$lhs ... := Try($inner)`)
	if err != nil {
		t.Fatal(err)
	}
	lhs, ok := binds.Elems("lhs")
	if !ok || len(lhs) != 2 {
		t.Fatalf("lhs elems: ok=%v len=%d", ok, len(lhs))
	}
	inner, ok := binds.Get("inner")
	if !ok {
		t.Fatal("missing inner")
	}
	if _, ok := inner.Underlying().(*ast.CallExpr); !ok {
		t.Fatalf("inner type %T", inner.Underlying())
	}
	meta, ok := MatchMetaFromSite(site)
	if !ok || meta.MatchRoot != MatchRootStmt {
		t.Fatalf("meta: ok=%v root=%v", ok, meta.MatchRoot)
	}
	if _, ok := meta.MatchedSpan.(*ast.AssignStmt); !ok {
		t.Fatalf("span %T", meta.MatchedSpan)
	}
}

func TestMatchAssignPlain(t *testing.T) {
	src := `package p
func f() {
	x, err = Try(helper())
}
func helper() error { return nil }
`
	fset, f := parseMatchFile(t, src)
	call := findCall(t, f, "Try")
	site, _ := ResolveSite(fset, f, call)
	if _, err := site.Match(`$lhs ... = Try($inner)`); err != nil {
		t.Fatal(err)
	}
}

func TestMatchVarAssign(t *testing.T) {
	src := `package p
func f() {
	var x, err = Try(helper())
}
func helper() error { return nil }
`
	fset, f := parseMatchFile(t, src)
	call := findCall(t, f, "Try")
	site, _ := ResolveSite(fset, f, call)
	if _, err := site.Match(`var $lhs ... = Try($inner)`); err != nil {
		t.Fatal(err)
	}
	meta, _ := MatchMetaFromSite(site)
	if _, ok := meta.MatchedSpan.(*ast.DeclStmt); !ok {
		t.Fatalf("span %T", meta.MatchedSpan)
	}
}

func TestMatchReturnVals(t *testing.T) {
	src := `package p
func f() int {
	return a, Try(helper())
}
func helper() (int, error) { return 0, nil }
`
	fset, f := parseMatchFile(t, src)
	call := findCall(t, f, "Try")
	site, _ := ResolveSite(fset, f, call)
	binds, err := site.Match(`return $vals ... , Try($inner)`)
	if err != nil {
		t.Fatal(err)
	}
	vals, ok := binds.Elems("vals")
	if !ok || len(vals) != 1 {
		t.Fatalf("vals: ok=%v len=%d", ok, len(vals))
	}
}

func TestMatchReturnOnlyTry(t *testing.T) {
	src := `package p
func f() (int, error) {
	return Try(helper())
}
func helper() (int, error) { return 0, nil }
`
	fset, f := parseMatchFile(t, src)
	call := findCall(t, f, "Try")
	site, _ := ResolveSite(fset, f, call)
	if _, err := site.Match(`return $vals ... , Try($inner)`); err != nil {
		t.Fatal(err)
	}
}

func TestMatchCallVsExprStmt(t *testing.T) {
	src := `package p
func f() {
	Try(helper());
}
func helper() error { return nil }
`
	fset, f := parseMatchFile(t, src)
	call := findCall(t, f, "Try")

	siteCall, _ := ResolveSite(fset, f, call)
	if _, err := siteCall.Match(`Try($inner)`); err != nil {
		t.Fatal(err)
	}
	metaCall, _ := MatchMetaFromSite(siteCall)
	if metaCall.Plan[0].Replace.ContainerField != ContainerExprSlot {
		t.Fatalf("call plan field %v", metaCall.Plan[0].Replace.ContainerField)
	}

	siteStmt, _ := ResolveSite(fset, f, call)
	if _, err := siteStmt.Match(`Try($inner);`); err != nil {
		t.Fatal(err)
	}
	metaStmt, _ := MatchMetaFromSite(siteStmt)
	if metaStmt.Plan[0].Replace.ContainerField != ContainerBlockStmts {
		t.Fatalf("stmt plan field %v", metaStmt.Plan[0].Replace.ContainerField)
	}
}

func TestMatchInvokedName(t *testing.T) {
	src := `package p
import tr "example.com/try"
func f() {
	tr.Try(helper())
}
func helper() error { return nil }
`
	fset, f := parseMatchFile(t, src)
	call := findCall(t, f, "Try")
	site, _ := ResolveSite(fset, f, call)
	if _, err := site.Match(`Try($inner)`); err != nil {
		t.Fatal(err)
	}
}

func TestMatchDeclEmbedOrder(t *testing.T) {
	for _, src := range []string{
		`package p
type Item struct {
	Derive[Stringer]
	Name string ` + "`json:\"name\"`" + `
}
type Stringer interface { String() string }
`,
		`package p
type Item struct {
	Name string ` + "`json:\"name\"`" + `
	Derive[Stringer]
}
type Stringer interface { String() string }
`,
	} {
		fset, f := parseMatchFile(t, src)
		embed := findEmbedField(t, f, "Derive")
		site, err := ResolveSite(fset, f, embed)
		if err != nil {
			t.Fatal(err)
		}
		if !site.MacroPos().IsValid() {
			t.Fatal("MacroPos invalid")
		}
		binds, err := site.Match(`type $item struct { Derive[$iface] $field ... }`)
		if err != nil {
			t.Fatal(err)
		}
		fields, ok := binds.Elems("field")
		if !ok || len(fields) != 1 {
			t.Fatalf("field elems: ok=%v len=%d", ok, len(fields))
		}
		fld, ok := fields[0].Underlying().(*ast.Field)
		if !ok || fld.Tag == nil {
			t.Fatal("field missing tag")
		}
	}
}

func TestMatchTwoMacrosSameStmt(t *testing.T) {
	src := `package p
func f() {
	a, b := Try(f1()), Try(f2())
}
func f1() error { return nil }
func f2() error { return nil }
`
	fset, f := parseMatchFile(t, src)
	var calls []*ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok && invokedName(c.Fun) == "Try" {
			calls = append(calls, c)
		}
		return true
	})
	if len(calls) != 2 {
		t.Fatalf("calls %d", len(calls))
	}
	// Second Try in source order is f2 — pick rightmost by position
	var anchor *ast.CallExpr
	for _, c := range calls {
		if anchor == nil || c.Pos() > anchor.Pos() {
			anchor = c
		}
	}
	site, _ := ResolveSite(fset, f, anchor)
	binds, err := site.Match(`$lhs ... := Try($inner)`)
	if err != nil {
		t.Fatal(err)
	}
	inner, _ := binds.Get("inner")
	ic, ok := inner.Underlying().(*ast.CallExpr)
	if !ok || invokedName(ic.Fun) != "f2" {
		t.Fatalf("inner %#v", inner.Underlying())
	}
}
