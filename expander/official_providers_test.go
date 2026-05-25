package expander

import "testing"

func TestOfficialProvidersForImports(t *testing.T) {
	got := officialProvidersForImports(map[string]bool{
		"github.com/arcane-craft/go-macro/try": true,
	})
	if len(got) != 1 || got[0].SyntaxID != "syntax-try" {
		t.Fatalf("want only try, got %+v", got)
	}
	if len(officialProvidersForImports(nil)) != 0 {
		t.Fatal("expected no providers without imports")
	}
}

func TestMergeProvidersExtraOverridesOfficial(t *testing.T) {
	imported := map[string]bool{"github.com/arcane-craft/go-macro/inline": true}
	custom := Provider{ImportPath: "github.com/arcane-craft/go-macro/inline", SyntaxID: "syntax-custom"}
	merged := mergeProviders(imported, []Provider{custom})
	if len(merged) != 1 || merged[0].SyntaxID != "syntax-custom" {
		t.Fatalf("extra should override official, got %+v", merged)
	}
}
