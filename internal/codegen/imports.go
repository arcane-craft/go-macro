package codegen

import (
	"go/ast"
	"strings"
)

func filterUnusedImports(file *ast.File) {
	used := make(map[string]bool)
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.SelectorExpr:
			if id, ok := x.X.(*ast.Ident); ok {
				used[id.Name] = true
			}
		}
		return true
	})
	var kept []*ast.ImportSpec
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		local := path
		if imp.Name != nil {
			local = imp.Name.Name
		} else if i := strings.LastIndex(path, "/"); i >= 0 {
			local = path[i+1:]
		}
		if strings.HasSuffix(path, "/try") && !used[local] {
			continue
		}
		kept = append(kept, imp)
	}
	file.Imports = kept
}
