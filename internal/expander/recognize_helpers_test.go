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

func setupProviderReg(t *testing.T, fset *token.FileSet, providerPath string) *macro.Registry {
	t.Helper()
	pf, err := parser.ParseFile(fset, "p.go", `package macprov
import "github.com/arcane-craft/go-macro/macro"
//macro: syntax-test
func MacroStub(int) int { panic("x") }
//macro: syntax-test
func MacroExpand(ctx macro.Context, site macro.Syntax) (macro.Syntax, error) {
	return nil, nil
}
`, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}
	reg := macro.NewRegistry()
	if err := reg.RegisterProviderSources(providerPath, []*ast.File{pf}); err != nil {
		t.Fatal(err)
	}
	reg.RegisterExpander("syntax-test", func(macro.Context, macro.Syntax) (macro.Syntax, error) {
		return nil, nil
	})
	return reg
}

func typecheckWithProvider(t *testing.T, fset *token.FileSet, f *ast.File, providerPath, stubName string) *types.Info {
	t.Helper()
	providerPkg := types.NewPackage(providerPath, "macprov")
	params := types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.Int]))
	results := types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.Int]))
	sig := types.NewSignature(nil, params, results, false)
	stub := types.NewFunc(token.NoPos, providerPkg, stubName, sig)

	var call *ast.CallExpr
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			switch fun := c.Fun.(type) {
			case *ast.Ident:
				if fun.Name == stubName {
					call = c
				}
			case *ast.SelectorExpr:
				if fun.Sel.Name == stubName {
					call = c
				}
			case *ast.ParenExpr:
				if sel, ok := fun.X.(*ast.SelectorExpr); ok && sel.Sel.Name == stubName {
					call = c
				}
			}
		}
		return true
	})
	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
		Defs: make(map[*ast.Ident]types.Object),
	}
	cfg := &types.Config{
		Importer: importer.Default(),
	}
	if _, err := cfg.Check("u", fset, []*ast.File{f}, info); err == nil && call != nil {
		return info
	}
	// Fallback: synthetic Uses for selector/dot-import tests.
	if call != nil {
		switch fun := call.Fun.(type) {
		case *ast.Ident:
			info.Uses[fun] = stub
		case *ast.SelectorExpr:
			impPkg := types.NewPackage("u", "u")
			info.Uses[fun.X.(*ast.Ident)] = types.NewPkgName(token.NoPos, impPkg, fun.X.(*ast.Ident).Name, providerPkg)
			info.Uses[fun.Sel] = stub
		case *ast.ParenExpr:
			if sel, ok := fun.X.(*ast.SelectorExpr); ok {
				impPkg := types.NewPackage("u", "u")
				info.Uses[sel.X.(*ast.Ident)] = types.NewPkgName(token.NoPos, impPkg, sel.X.(*ast.Ident).Name, providerPkg)
				info.Uses[sel.Sel] = stub
			}
		}
	}
	return info
}

// typecheckFileWithStubUses type-checks f or wires provider stub selectors/idents for tests.
func typecheckFileWithStubUses(t *testing.T, fset *token.FileSet, f *ast.File, providerPath, stubName string) *types.Info {
	t.Helper()
	providerPkg := types.NewPackage(providerPath, "macprov")
	params := types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.Int]))
	results := types.NewTuple(types.NewParam(0, nil, "", types.Typ[types.Int]))
	sig := types.NewSignature(nil, params, results, false)
	stub := types.NewFunc(token.NoPos, providerPkg, stubName, sig)

	info := &types.Info{
		Uses: make(map[*ast.Ident]types.Object),
		Defs: make(map[*ast.Ident]types.Object),
	}
	cfg := &types.Config{Importer: importer.Default()}
	if _, err := cfg.Check("u", fset, []*ast.File{f}, info); err == nil {
		return info
	}
	impPkg := types.NewPackage("u", "u")
	ast.Inspect(f, func(n ast.Node) bool {
		switch e := n.(type) {
		case *ast.SelectorExpr:
			if e.Sel.Name != stubName {
				return true
			}
			if id, ok := e.X.(*ast.Ident); ok {
				info.Uses[id] = types.NewPkgName(token.NoPos, impPkg, id.Name, providerPkg)
				info.Uses[e.Sel] = stub
			}
		case *ast.Ident:
			if e.Name == stubName {
				info.Uses[e] = stub
			}
		}
		return true
	})
	return info
}

func TestModuleRoot(t *testing.T) {
	root, err := expander.ModuleRoot([]string{"github.com/arcane-craft/go-macro/internal/expander"})
	if err != nil {
		t.Fatal(err)
	}
	if root == "" {
		t.Fatal("empty module root")
	}
}

