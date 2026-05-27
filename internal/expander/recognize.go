package expander

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"github.com/arcane-craft/go-macro/macro"
)

// MacroCall describes a recognized macro invocation.
type MacroCall struct {
	Call       *ast.CallExpr
	StubName   string
	SyntaxID   string
	ImportPath string
}

// RecognizeMacroCalls finds macro calls in a file using types info and registry.
func RecognizeMacroCalls(
	file *ast.File,
	info *types.Info,
	imports map[string]string, // local import name -> package path
	reg *macro.Registry,
) ([]MacroCall, error) {
	var out []MacroCall
	seen := make(map[token.Pos]bool)
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		stub, pkgPath, ok := resolveStubCall(call, info, imports)
		if !ok {
			return true
		}
		if !reg.HasStub(pkgPath, stub) {
			return true
		}
		syntaxID, _, ok := reg.Lookup(pkgPath, stub)
		if !ok {
			return true
		}
		pos := call.Pos()
		if seen[pos] {
			return true
		}
		seen[pos] = true
		out = append(out, MacroCall{
			Call:       call,
			StubName:   stub,
			SyntaxID:   syntaxID,
			ImportPath: pkgPath,
		})
		return true
	})
	return out, nil
}

func resolveStubCall(call *ast.CallExpr, info *types.Info, imports map[string]string) (stubName, pkgPath string, ok bool) {
	fun := unwrapParen(call.Fun)
	switch f := fun.(type) {
	case *ast.Ident:
		obj := info.Uses[f]
		if obj == nil {
			return "", "", false
		}
		return objectStub(obj)
	case *ast.SelectorExpr:
		if !isPackageSelector(f, info) {
			return "", "", false
		}
		obj := info.Uses[f.Sel]
		if obj == nil {
			return "", "", false
		}
		return objectStub(obj)
	default:
		return "", "", false
	}
}

func objectStub(obj types.Object) (stubName, pkgPath string, ok bool) {
	fn, ok := obj.(*types.Func)
	if !ok {
		return "", "", false
	}
	typ := fn.Type()
	sig, ok := typ.(*types.Signature)
	if !ok || sig == nil {
		return "", "", false
	}
	if sig.Recv() != nil {
		return "", "", false
	}
	pkg := fn.Pkg()
	if pkg == nil {
		return "", "", false
	}
	return fn.Name(), pkg.Path(), true
}

func isPackageSelector(sel *ast.SelectorExpr, info *types.Info) bool {
	id, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	obj := info.Uses[id]
	if obj == nil {
		return false
	}
	_, ok = obj.(*types.PkgName)
	return ok
}

func unwrapParen(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}

// BuildImportMap returns local name -> import path from file imports.
func BuildImportMap(file *ast.File, pkgPath string) map[string]string {
	m := make(map[string]string)
	for _, spec := range file.Imports {
		path := trimImportPath(spec.Path.Value)
		local := path
		if spec.Name != nil {
			local = spec.Name.Name
			if local == "." {
				// dot import: stubs appear as bare Ident
				m["."] = path
				continue
			}
			if local == "_" {
				continue
			}
		} else {
			local = defaultImportName(path)
		}
		m[local] = path
	}
	return m
}

func trimImportPath(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func defaultImportName(path string) string {
	i := len(path) - 1
	for i >= 0 && path[i] != '/' {
		i--
	}
	name := path[i+1:]
	if name == "" {
		return "unknown"
	}
	return name
}

// ValidateRecognize is used in tests for expected recognition.
func ValidateRecognize(err error) error {
	if err != nil {
		return fmt.Errorf("recognize: %w", err)
	}
	return nil
}
