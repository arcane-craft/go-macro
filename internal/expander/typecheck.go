package expander

import (
	"go/ast"
	"go/importer"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// typecheckFile type-checks a freshly parsed file using dependency types from pkg.
func typecheckFile(pkg *packages.Package, file *ast.File) (*types.Info, *types.Package, error) {
	info := &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Instances:  make(map[*ast.Ident]types.Instance),
		Scopes:     make(map[ast.Node]*types.Scope),
		Implicits:  make(map[ast.Node]types.Object),
	}
	cfg := &types.Config{
		Importer: &pkgImporter{root: pkg},
	}
	typesPkg, err := cfg.Check(file.Name.Name, pkg.Fset, []*ast.File{file}, info)
	if err != nil {
		return nil, nil, err
	}
	return info, typesPkg, nil
}

type pkgImporter struct {
	root *packages.Package
}

func (i *pkgImporter) Import(path string) (*types.Package, error) {
	if p := findImportedPackage(i.root, path); p != nil && p.Types != nil {
		return p.Types, nil
	}
	return importer.Default().Import(path)
}
