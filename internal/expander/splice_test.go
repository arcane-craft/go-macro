package expander

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
)

func TestSpliceTryAssignIntoFunction(t *testing.T) {
	const src = `package p
import tr "example.com/try"
func helper() (int, error) { return 7, nil }
func f() (int, error) {
	x := tr.Try(helper())
	return x, nil
}
`
	fset, file, _, info, _ := parseSpliceTryFile(t, src)
	reg := registerTryProvider(t, fset)
	calls, err := RecognizeMacroCalls(file, info, map[string]string{"tr": "example.com/try"}, reg)
	if err != nil || len(calls) != 1 {
		t.Fatalf("calls: %v err=%v", calls, err)
	}
	mc := calls[0]
	site := classifySiteInFile(file, mc.Call)
	result := stubTryAssignExpandResult(mc.Call)
	if err := ApplyExpandResult(file, mc.Call, site, result); err != nil {
		t.Fatal(err)
	}
	body := formatFuncBody(t, fset, file, "f")
	for _, want := range []string{
		"_err1 :=", "helper()", "if _err1 != nil", "return 0, _err1", "x := _v2",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q\n%s", want, body)
		}
	}
	if strings.Contains(body, "try.Try") {
		t.Fatalf("macro call still present:\n%s", body)
	}
}

func TestApplyExpandResultSiteExpr(t *testing.T) {
	_, f, call := parseFileSplice(t, `package p
func f() int { return 1 + M(2) }
`)
	if err := ApplyExpandResult(f, call, macro.SiteExpr, macro.ExpandResult{Expr: ast.NewIdent("9")}); err != nil {
		t.Fatal(err)
	}
	ret := f.Decls[0].(*ast.FuncDecl).Body.List[0].(*ast.ReturnStmt)
	bin := ret.Results[0].(*ast.BinaryExpr)
	if id, ok := bin.Y.(*ast.Ident); !ok || id.Name != "9" {
		t.Fatalf("got rhs %#v", bin.Y)
	}
}

func TestApplyExpandResultPreservesTrailingStmts(t *testing.T) {
	_, f, call := parseFileSplice(t, `package p
func f() error {
	x := M(g())
	defer cleanup()
	return nil
}
func g() (int, error) { return 0, nil }
func cleanup() {}
`)
	block, _, _ := findEnclosingBlockStmt(f, call)
	if err := ApplyExpandResult(f, call, macro.SiteAssign, macro.ExpandResult{
		Stmts: []ast.Stmt{
			&ast.AssignStmt{Tok: token.DEFINE, Lhs: []ast.Expr{ast.NewIdent("a")}, Rhs: []ast.Expr{ast.NewIdent("b")}},
			&ast.AssignStmt{Tok: token.ASSIGN, Lhs: []ast.Expr{ast.NewIdent("x")}, Rhs: []ast.Expr{ast.NewIdent("a")}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if len(block.List) != 4 {
		t.Fatalf("want 4 stmts (expand x2 + defer + return), got %d", len(block.List))
	}
	if _, ok := block.List[2].(*ast.DeferStmt); !ok {
		t.Fatalf("stmt[2] = %T, want defer", block.List[2])
	}
}

func TestApplyExpandResultErrors(t *testing.T) {
	_, f, call := parseFileSplice(t, `package p
func f() { M(1) }
`)
	for _, tc := range []struct {
		site macro.CallSiteKind
		res  macro.ExpandResult
		frag string
	}{
		{macro.SiteAssign, macro.ExpandResult{}, "SiteAssign requires Stmts"},
		{macro.SiteReturn, macro.ExpandResult{}, "SiteReturn requires"},
		{macro.SiteExpr, macro.ExpandResult{}, "SiteExpr requires Expr"},
	} {
		err := ApplyExpandResult(f, call, tc.site, tc.res)
		if err == nil || !strings.Contains(err.Error(), tc.frag) {
			t.Fatalf("site %d: got %v want %q", tc.site, err, tc.frag)
		}
	}
}

func parseFileSplice(t *testing.T, src string) (*token.FileSet, *ast.File, *ast.CallExpr) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "t.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "M" {
				call = c
			}
		}
		return true
	})
	if call == nil {
		t.Fatal("no macro call M")
	}
	return fset, f, call
}

