package expandtool

import (
	"fmt"
	"os"
	"sync"

	"github.com/arcane-craft/go-macro/internal/expander"
	"github.com/arcane-craft/go-macro/macro"
)

var (
	mu       sync.RWMutex
	registry = make(map[string]macro.Expander)
)

// Register records an Expander for importPath. Later Register calls override the same path.
func Register(importPath string, expand macro.Expander) {
	mu.Lock()
	defer mu.Unlock()
	registry[importPath] = expand
}

// Registered returns a copy of the current registration table.
func Registered() map[string]macro.Expander {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]macro.Expander, len(registry))
	for k, v := range registry {
		out[k] = v
	}
	return out
}

// Run expands macro-tagged files. Empty args default to ./...; nil linked uses Registered().
func Run(args []string, linked map[string]macro.Expander) error {
	patterns := []string{"./..."}
	if len(args) > 0 {
		patterns = args
	}
	if linked == nil {
		linked = Registered()
	}
	return expander.ExpandPackages(patterns, linked)
}

// Main runs Run(os.Args[1:], nil) and exits non-zero on error.
func Main() {
	if err := Run(os.Args[1:], nil); err != nil {
		fmt.Fprintf(os.Stderr, "macroexpand: %v\n", err)
		os.Exit(1)
	}
}
