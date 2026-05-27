package expander

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/arcane-craft/go-macro/internal/codegen"
	"github.com/arcane-craft/go-macro/macro"
)

// ProviderLink is a provider package to register for expand.
type ProviderLink struct {
	ImportPath   string
	PackageName  string
	ExpanderName string
}

// DiscoverProviderLinks finds macro providers imported by macro-tagged files in patterns.
func DiscoverProviderLinks(patterns []string) ([]ProviderLink, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedDeps | packages.NeedImports,
		BuildFlags: []string{"-tags=macro"},
	}
	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, err
	}
	want := make(map[string]bool)
	for _, pkg := range pkgs {
		for _, filename := range pkg.GoFiles {
			src, err := os.ReadFile(filename)
			if err != nil {
				continue
			}
			if !codegen.IsMacroMainFile(string(src)) {
				continue
			}
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, filename, src, 0)
			if err != nil {
				continue
			}
			imports := BuildImportMap(f, pkg.PkgPath)
			for _, path := range imports {
				if path == "C" || strings.HasPrefix(path, "internal/") {
					continue
				}
				want[path] = true
			}
		}
	}
	byPath := make(map[string]*packages.Package)
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		if pkg.PkgPath != "" {
			byPath[pkg.PkgPath] = pkg
		}
	})
	var links []ProviderLink
	for path := range want {
		dep := byPath[path]
		if dep == nil {
			dep = loadProviderPackage(path)
		}
		if dep == nil || len(dep.Syntax) == 0 {
			continue
		}
		info, err := macro.ScanProviderFiles(dep.Syntax)
		if err != nil {
			continue
		}
		links = append(links, ProviderLink{
			ImportPath:   path,
			PackageName:  dep.Name,
			ExpanderName: info.ExpanderName,
		})
	}
	if len(links) == 0 {
		return nil, fmt.Errorf("macro expand: no macro providers found (import providers in //go:build macro files)")
	}
	sort.Slice(links, func(i, j int) bool { return links[i].ImportPath < links[j].ImportPath })
	return links, nil
}

func loadProviderPackage(importPath string) *packages.Package {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax,
	}, importPath)
	if err != nil || len(pkgs) == 0 {
		return nil
	}
	return pkgs[0]
}
