package expandtool_test

import (
	"go/ast"
	"testing"

	"github.com/arcane-craft/go-macro/macro"
	"github.com/arcane-craft/go-macro/macro/expandtool"
)

func noopExpand(macro.CallContext, *ast.CallExpr) (macro.CallExpandResult, error) {
	return macro.CallExpandResult{}, nil
}

func TestRegisteredCallReturnsCopy(t *testing.T) {
	const sid = "syntax-copy-test"
	expandtool.RegisterCall(sid, noopExpand)
	t.Cleanup(func() { expandtool.RegisterCall(sid, nil) })

	got := expandtool.RegisteredCall()
	got["other"] = noopExpand
	if _, ok := expandtool.RegisteredCall()[sid]; !ok {
		t.Fatal("RegisteredCall missing registered syntax-id")
	}
	if _, ok := expandtool.RegisteredCall()["other"]; ok {
		t.Fatal("RegisteredCall must not reflect caller mutations")
	}
}

func TestRunNilLinkedUsesRegistered(t *testing.T) {
	const sid = "syntax-nil-linked"
	expandtool.RegisterCall(sid, noopExpand)
	t.Cleanup(func() { expandtool.RegisterCall(sid, nil) })

	if err := expandtool.Run([]string{"./..."}, nil); err != nil {
		t.Fatalf("Run with nil linked: %v", err)
	}
}

func TestRunDefaultPatterns(t *testing.T) {
	var nilArgs []string
	var emptyArgs []string
	linked := &macro.LinkedExpanders{Call: map[string]macro.CallExpander{}}
	if err := expandtool.Run(nilArgs, linked); err != nil {
		t.Fatalf("Run(nil, {}): %v", err)
	}
	if err := expandtool.Run(emptyArgs, linked); err != nil {
		t.Fatalf("Run([]string{}, {}): %v", err)
	}
}

func TestRunPropagatesError(t *testing.T) {
	linked := &macro.LinkedExpanders{Call: map[string]macro.CallExpander{}}
	err := expandtool.Run([]string{"\x00invalid"}, linked)
	if err == nil {
		t.Fatal("expected error from invalid pattern")
	}
}
