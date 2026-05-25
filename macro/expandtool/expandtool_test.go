package expandtool_test

import (
	"go/ast"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
	"github.com/arcane-craft/go-macro/macro/expandtool"
)

func noopExpand(macro.Context, *ast.CallExpr) (macro.ExpandResult, error) {
	return macro.ExpandResult{}, nil
}

func TestRegisteredReturnsCopy(t *testing.T) {
	const path = "example.com/copy-test"
	expandtool.Register(path, noopExpand)
	t.Cleanup(func() { expandtool.Register(path, nil) })

	got := expandtool.Registered()
	got["other"] = noopExpand
	if _, ok := expandtool.Registered()[path]; !ok {
		t.Fatal("Registered missing registered path")
	}
	if _, ok := expandtool.Registered()["other"]; ok {
		t.Fatal("Registered must not reflect caller mutations")
	}
}

func TestRunNilLinkedUsesRegistered(t *testing.T) {
	const path = "example.com/nil-linked"
	expandtool.Register(path, noopExpand)
	t.Cleanup(func() { expandtool.Register(path, nil) })

	// Vacuous expand: no macro main files under this package's test scope.
	if err := expandtool.Run([]string{"./..."}, nil); err != nil {
		t.Fatalf("Run with nil linked: %v", err)
	}
}

func TestRunDefaultPatterns(t *testing.T) {
	var nilArgs []string
	var emptyArgs []string
	if err := expandtool.Run(nilArgs, map[string]macro.Expander{}); err != nil {
		t.Fatalf("Run(nil, {}): %v", err)
	}
	if err := expandtool.Run(emptyArgs, map[string]macro.Expander{}); err != nil {
		t.Fatalf("Run([]string{}, {}): %v", err)
	}
}

func TestRunPropagatesError(t *testing.T) {
	err := expandtool.Run([]string{"\x00invalid"}, map[string]macro.Expander{})
	if err == nil {
		t.Fatal("expected error from invalid pattern")
	}
}
