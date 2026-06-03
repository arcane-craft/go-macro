package macro

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

const macroCommentPrefix = "macro:"

// Registry maps provider import paths to stubs, markers, and expanders.
type Registry struct {
	stubSyntax      map[string]map[string]string // importPath -> stubName -> syntaxID
	markerSyntax    map[string]map[string]string // importPath -> markerBaseName -> syntaxID
	syntaxToCall    map[string]CallExpander
	syntaxToDecl    map[string]DeclExpander
	importCall      map[string]CallExpander
	providerStubs   map[string]map[string]struct{}
	providerMarkers map[string]map[string]struct{}
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		stubSyntax:      make(map[string]map[string]string),
		markerSyntax:    make(map[string]map[string]string),
		syntaxToCall:    make(map[string]CallExpander),
		syntaxToDecl:    make(map[string]DeclExpander),
		importCall:      make(map[string]CallExpander),
		providerStubs:   make(map[string]map[string]struct{}),
		providerMarkers: make(map[string]map[string]struct{}),
	}
}

// ProviderInfo describes one syntax-id entry in a provider package.
type ProviderInfo struct {
	SyntaxID       string
	CallExpander   string
	DeclExpander   string
	StubNames      []string
	MarkerTypeNames []string
}

// ProviderScan aggregates all syntax-ids discovered in provider sources.
type ProviderScan struct {
	Entries []ProviderInfo
}

// ScanProviderFiles parses provider sources for //macro: on stubs, markers, and expanders.
func ScanProviderFiles(files []*ast.File) (ProviderScan, error) {
	type entry struct {
		stubs        []string
		markers      []string
		callExpander string
		declExpander string
	}
	bySyntax := make(map[string]*entry)

	for _, f := range files {
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Recv != nil || d.Name == nil {
					continue
				}
				sid, ok := funcMacroSyntaxID(d)
				if !ok {
					continue
				}
				e := bySyntax[sid]
				if e == nil {
					e = &entry{}
					bySyntax[sid] = e
				}
				if isCallExpanderDecl(d) {
					if e.callExpander != "" && e.callExpander != d.Name.Name {
						return ProviderScan{}, fmt.Errorf("macro: multiple call expanders for syntax-id %q", sid)
					}
					e.callExpander = d.Name.Name
					continue
				}
				if isDeclExpanderDecl(d) {
					if e.declExpander != "" && e.declExpander != d.Name.Name {
						return ProviderScan{}, fmt.Errorf("macro: multiple decl expanders for syntax-id %q", sid)
					}
					e.declExpander = d.Name.Name
					continue
				}
				e.stubs = append(e.stubs, d.Name.Name)
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.Name == nil {
						continue
					}
					sid, ok := typeMacroSyntaxID(ts)
					if !ok {
						continue
					}
					e := bySyntax[sid]
					if e == nil {
						e = &entry{}
						bySyntax[sid] = e
					}
					e.markers = append(e.markers, ts.Name.Name)
				}
			}
		}
	}

	if len(bySyntax) == 0 {
		return ProviderScan{}, fmt.Errorf("macro: provider has no //macro: directives")
	}

	var scan ProviderScan
	for sid, e := range bySyntax {
		if len(e.stubs) == 0 && len(e.markers) == 0 {
			return ProviderScan{}, fmt.Errorf("macro: syntax-id %q has no stubs or markers", sid)
		}
		scan.Entries = append(scan.Entries, ProviderInfo{
			SyntaxID:        sid,
			CallExpander:    e.callExpander,
			DeclExpander:    e.declExpander,
			StubNames:       e.stubs,
			MarkerTypeNames: e.markers,
		})
	}
	return scan, nil
}

// ProviderSyntaxID returns the first syntax-id with a call expander (legacy helper).
func ProviderSyntaxID(files []*ast.File) (string, error) {
	scan, err := ScanProviderFiles(files)
	if err != nil {
		return "", err
	}
	for _, e := range scan.Entries {
		if e.CallExpander != "" {
			return e.SyntaxID, nil
		}
	}
	if len(scan.Entries) > 0 {
		return scan.Entries[0].SyntaxID, nil
	}
	return "", fmt.Errorf("macro: no syntax-id found")
}

// RegisterProviderSources registers stubs and markers from provider files.
func (r *Registry) RegisterProviderSources(importPath string, files []*ast.File) error {
	scan, err := ScanProviderFiles(files)
	if err != nil {
		return fmt.Errorf("macro: provider %s: %w", importPath, err)
	}
	stubs := make(map[string]struct{})
	markers := make(map[string]struct{})
	if r.stubSyntax[importPath] == nil {
		r.stubSyntax[importPath] = make(map[string]string)
	}
	if r.markerSyntax[importPath] == nil {
		r.markerSyntax[importPath] = make(map[string]string)
	}
	for _, e := range scan.Entries {
		for _, name := range e.StubNames {
			stubs[name] = struct{}{}
			r.stubSyntax[importPath][name] = e.SyntaxID
		}
		for _, name := range e.MarkerTypeNames {
			markers[name] = struct{}{}
			r.markerSyntax[importPath][name] = e.SyntaxID
		}
	}
	r.providerStubs[importPath] = stubs
	r.providerMarkers[importPath] = markers
	return nil
}

