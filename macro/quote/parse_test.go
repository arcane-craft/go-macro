package quote_test

import (
	"strings"
	"testing"

	"github.com/arcane-craft/go-macro/macro/quote"
)

func TestParseFailures(t *testing.T) {
	t.Parallel()
	_, err := quote.Stmts(`:= broken`, nil)
	if err == nil {
		t.Fatal("expected parse error")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("expected parse context in error: %v", err)
	}
}

func TestKindMismatch(t *testing.T) {
	t.Parallel()
	_, err := quote.Stmts(`@expr{ 1 }`, nil)
	if err == nil {
		t.Fatal("expected kind mismatch")
	}
	if !strings.Contains(err.Error(), "expr") || !strings.Contains(err.Error(), "Stmts") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissingBinding(t *testing.T) {
	t.Parallel()
	_, err := quote.Expr(`#x`, nil)
	if err == nil {
		t.Fatal("expected missing binding")
	}
	if !strings.Contains(err.Error(), `hole "x"`) && !strings.Contains(err.Error(), "hole #x") {
		t.Fatalf("unexpected error: %v", err)
	}
}
