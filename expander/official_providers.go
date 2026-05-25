package expander

import (
	"sort"

	"github.com/arcane-craft/go-macro/inline"
	"github.com/arcane-craft/go-macro/try"
)

// officialMacroLibraries are maintained macro providers in this module.
// They participate in expand only when the macro main file imports the package.
var officialMacroLibraries = []Provider{
	{ImportPath: "github.com/arcane-craft/go-macro/inline", SyntaxID: "syntax-inline", Expand: inline.InlineExpand},
	{ImportPath: "github.com/arcane-craft/go-macro/try", SyntaxID: "syntax-try", Expand: try.TryExpand},
}

// officialProvidersForImports returns official libraries whose import path appears in imported.
func officialProvidersForImports(imported map[string]bool) []Provider {
	var out []Provider
	for _, p := range officialMacroLibraries {
		if imported[p.ImportPath] {
			out = append(out, p)
		}
	}
	return out
}

// mergeProviders combines extra providers (e.g. from a custom expand tool) with official
// libraries selected by import. Later entries with the same ImportPath override earlier ones.
func mergeProviders(imported map[string]bool, extra []Provider) []Provider {
	byPath := make(map[string]Provider)
	for _, p := range officialProvidersForImports(imported) {
		byPath[p.ImportPath] = p
	}
	for _, p := range extra {
		if imported[p.ImportPath] {
			byPath[p.ImportPath] = p
		}
	}
	out := make([]Provider, 0, len(byPath))
	for path := range byPath {
		out = append(out, byPath[path])
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ImportPath < out[j].ImportPath })
	return out
}