// RegisterProvider registers stubs and binds the linked call expander for importPath (legacy).
func (r *Registry) RegisterProvider(importPath string, files []*ast.File, expand CallExpander) error {
	if expand == nil {
		return fmt.Errorf("macro: expander for %q is nil", importPath)
	}
	if err := r.RegisterProviderSources(importPath, files); err != nil {
		return err
	}
	r.importCall[importPath] = expand
	scan, _ := ScanProviderFiles(files)
	for _, e := range scan.Entries {
		if len(e.StubNames) > 0 {
			r.syntaxToCall[e.SyntaxID] = expand
		}
	}
	return nil
}

// RegisterCallSyntax binds a syntax-id to a call expander.
func (r *Registry) RegisterCallSyntax(syntaxID string, expand CallExpander) {
	r.syntaxToCall[syntaxID] = expand
}

// RegisterDeclSyntax binds a syntax-id to a decl expander.
func (r *Registry) RegisterDeclSyntax(syntaxID string, expand DeclExpander) {
	r.syntaxToDecl[syntaxID] = expand
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

// RegisterMarker maps a marker type name to syntax ID for importPath (tests).
func (r *Registry) RegisterMarker(importPath, markerName, syntaxID string) {
	if r.markerSyntax[importPath] == nil {
		r.markerSyntax[importPath] = make(map[string]string)
	}
	r.markerSyntax[importPath][markerName] = syntaxID
	if r.providerMarkers[importPath] == nil {
		r.providerMarkers[importPath] = make(map[string]struct{})
	}
	r.providerMarkers[importPath][markerName] = struct{}{}
}

// RegisterSyntax binds a syntax ID to a call expander (tests).
func (r *Registry) RegisterSyntax(syntaxID string, expand CallExpander) {
	r.syntaxToCall[syntaxID] = expand
}

// RegisterImportCallExpander binds importPath to expander (tests).
func (r *Registry) RegisterImportCallExpander(importPath string, expand CallExpander) {
	r.importCall[importPath] = expand
}

// Lookup returns syntax ID and expander for a stub in importPath.
func (r *Registry) Lookup(importPath, stubName string) (syntaxID string, expand CallExpander, ok bool) {
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
	expand, ok = r.importCall[importPath]
	if !ok {
		expand, ok = r.syntaxToCall[syntaxID]
	}
	return syntaxID, expand, ok
}

// LookupDeclMarker returns syntax ID and decl expander for a marker base name in importPath.
func (r *Registry) LookupDeclMarker(importPath, markerBaseName string) (syntaxID string, expand DeclExpander, ok bool) {
	markers, ok := r.providerMarkers[importPath]
	if !ok {
		return "", nil, false
	}
	if _, ok = markers[markerBaseName]; !ok {
		return "", nil, false
	}
	syntaxID, ok = r.markerSyntax[importPath][markerBaseName]
	if !ok {
		return "", nil, false
	}
	expand, ok = r.syntaxToDecl[syntaxID]
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

// HasMarker reports whether markerBaseName is registered for importPath.
func (r *Registry) HasMarker(importPath, markerBaseName string) bool {
	markers, ok := r.providerMarkers[importPath]
	if !ok {
		return false
	}
	_, ok = markers[markerBaseName]
	return ok
}

// ProviderStubs returns stub names for a provider import path.
func (r *Registry) ProviderStubs(importPath string) map[string]struct{} {
	return r.providerStubs[importPath]
}

// ProviderMarkers returns marker type base names for a provider import path.
func (r *Registry) ProviderMarkers(importPath string) map[string]struct{} {
	return r.providerMarkers[importPath]
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

func typeMacroSyntaxID(ts *ast.TypeSpec) (string, bool) {
	if ts.Doc == nil {
		return "", false
	}
	for _, c := range ts.Doc.List {
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

func isCallExpanderDecl(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || fn.Type.Results == nil {
		return false
	}
	if len(fn.Type.Params.List) != 2 || len(fn.Type.Results.List) != 2 {
		return false
	}
	return expanderSecondParam(fn) == "CallExpr"
}

func isDeclExpanderDecl(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || fn.Type.Results == nil {
		return false
	}
	if len(fn.Type.Params.List) != 2 || len(fn.Type.Results.List) != 2 {
		return false
	}
	return expanderSecondParam(fn) == "DeclSite"
}

func expanderSecondParam(fn *ast.FuncDecl) string {
	if len(fn.Type.Params.List) < 2 {
		return ""
	}
	return typeExprBaseName(fn.Type.Params.List[1].Type)
}

func typeExprBaseName(t ast.Expr) string {
	switch e := t.(type) {
	case *ast.StarExpr:
		return typeExprBaseName(e.X)
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	}
	return ""
}
