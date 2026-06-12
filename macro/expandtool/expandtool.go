package expandtool

import (
	"os"
	"sync"

	"github.com/arcane-craft/go-macro/internal/expander"
	"github.com/arcane-craft/go-macro/macro"
)

var (
	expandMu       sync.RWMutex
	expandRegistry = make(map[string]macro.Expander)
)

// Register records a unified Expander for syntaxID.
func Register(syntaxID string, expand macro.Expander) {
	expandMu.Lock()
	defer expandMu.Unlock()
	expandRegistry[syntaxID] = expand
}

// Registered returns a copy of the Expander registration table.
func Registered() map[string]macro.Expander {
	expandMu.RLock()
	defer expandMu.RUnlock()
	out := make(map[string]macro.Expander, len(expandRegistry))
	for k, v := range expandRegistry {
		out[k] = v
	}
	return out
}

// Run expands macro-tagged files. Empty args default to ./...; nil linked uses Registered().
func Run(args []string, linked *macro.LinkedExpanders) error {
	patterns := []string{"./..."}
	if len(args) > 0 {
		patterns = args
	}
	if linked == nil {
		linked = &macro.LinkedExpanders{Expand: Registered()}
	}
	return expander.ExpandPackages(patterns, linked)
}

// Main runs Run(os.Args[1:], nil) and exits non-zero on error.
func Main() {
	if err := Run(os.Args[1:], nil); err != nil {
		ExitWithError(err)
	}
}
