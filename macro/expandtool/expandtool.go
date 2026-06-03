package expandtool

import (
	"os"
	"sync"

	"github.com/arcane-craft/go-macro/internal/expander"
	"github.com/arcane-craft/go-macro/macro"
)

var (
	callMu       sync.RWMutex
	callRegistry = make(map[string]macro.CallExpander)
	declMu       sync.RWMutex
	declRegistry = make(map[string]macro.DeclExpander)
)

// RegisterCall records a CallExpander for syntaxID.
func RegisterCall(syntaxID string, expand macro.CallExpander) {
	callMu.Lock()
	defer callMu.Unlock()
	callRegistry[syntaxID] = expand
}

// RegisterDecl records a DeclExpander for syntaxID.
func RegisterDecl(syntaxID string, expand macro.DeclExpander) {
	declMu.Lock()
	defer declMu.Unlock()
	declRegistry[syntaxID] = expand
}

// RegisteredCall returns a copy of the Call registration table.
func RegisteredCall() map[string]macro.CallExpander {
	callMu.RLock()
	defer callMu.RUnlock()
	out := make(map[string]macro.CallExpander, len(callRegistry))
	for k, v := range callRegistry {
		out[k] = v
	}
	return out
}

// RegisteredDecl returns a copy of the Decl registration table.
func RegisteredDecl() map[string]macro.DeclExpander {
	declMu.RLock()
	defer declMu.RUnlock()
	out := make(map[string]macro.DeclExpander, len(declRegistry))
	for k, v := range declRegistry {
		out[k] = v
	}
	return out
}

// Run expands macro-tagged files. Empty args default to ./...; nil linked uses Registered* tables.
func Run(args []string, linked *macro.LinkedExpanders) error {
	patterns := []string{"./..."}
	if len(args) > 0 {
		patterns = args
	}
	if linked == nil {
		linked = &macro.LinkedExpanders{
			Call: RegisteredCall(),
			Decl: RegisteredDecl(),
		}
	}
	return expander.ExpandPackages(patterns, linked)
}

// Main runs Run(os.Args[1:], nil) and exits non-zero on error.
func Main() {
	if err := Run(os.Args[1:], nil); err != nil {
		ExitWithError(err)
	}
}
