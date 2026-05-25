package macro

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

const macroCommentPrefix = "macro:"

// Registry maps stub names to syntax IDs and expanders.
type Registry struct {
	stubToSyntax   map[string]string
	syntaxToExpand map[string]Expander
	providerStubs  map[string]map[string]struct{}
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		stubToSyntax:   make(map[string]string),
		syntaxToExpand: make(map[string]Expander),
		providerStubs:  make(map[string]map[string]struct{}),
	}
}

// RegisterProvider scans provider AST files, registers panic stubs, and binds syntax-id to expander.
func (r *Registry) RegisterProvider(importPath string, files []*ast.File, syntaxID string, expand Expander) error {
	if expand == nil {
		return fmt.Errorf("macro: expander for %q is nil", syntaxID)
	}
	foundDirective := false
	stubs := make(map[string]struct{})
	for _, f := range files {
		sid, ok := fileMacroSyntaxID(f)
		if ok {
			if sid != syntaxID {
				return fmt.Errorf("macro: conflicting syntax-id in provider %s: %q vs %q", importPath, sid, syntaxID)
			}
			foundDirective = true
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || fn.Recv != nil || fn.Name == nil {
				continue
			}
			if isPanicStub(fn) {
				stubs[fn.Name.Name] = struct{}{}
			}
		}
	}
	if !foundDirective {
		return fmt.Errorf("macro: provider %s missing //macro: directive", importPath)
	}
	r.syntaxToExpand[syntaxID] = expand
	r.providerStubs[importPath] = stubs
	for stub := range stubs {
		r.stubToSyntax[stub] = syntaxID
	}
	return nil
}

// RegisterStub maps a stub name to a syntax ID.
func (r *Registry) RegisterStub(stubName, syntaxID string) {
	r.stubToSyntax[stubName] = syntaxID
}

// RegisterSyntax binds a syntax ID to an expander.
func (r *Registry) RegisterSyntax(syntaxID string, expand Expander) {
	r.syntaxToExpand[syntaxID] = expand
}

// Lookup returns syntax ID and expander for a stub name.
func (r *Registry) Lookup(stubName string) (syntaxID string, expand Expander, ok bool) {
	syntaxID, ok = r.stubToSyntax[stubName]
	if !ok {
		return "", nil, false
	}
	expand, ok = r.syntaxToExpand[syntaxID]
	return syntaxID, expand, ok
}

// HasStub reports whether stubName is registered for importPath.
func (r *Registry) HasStub(importPath, stubName string) bool {
	stubs, ok := r.providerStubs[importPath]
	if !ok {
		return false
	}
	_, ok = stubs[stubName]
	return ok
}

// ProviderStubs returns stub names for a provider import path.
func (r *Registry) ProviderStubs(importPath string) map[string]struct{} {
	return r.providerStubs[importPath]
}

// ParseProviderFiles parses Go source into AST files.
func ParseProviderFiles(fset *token.FileSet, sources map[string][]byte) ([]*ast.File, error) {
	var files []*ast.File
	for name, src := range sources {
		f, err := parser.ParseFile(fset, name, src, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", name, err)
		}
		files = append(files, f)
	}
	return files, nil
}

func fileMacroSyntaxID(f *ast.File) (string, bool) {
	for _, cg := range f.Comments {
		for _, c := range cg.List {
			if sid, ok := parseMacroComment(c.Text); ok {
				return sid, true
			}
		}
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Doc == nil {
			continue
		}
		for _, c := range fn.Doc.List {
			if sid, ok := parseMacroComment(c.Text); ok {
				return sid, true
			}
		}
	}
	return "", false
}

func parseMacroComment(text string) (string, bool) {
	text = strings.TrimSpace(strings.TrimPrefix(text, "//"))
	if !strings.HasPrefix(text, macroCommentPrefix) {
		return "", false
	}
	rest := strings.TrimSpace(strings.TrimPrefix(text, macroCommentPrefix))
	parts := strings.Fields(rest)
	if len(parts) == 0 {
		return "", false
	}
	return parts[0], true
}

func isPanicStub(fn *ast.FuncDecl) bool {
	if fn.Body == nil || len(fn.Body.List) != 1 {
		return false
	}
	exprStmt, ok := fn.Body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "panic"
}
