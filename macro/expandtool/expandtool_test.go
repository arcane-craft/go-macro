package expandtool_test

import (
	"testing"

	"github.com/arcane-craft/go-macro/macro"
	"github.com/arcane-craft/go-macro/macro/expandtool"
)

func noopUnified(macro.Context, macro.Syntax) (macro.Syntax, error) {
	return nil, nil
}

func TestRegisteredReturnsCopy(t *testing.T) {
	const sid = "syntax-unified-copy"
	expandtool.Register(sid, noopUnified)
	t.Cleanup(func() { expandtool.Register(sid, nil) })

	got := expandtool.Registered()
	got["other"] = noopUnified
	if _, ok := expandtool.Registered()[sid]; !ok {
		t.Fatal("Registered missing syntax-id")
	}
	if _, ok := expandtool.Registered()["other"]; ok {
		t.Fatal("Registered must not reflect caller mutations")
	}
}

func TestRunNilLinkedUsesRegistered(t *testing.T) {
	const sid = "syntax-nil-linked"
	expandtool.Register(sid, noopUnified)
	t.Cleanup(func() { expandtool.Register(sid, nil) })

	if err := expandtool.Run([]string{"./..."}, nil); err != nil {
		t.Fatalf("Run with nil linked: %v", err)
	}
}

func TestRunDefaultPatterns(t *testing.T) {
	var nilArgs []string
	var emptyArgs []string
	linked := &macro.LinkedExpanders{Expand: map[string]macro.Expander{}}
	if err := expandtool.Run(nilArgs, linked); err != nil {
		t.Fatalf("Run(nil, {}): %v", err)
	}
	if err := expandtool.Run(emptyArgs, linked); err != nil {
		t.Fatalf("Run([]string{}, {}): %v", err)
	}
}

func TestRunPropagatesError(t *testing.T) {
	linked := &macro.LinkedExpanders{Expand: map[string]macro.Expander{}}
	err := expandtool.Run([]string{"\x00invalid"}, linked)
	if err == nil {
		t.Fatal("expected error from invalid pattern")
	}
}