func parseSpliceTryFile(t *testing.T, src string) (*token.FileSet, *ast.File, *ast.CallExpr, *types.Info, *types.Package) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "p.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if sel, ok := c.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Try" {
				call = c
			}
		}
		return true
	})
	if call == nil {
		t.Fatal("no Try call")
	}
	providerPkg := types.NewPackage("example.com/try", "try")
	params := types.NewTuple(
		types.NewParam(0, nil, "", types.Typ[types.Int]),
		types.NewParam(0, nil, "", types.Universe.Lookup("error").Type()),
	)
	results := types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.Int]))
	tryFn := types.NewFunc(token.NoPos, providerPkg, "Try", types.NewSignature(nil, params, results, false))

	helperResults := types.NewTuple(
		types.NewParam(0, nil, "", types.Typ[types.Int]),
		types.NewParam(0, nil, "", types.Universe.Lookup("error").Type()),
	)
	helperFn := types.NewFunc(token.NoPos, types.NewPackage("p", "p"), "helper", types.NewSignature(nil, nil, helperResults, false))

	userPkg := types.NewPackage("p", "p")
	sel := call.Fun.(*ast.SelectorExpr)
	helperCall := call.Args[0].(*ast.CallExpr)

	var fn *ast.FuncDecl
	ast.Inspect(f, func(n ast.Node) bool {
		if fd, ok := n.(*ast.FuncDecl); ok && fd.Name.Name == "f" {
			fn = fd
		}
		return true
	})
	outerResults := types.NewTuple(
		types.NewParam(0, nil, "", types.Typ[types.Int]),
		types.NewParam(0, nil, "", types.Universe.Lookup("error").Type()),
	)
	fFn := types.NewFunc(token.NoPos, userPkg, "f", types.NewSignature(nil, nil, outerResults, false))

	info := &types.Info{
		Defs: map[*ast.Ident]types.Object{
			fn.Name: fFn,
		},
		Uses: map[*ast.Ident]types.Object{
			sel.X.(*ast.Ident):          types.NewPkgName(token.NoPos, userPkg, "tr", providerPkg),
			sel.Sel:                     tryFn,
			helperCall.Fun.(*ast.Ident): helperFn,
		},
		Types: map[ast.Expr]types.TypeAndValue{
			helperCall: {Type: helperResults},
		},
	}
	return fset, f, call, info, userPkg
}

func noopExpander(macro.Context, *ast.CallExpr) (macro.ExpandResult, error) {
	return macro.ExpandResult{}, nil
}

// stubTryAssignExpandResult builds Try-assign expansion for k=1 (used by splice tests only).
func stubTryAssignExpandResult(call *ast.CallExpr) macro.ExpandResult {
	expr := call.Args[0]
	errIdent := ast.NewIdent("_err1")
	valIdent := ast.NewIdent("_v2")
	assign := &ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{valIdent, errIdent},
		Rhs: []ast.Expr{expr},
	}
	ifStmt := &ast.IfStmt{
		Cond: &ast.BinaryExpr{
			X:  errIdent,
			Op: token.NEQ,
			Y:  ast.NewIdent("nil"),
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{ast.NewIdent("0"), errIdent},
				},
			},
		},
	}
	success := &ast.AssignStmt{
		Tok: token.DEFINE,
		Lhs: []ast.Expr{ast.NewIdent("x")},
		Rhs: []ast.Expr{valIdent},
	}
	return macro.ExpandResult{Stmts: []ast.Stmt{assign, ifStmt, success}}
}

func registerTryProvider(t *testing.T, fset *token.FileSet) *macro.Registry {
	t.Helper()
	pf, _ := parser.ParseFile(fset, "try.go", `package try
//macro: syntax-try
func Try[T any](v T, err error) T { panic("x") }
`, parser.ParseComments)
	reg := macro.NewRegistry()
	if err := reg.RegisterProvider("example.com/try", []*ast.File{pf}, "syntax-try", noopExpander); err != nil {
		t.Fatal(err)
	}
	return reg
}

func formatFuncBody(t *testing.T, fset *token.FileSet, file *ast.File, name string) string {
	t.Helper()
	var fn *ast.FuncDecl
	for _, d := range file.Decls {
		if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == name {
			fn = f
		}
	}
	var buf bytes.Buffer
	cfg := &printer.Config{Mode: printer.RawFormat, Tabwidth: 8}
	for _, s := range fn.Body.List {
		_ = cfg.Fprint(&buf, fset, s)
	}
	return buf.String()
}
