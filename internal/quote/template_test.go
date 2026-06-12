package quote_test

import (
	"testing"

	"github.com/arcane-craft/go-macro/internal/quote"
)

func TestBodyOnlyTemplate(t *testing.T) {
	t.Parallel()
	stmts, err := quote.Stmts(`x := 1`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(stmts) != 1 {
		t.Fatalf("len=%d", len(stmts))
	}

	// explicit @kind{ } wrapper remains supported
	stmts2, err := quote.Stmts(`@stmts{ y := 2 }`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(stmts2) != 1 {
		t.Fatalf("len=%d", len(stmts2))
	}
}

func TestParseTemplate_InvalidRoots(t *testing.T) {
	t.Parallel()
	cases := []string{
		"x := 1",
		"@stmts( x := 1 )",
		"@unknown{ x }",
		"@stmts{ x := 1 } extra",
		"@stmts{ unclosed",
	}
	for _, tpl := range cases {
		t.Run(tpl, func(t *testing.T) {
			t.Parallel()
			_, err := quote.Quote(tpl, nil)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestHoles_SkipCommentsAndStrings(t *testing.T) {
	t.Parallel()
	stmts, err := quote.Stmts(`// #not_a_hole
		x := 1`, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(stmts) != 1 {
		t.Fatalf("len=%d", len(stmts))
	}

	_, err = quote.Expr(`"#hash"`, nil)
	if err != nil {
		t.Fatal(err)
	}

	_, err = quote.Stmts(`s := "#x"`, nil)
	if err != nil {
		t.Fatal(err)
	}
}
