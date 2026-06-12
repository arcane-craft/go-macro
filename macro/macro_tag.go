package macro

import (
	"go/ast"
	"strings"
)

// MacroTag holds optional key=value pairs from a struct field's `macro:"..."` tag.
type MacroTag map[string]string

// ParseMacroTag parses `macro:"k=v"` or `macro:"k=v,k2=v2"` from a struct field tag.
func ParseMacroTag(tag *ast.BasicLit) MacroTag {
	if tag == nil {
		return nil
	}
	s := strings.Trim(tag.Value, "`")
	parts := strings.Split(s, " ")
	for _, p := range parts {
		if !strings.HasPrefix(p, "macro:") {
			continue
		}
		inner := strings.TrimPrefix(p, "macro:")
		inner = strings.Trim(inner, `"`)
		return parseMacroTagKV(inner)
	}
	return nil
}

func parseMacroTagKV(s string) MacroTag {
	out := make(MacroTag)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			out[part] = ""
		} else {
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
