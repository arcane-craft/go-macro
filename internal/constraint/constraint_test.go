package constraint_test

import (
	"testing"

	"github.com/arcane-craft/go-macro/internal/constraint"
)

func TestComplementMacroConstraint(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"macro", "!macro"},
		{"macro && linux", "!macro && linux"},
		{"macro && (linux || darwin)", "!macro && (linux || darwin)"},
	}
	for _, tc := range tests {
		got, err := constraint.ComplementMacroConstraint(tc.in)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("%q: got %q want %q", tc.in, got, tc.want)
		}
	}
}
