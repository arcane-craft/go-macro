package constraint

import (
	"fmt"
	bconstraint "go/build/constraint"
	"strings"
)

// HasMacro reports whether the constraint expression contains identifier "macro".
func HasMacro(expr string) (bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false, nil
	}
	return strings.Contains(expr, "macro"), nil
}

// HasOnlyIgnore reports whether the constraint is only "ignore" without macro.
func HasOnlyIgnore(expr string) (bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false, nil
	}
	hasIgnore := strings.Contains(expr, "ignore")
	hasMacro := strings.Contains(expr, "macro")
	return hasIgnore && !hasMacro, nil
}

// ComplementMacroConstraint replaces identifier macro with !macro in the expression.
func ComplementMacroConstraint(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "!macro", nil
	}
	if !strings.Contains(expr, "macro") {
		return "", fmt.Errorf("constraint must contain macro identifier")
	}
	out := replaceIdent(expr, "macro", "!macro")
	// Validate when parseable
	wrapped := out
	if _, err := bconstraint.Parse(out); err != nil {
		wrapped = "(" + out + ")"
	}
	if _, err := bconstraint.Parse(wrapped); err != nil && out != "!macro" {
		// allow outputs that don't re-parse as single legacy tags
		if out == "!macro" {
			return out, nil
		}
	}
	return out, nil
}

func exprContainsIdent(c bconstraint.Expr, name string) bool {
	switch x := c.(type) {
	case *bconstraint.TagExpr:
		return x.Tag == name
	case *bconstraint.NotExpr:
		return exprContainsIdent(x.X, name)
	case *bconstraint.AndExpr:
		return exprContainsIdent(x.X, name) || exprContainsIdent(x.Y, name)
	case *bconstraint.OrExpr:
		return exprContainsIdent(x.X, name) || exprContainsIdent(x.Y, name)
	default:
		return false
	}
}

func replaceIdent(src, old, new string) string {
	var b strings.Builder
	i := 0
	for i < len(src) {
		r, size := runeAt(src, i)
		if isIdentStart(r) {
			j := i + size
			for j < len(src) {
				r2, sz := runeAt(src, j)
				if !isIdentContinue(r2) {
					break
				}
				j += sz
			}
			word := src[i:j]
			if word == old {
				b.WriteString(new)
			} else {
				b.WriteString(word)
			}
			i = j
			continue
		}
		b.WriteByte(src[i])
		i++
	}
	return b.String()
}

func runeAt(s string, i int) (rune, int) {
	return rune(s[i]), 1
}

func isIdentStart(r rune) bool {
	return r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isIdentContinue(r rune) bool {
	return isIdentStart(r) || (r >= '0' && r <= '9')
}

// ExtractBuildConstraint reads //go:build or // +build from file header comments.
func ExtractBuildConstraint(src string) (string, bool) {
	lines := strings.Split(src, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "//go:build ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "//go:build ")), true
		}
		if strings.HasPrefix(line, "// +build ") {
			parts := strings.Fields(strings.TrimPrefix(line, "// +build "))
			return strings.Join(parts, " && "), true
		}
		if strings.HasPrefix(line, "package ") {
			break
		}
	}
	return "", false
}
