package macro

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

const macroCommentPrefix = "macro:"

// Registry maps provider import paths to stubs and expanders.
type Registry struct {
	stubSyntax     map[string]map[string]string // importPath -> stubName -> syntaxID
	syntaxToExpand map[string]Expander
	importExpand   map[string]Expander // importPath -> expander
	providerStubs  map[string]map[string]struct{}
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		stubSyntax:     make(map[string]map[string]string),
		syntaxToExpand: make(map[string]Expander),
		importExpand:   make(map[string]Expander),
		providerStubs:  make(map[string]map[string]struct{}),
	}
}

// ProviderInfo describes a macro provider package discovered from source.
type ProviderInfo struct {
	SyntaxID     string
	ExpanderName string
	StubNames    []string
}

// ScanProviderFiles parses provider sources for per-function //macro: directives.
func ScanProviderFiles(files []*ast.File) (ProviderInfo, error) {
	var info ProviderInfo
	expanders := make(map[string]string) // syntaxID -> func name

	for _, f := range files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name == nil {
				continue
			}
			sid, ok := funcMacroSyntaxID(fn)
			if !ok {
				continue
			}
			if isExpanderDecl(fn) {
				if prev, exists := expanders[sid]; exists && prev != fn.Name.Name {
					return ProviderInfo{}, fmt.Errorf("macro: multiple expanders for syntax-id %q", sid)
				}
				expanders[sid] = fn.Name.Name
				continue
			}
			if info.SyntaxID == "" {
				info.SyntaxID = sid
			} else if sid != info.SyntaxID {
				return ProviderInfo{}, fmt.Errorf("macro: stub %q syntax-id %q != %q",
					fn.Name.Name, sid, info.SyntaxID)
			}
			info.StubNames = append(info.StubNames, fn.Name.Name)
		}
	}

	if len(expanders) == 0 {
		return ProviderInfo{}, fmt.Errorf("macro: provider missing Expander with //macro: directive")
	}
	if len(expanders) > 1 {
		return ProviderInfo{}, fmt.Errorf("macro: provider has multiple syntax-id expanders (%d)", len(expanders))
	}
	for sid, name := range expanders {
		if info.SyntaxID != "" && sid != info.SyntaxID {
			return ProviderInfo{}, fmt.Errorf("macro: expander syntax-id %q != stub syntax-id %q", sid, info.SyntaxID)
		}
		info.SyntaxID = sid
		info.ExpanderName = name
	}
	if len(info.StubNames) == 0 {
		return ProviderInfo{}, fmt.Errorf("macro: provider has no stub functions with //macro: directive")
	}
	return info, nil
}

// ProviderSyntaxID returns the syntax-id from the sole Expander's //macro: directive.
func ProviderSyntaxID(files []*ast.File) (string, error) {
	info, err := ScanProviderFiles(files)
	if err != nil {
		return "", err
	}
	return info.SyntaxID, nil
}

// RegisterProvider registers stubs and binds the linked expander for importPath.
func (r *Registry) RegisterProvider(importPath string, files []*ast.File, expand Expander) error {
	if expand == nil {
		return fmt.Errorf("macro: expander for %q is nil", importPath)
	}
	info, err := ScanProviderFiles(files)
	if err != nil {
		return fmt.Errorf("macro: provider %s: %w", importPath, err)
	}

	stubs := make(map[string]struct{})
	stubSyntax := make(map[string]string)
	for _, f := range files {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name == nil || isExpanderDecl(fn) {
				continue
			}
			sid, ok := funcMacroSyntaxID(fn)
			if !ok {
				continue
			}
			if sid != info.SyntaxID {
				return fmt.Errorf("macro: conflicting syntax-id in provider %s", importPath)
			}
			stubs[fn.Name.Name] = struct{}{}
			stubSyntax[fn.Name.Name] = sid
		}
	}

	r.importExpand[importPath] = expand
	r.syntaxToExpand[info.SyntaxID] = expand
	r.providerStubs[importPath] = stubs
	if r.stubSyntax[importPath] == nil {
		r.stubSyntax[importPath] = make(map[string]string)
	}
	for k, v := range stubSyntax {
		r.stubSyntax[importPath][k] = v
	}
	return nil
}

// RegisterStub maps a stub name to a syntax ID for importPath (tests).
func (r *Registry) RegisterStub(importPath, stubName, syntaxID string) {
	if r.stubSyntax[importPath] == nil {
		r.stubSyntax[importPath] = make(map[string]string)
	}
	r.stubSyntax[importPath][stubName] = syntaxID
	if r.providerStubs[importPath] == nil {
		r.providerStubs[importPath] = make(map[string]struct{})
	}
	r.providerStubs[importPath][stubName] = struct{}{}
}

// RegisterSyntax binds a syntax ID to an expander (tests).
func (r *Registry) RegisterSyntax(syntaxID string, expand Expander) {
	r.syntaxToExpand[syntaxID] = expand
}

// RegisterImportExpander binds importPath to expander (tests).
func (r *Registry) RegisterImportExpander(importPath string, expand Expander) {
	r.importExpand[importPath] = expand
}

// Lookup returns syntax ID and expander for a stub in importPath.
func (r *Registry) Lookup(importPath, stubName string) (syntaxID string, expand Expander, ok bool) {
	stubs, ok := r.providerStubs[importPath]
	if !ok {
		return "", nil, false
	}
	if _, ok = stubs[stubName]; !ok {
		return "", nil, false
	}
	syntaxID, ok = r.stubSyntax[importPath][stubName]
	if !ok {
		return "", nil, false
	}
	expand, ok = r.importExpand[importPath]
	if !ok {
		expand, ok = r.syntaxToExpand[syntaxID]
	}
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

func funcMacroSyntaxID(fn *ast.FuncDecl) (string, bool) {
	if fn.Doc == nil {
		return "", false
	}
	for _, c := range fn.Doc.List {
		if sid, ok := parseMacroComment(c.Text); ok {
			return sid, true
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

// isExpanderDecl reports whether fn looks like func(Context, *ast.CallExpr) (ExpandResult, error).
func isExpanderDecl(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || fn.Type.Results == nil {
		return false
	}
	if len(fn.Type.Params.List) != 2 || len(fn.Type.Results.List) != 2 {
		return false
	}
	return true
}
