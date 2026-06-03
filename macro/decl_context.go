package macro

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"
	"sync/atomic"
)

type implDeclContext struct {
	fset        *token.FileSet
	file        *ast.File
	info        *types.Info
	pkg         *types.Package
	site        DeclSite
	syntaxID    string
	tempCounter *atomic.Uint64
	macroPos    token.Pos
}

// NewDeclContext builds a DeclContext for expanders.
func NewDeclContext(
	fset *token.FileSet,
	file *ast.File,
	info *types.Info,
	pkg *types.Package,
	site DeclSite,
	syntaxID string,
) DeclContext {
	pos := site.EmbedField.Pos()
	if site.EmbedField.Type != nil {
		pos = site.EmbedField.Type.Pos()
	}
	return &implDeclContext{
		fset:        fset,
		file:        file,
		info:        info,
		pkg:         pkg,
		site:        site,
		syntaxID:    syntaxID,
		tempCounter: &atomic.Uint64{},
		macroPos:    pos,
	}
}

func (c *implDeclContext) FileSet() *token.FileSet       { return c.fset }
func (c *implDeclContext) File() *ast.File                 { return c.file }
func (c *implDeclContext) Types() *types.Info              { return c.info }
func (c *implDeclContext) Package() *types.Package         { return c.pkg }
func (c *implDeclContext) Site() DeclSite                  { return c.site }
func (c *implDeclContext) SyntaxID() string                { return c.syntaxID }
func (c *implDeclContext) MarkerTypeName() string          { return c.site.MarkerTypeName }
func (c *implDeclContext) MacroPos() token.Pos             { return c.macroPos }

func (c *implDeclContext) TargetMethods() []*ast.FuncDecl {
	targetName := c.site.Target.Name.Name
	var out []*ast.FuncDecl
	for _, decl := range c.file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			continue
		}
		recv := fn.Recv.List[0].Type
		if recvNameForType(recv) == targetName {
			out = append(out, fn)
		}
	}
	return out
}

func (c *implDeclContext) TempIdent(prefix string) *ast.Ident {
	n := c.tempCounter.Add(1)
	name := fmt.Sprintf("%s%d", prefix, n)
	return ast.NewIdent(name)
}

func recvNameForType(t ast.Expr) string {
	switch r := t.(type) {
	case *ast.Ident:
		return r.Name
	case *ast.StarExpr:
		if id, ok := r.X.(*ast.Ident); ok {
			return id.Name
		}
	case *ast.IndexExpr:
		return recvNameForType(r.X)
	case *ast.IndexListExpr:
		return recvNameForType(r.X)
	}
	return ""
}

// ParseMacroTag parses `macro:"k=v"` or `macro:"k=v,k2=v2"` from a struct field tag.
func ParseMacroTag(tag *ast.BasicLit) MacroTag {
	if tag == nil {
		return nil
	}
	s := strings.Trim(tag.Value, "`")
	parts := strings.Split(s, " ")
	for _, p := range parts {
		if !strings.HasPrefix(p, "macro:") {
			continue
		}
		inner := strings.TrimPrefix(p, "macro:")
		inner = strings.Trim(inner, `"`)
		return parseMacroTagKV(inner)
	}
	return nil
}

func parseMacroTagKV(s string) MacroTag {
	out := make(MacroTag)
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			out[part] = ""
		} else {
			out[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
