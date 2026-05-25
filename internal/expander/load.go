package expander

import (
	"fmt"
	"go/ast"
	"go/parser"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/arcane-craft/go-macro/internal/codegen"
	"github.com/arcane-craft/go-macro/internal/constraint"
	"github.com/arcane-craft/go-macro/macro"
)

// ExpandPackages expands macro-tagged main files and writes *_macro_gen.go.
// linked maps provider import paths to Expander functions; only paths both linked and
// imported by the macro main package are activated.
func ExpandPackages(patterns []string, linked map[string]macro.Expander) error {
	expandedFiles := make(map[string]bool)
	nameCfg := &packages.Config{Mode: packages.NeedName}
	roots, err := packages.Load(nameCfg, patterns...)
	if err != nil {
		return err
	}
	rootPaths := make(map[string]bool)
	for _, r := range roots {
		if r.PkgPath != "" {
			rootPaths[r.PkgPath] = true
		}
	}
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports | packages.NeedDeps,
		BuildFlags: []string{"-tags=macro"},
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return err
	}
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, e := range pkg.Errors {
			if e.Kind == packages.ListError {
				continue
			}
		}
	})
	for _, pkg := range pkgs {
		if strings.HasSuffix(pkg.PkgPath, ".test") {
			continue
		}
		if len(rootPaths) > 0 && !rootPaths[pkg.PkgPath] {
			continue
		}
		if err := expandOnePackage(pkg, linked, expandedFiles); err != nil {
			return err
		}
	}
	return nil
}

func expandOnePackage(pkg *packages.Package, linked map[string]macro.Expander, expandedFiles map[string]bool) error {
	if pkg.Types == nil || pkg.Fset == nil {
		return nil
	}
	engine := &Engine{Registry: macro.NewRegistry()}

	imported := importedProviderPaths(pkg)
	active := make(map[string]macro.Expander)
	for path, expand := range linked {
		if imported[path] {
			active[path] = expand
		}
	}
	filesByPath := make(map[string][]*ast.File)
	for path := range active {
		files, err := providerFiles(pkg, path)
		if err != nil {
			return err
		}
		filesByPath[path] = files
	}
	if len(active) > 0 {
		if err := engine.RegisterLinked(active, filesByPath); err != nil {
			return err
		}
	}

	seenGoFile := make(map[string]bool)
	for _, filename := range pkg.GoFiles {
		absFile, _ := filepath.Abs(filename)
		if seenGoFile[absFile] {
			continue
		}
		seenGoFile[absFile] = true
		if codegen.IsGenFile(filename) || codegen.IsGenFile(absFile) {
			continue
		}
		if expandedFiles[absFile] {
			continue
		}
		src, err := os.ReadFile(filename)
		if err != nil {
			return err
		}
		if !codegen.IsMacroMainFile(string(src)) {
			continue
		}
		if err := codegen.ValidateMainFile(string(src)); err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}
		file, err := parser.ParseFile(pkg.Fset, filename, src, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse %s: %w", filename, err)
		}
		info, typesPkg, err := typecheckFile(pkg, file)
		if err != nil {
			return fmt.Errorf("typecheck %s: %w", filename, err)
		}
		imports := BuildImportMap(file, pkg.PkgPath)
		if err := engine.ExpandFile(pkg.Fset, file, info, typesPkg, imports); err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}
		mainExpr, _ := constraint.ExtractBuildConstraint(string(src))
		genExpr, err := codegen.GenConstraint(mainExpr)
		if err != nil {
			return fmt.Errorf("%s: %w", filename, err)
		}
		if err := codegen.WriteGenFile(filename, genExpr, pkg.Fset, file); err != nil {
			return err
		}
		expandedFiles[absFile] = true
	}
	return nil
}

func importedProviderPaths(pkg *packages.Package) map[string]bool {
	m := make(map[string]bool)
	for path := range pkg.Imports {
		m[path] = true
	}
	return m
}

func providerFiles(pkg *packages.Package, importPath string) ([]*ast.File, error) {
	dep := findImportedPackage(pkg, importPath)
	if dep == nil {
		return nil, fmt.Errorf("provider %s not found in dependencies", importPath)
	}
	var files []*ast.File
	for _, f := range dep.Syntax {
		files = append(files, f)
	}
	return files, nil
}

func findImportedPackage(pkg *packages.Package, importPath string) *packages.Package {
	if p := pkg.Imports[importPath]; p != nil {
		return p
	}
	seen := make(map[*packages.Package]bool)
	var walk func(*packages.Package) *packages.Package
	walk = func(p *packages.Package) *packages.Package {
		if p == nil || seen[p] {
			return nil
		}
		seen[p] = true
		if p.PkgPath == importPath {
			return p
		}
		for _, imp := range p.Imports {
			if q := walk(imp); q != nil {
				return q
			}
		}
		return nil
	}
	return walk(pkg)
}

// ModuleRoot returns the directory containing go.mod for the first pattern package.
func ModuleRoot(patterns []string) (string, error) {
	cfg := &packages.Config{Mode: packages.NeedModule}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return "", err
	}
	for _, p := range pkgs {
		if p.Module != nil {
			return filepath.Dir(p.Module.GoMod), nil
		}
	}
	return "", fmt.Errorf("no module found for %v", patterns)
}
