package constraint_test

import (
	"testing"

	"github.com/arcane-craft/go-macro/internal/constraint"
)

func TestHasMacro(t *testing.T) {
	yes, err := constraint.HasMacro("macro && linux")
	if err != nil || !yes {
		t.Fatalf("HasMacro: yes=%v err=%v", yes, err)
	}
	no, err := constraint.HasMacro("linux")
	if err != nil || no {
		t.Fatalf("HasMacro linux: no=%v err=%v", no, err)
	}
	empty, err := constraint.HasMacro("  ")
	if err != nil || empty {
		t.Fatalf("empty: %v %v", empty, err)
	}
}

func TestHasOnlyIgnore(t *testing.T) {
	yes, err := constraint.HasOnlyIgnore("ignore")
	if err != nil || !yes {
		t.Fatalf("ignore only: %v %v", yes, err)
	}
	no, err := constraint.HasOnlyIgnore("macro")
	if err != nil || no {
		t.Fatalf("macro: %v %v", no, err)
	}
	both, err := constraint.HasOnlyIgnore("ignore && macro")
	if err != nil || both {
		t.Fatalf("both: %v %v", both, err)
	}
}

func TestComplementMacroConstraintErrors(t *testing.T) {
	_, err := constraint.ComplementMacroConstraint("linux")
	if err == nil {
		t.Fatal("expected error without macro")
	}
	got, err := constraint.ComplementMacroConstraint("")
	if err != nil || got != "!macro" {
		t.Fatalf("empty: got %q err=%v", got, err)
	}
}

func TestExtractBuildConstraint(t *testing.T) {
	const src = `//go:build macro && linux

package p
`
	expr, ok := constraint.ExtractBuildConstraint(src)
	if !ok || expr != "macro && linux" {
		t.Fatalf("go:build: ok=%v expr=%q", ok, expr)
	}
	const legacy = `// +build macro linux

package p
`
	expr2, ok := constraint.ExtractBuildConstraint(legacy)
	if !ok || expr2 != "macro && linux" {
		t.Fatalf("legacy: ok=%v expr=%q", ok, expr2)
	}
	_, ok = constraint.ExtractBuildConstraint("package p\n")
	if ok {
		t.Fatal("expected no constraint")
	}
}
