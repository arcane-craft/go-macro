package pattern

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ArgKind classifies pattern arguments.
type ArgKind int

const (
	ArgCapture ArgKind = iota
	ArgDiscard
)

// Arg is one call argument pattern.
type Arg struct {
	Kind ArgKind
	Name string
}

// Callee is a callee literal pattern.
type Callee struct {
	Ident    string
	Selector string
}

// Call is a CallPattern.
type Call struct {
	Callee Callee
	Args   []Arg
}

// StmtKind classifies StmtPattern variants.
type StmtKind int

const (
	StmtAssignDefine StmtKind = iota
	StmtAssignPlain
	StmtVar
	StmtReturn
	StmtExprSemi
)

// Stmt is a StmtPattern.
type Stmt struct {
	Kind StmtKind
	LHS  string
	Call Call
}

// DeclConstraintKind classifies decl field constraints.
type DeclConstraintKind int

const (
	DeclEmbedMarker DeclConstraintKind = iota
	DeclFieldEllipsis
)

// DeclConstraint is one decl struct constraint.
type DeclConstraint struct {
	Kind   DeclConstraintKind
	Marker Callee
	Index  Arg
	Field  string
}

// Decl is a DeclPattern.
type Decl struct {
	ItemName string
	Fields   []DeclConstraint
}

// Form classifies top-level pattern forms.
type Form int

const (
	FormCall Form = iota
	FormStmt
	FormDecl
)

// Top is a parsed top-level pattern.
type Top struct {
	Form Form
	Call *Call
	Stmt *Stmt
	Decl *Decl
}

// Parse parses a normative pattern string.
func Parse(src string) (Top, error) {
	src = strings.TrimSpace(src)
	if src == "" {
		return Top{}, fmt.Errorf("macro: empty pattern")
	}
	p := &parser{src: src}
	top, err := p.parseTop()
	if err != nil {
		return Top{}, err
	}
	if p.skipSpace(); p.pos < len(p.src) {
		return Top{}, fmt.Errorf("macro: unexpected trailing input in pattern %q", src)
	}
	return top, nil
}

// CaptureNames returns capture names referenced in pattern (for Quote binding).
func CaptureNames(src string) ([]string, error) {
	top, err := Parse(src)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{})
	var names []string
	add := func(n string) {
		if n == "" || n == "_" {
			return
		}
		if _, ok := seen[n]; ok {
			return
		}
		seen[n] = struct{}{}
		names = append(names, n)
	}
	switch top.Form {
	case FormCall:
		for _, a := range top.Call.Args {
			if a.Kind == ArgCapture {
				add(a.Name)
			}
		}
	case FormStmt:
		if top.Stmt.LHS != "" {
			add(top.Stmt.LHS)
		}
		for _, a := range top.Stmt.Call.Args {
			if a.Kind == ArgCapture {
				add(a.Name)
			}
		}
	case FormDecl:
		add(top.Decl.ItemName)
		for _, c := range top.Decl.Fields {
			if c.Kind == DeclFieldEllipsis {
				add(c.Field)
			}
			if c.Index.Kind == ArgCapture {
				add(c.Index.Name)
			}
		}
	}
	return names, nil
}

type parser struct {
	src string
	pos int
}

func (p *parser) skipSpace() {
	for p.pos < len(p.src) {
		r, w := utf8.DecodeRuneInString(p.src[p.pos:])
		if !unicode.IsSpace(r) {
			return
		}
		p.pos += w
	}
}

func (p *parser) peek() byte {
	p.skipSpace()
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

func (p *parser) consume(expect string) error {
	p.skipSpace()
	if !strings.HasPrefix(p.src[p.pos:], expect) {
		return fmt.Errorf("macro: pattern expected %q at %q", expect, p.src[p.pos:])
	}
	p.pos += len(expect)
	return nil
}

func (p *parser) parseIdent() (string, error) {
	p.skipSpace()
	start := p.pos
	if p.pos >= len(p.src) {
		return "", fmt.Errorf("macro: pattern expected identifier")
	}
	r, w := utf8.DecodeRuneInString(p.src[p.pos:])
	if r != '_' && !unicode.IsLetter(r) {
		return "", fmt.Errorf("macro: pattern expected identifier at %q", p.src[p.pos:])
	}
	p.pos += w
	for p.pos < len(p.src) {
		r, w = utf8.DecodeRuneInString(p.src[p.pos:])
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			break
		}
		p.pos += w
	}
	return p.src[start:p.pos], nil
}

func (p *parser) parseCapture() (Arg, error) {
	if err := p.consume("$"); err != nil {
		return Arg{}, err
	}
	name, err := p.parseIdent()
	if err != nil {
		return Arg{}, err
	}
	if name == "_" {
		return Arg{Kind: ArgDiscard}, nil
	}
	return Arg{Kind: ArgCapture, Name: name}, nil
}

func (p *parser) parseEllipsisCapture() (string, error) {
	if err := p.consume("$"); err != nil {
		return "", err
	}
	name, err := p.parseIdent()
	if err != nil {
		return "", err
	}
	p.skipSpace()
	if err := p.consume("..."); err != nil {
		return "", fmt.Errorf("macro: pattern expected ellipsis after $%s", name)
	}
	return name, nil
}

func (p *parser) parseCallee() (Callee, error) {
	left, err := p.parseIdent()
	if err != nil {
		return Callee{}, err
	}
	p.skipSpace()
	if p.peek() == '.' {
		if err := p.consume("."); err != nil {
			return Callee{}, err
		}
		right, err := p.parseIdent()
		if err != nil {
			return Callee{}, err
		}
		return Callee{Ident: left, Selector: right}, nil
	}
	return Callee{Ident: left}, nil
}

func (p *parser) parseCall() (Call, error) {
	callee, err := p.parseCallee()
	if err != nil {
		return Call{}, err
	}
	if err := p.consume("("); err != nil {
		return Call{}, err
	}
	var args []Arg
	p.skipSpace()
	if p.peek() != ')' {
		for {
			arg, err := p.parseCapture()
			if err != nil {
				return Call{}, err
			}
			args = append(args, arg)
			p.skipSpace()
			if p.peek() != ',' {
				break
			}
			if err := p.consume(","); err != nil {
				return Call{}, err
			}
		}
	}
	if err := p.consume(")"); err != nil {
		return Call{}, err
	}
	return Call{Callee: callee, Args: args}, nil
}

func (p *parser) parseTop() (Top, error) {
	if strings.HasPrefix(strings.TrimSpace(p.src[p.pos:]), "type ") {
		return p.parseDecl()
	}
	if strings.HasPrefix(strings.TrimSpace(p.src[p.pos:]), "return ") {
		return p.parseReturnStmt()
	}
	if strings.HasPrefix(strings.TrimSpace(p.src[p.pos:]), "var ") {
		return p.parseVarStmt()
	}
	if p.peek() == '$' {
		save := p.pos
		name, err := p.parseEllipsisCapture()
		if err != nil {
			p.pos = save
		} else {
			p.skipSpace()
			if strings.HasPrefix(p.src[p.pos:], ":=") {
				if err := p.consume(":="); err != nil {
					return Top{}, err
				}
				call, err := p.parseCall()
				if err != nil {
					return Top{}, err
				}
				return Top{Form: FormStmt, Stmt: &Stmt{Kind: StmtAssignDefine, LHS: name, Call: call}}, nil
			}
			if strings.HasPrefix(p.src[p.pos:], "=") && !strings.HasPrefix(p.src[p.pos:], "==") {
				if err := p.consume("="); err != nil {
					return Top{}, err
				}
				call, err := p.parseCall()
				if err != nil {
					return Top{}, err
				}
				return Top{Form: FormStmt, Stmt: &Stmt{Kind: StmtAssignPlain, LHS: name, Call: call}}, nil
			}
		}
		if save != p.pos {
			p.pos = save
		}
	}
	call, err := p.parseCall()
	if err != nil {
		return Top{}, err
	}
	p.skipSpace()
	if p.peek() == ';' {
		if err := p.consume(";"); err != nil {
			return Top{}, err
		}
		return Top{Form: FormStmt, Stmt: &Stmt{Kind: StmtExprSemi, Call: call}}, nil
	}
	return Top{Form: FormCall, Call: &call}, nil
}

func (p *parser) parseReturnStmt() (Top, error) {
	if err := p.consume("return"); err != nil {
		return Top{}, err
	}
	vals, err := p.parseEllipsisCapture()
	if err != nil {
		return Top{}, err
	}
	if err := p.consume(","); err != nil {
		return Top{}, err
	}
	call, err := p.parseCall()
	if err != nil {
		return Top{}, err
	}
	return Top{Form: FormStmt, Stmt: &Stmt{Kind: StmtReturn, LHS: vals, Call: call}}, nil
}

func (p *parser) parseVarStmt() (Top, error) {
	if err := p.consume("var"); err != nil {
		return Top{}, err
	}
	lhs, err := p.parseEllipsisCapture()
	if err != nil {
		return Top{}, err
	}
	if err := p.consume("="); err != nil {
		return Top{}, err
	}
	call, err := p.parseCall()
	if err != nil {
		return Top{}, err
	}
	return Top{Form: FormStmt, Stmt: &Stmt{Kind: StmtVar, LHS: lhs, Call: call}}, nil
}

func (p *parser) parseDecl() (Top, error) {
	if err := p.consume("type"); err != nil {
		return Top{}, err
	}
	if err := p.consume("$"); err != nil {
		return Top{}, err
	}
	item, err := p.parseIdent()
	if err != nil {
		return Top{}, err
	}
	if err := p.consume("struct"); err != nil {
		return Top{}, err
	}
	if err := p.consume("{"); err != nil {
		return Top{}, err
	}
	var constraints []DeclConstraint
	for p.skipSpace(); p.peek() != '}'; {
		if p.peek() == '$' {
			save := p.pos
			fieldName, err := p.parseEllipsisCapture()
			if err == nil {
				constraints = append(constraints, DeclConstraint{Kind: DeclFieldEllipsis, Field: fieldName})
				p.skipSpace()
				if p.peek() == ',' {
					_ = p.consume(",")
				}
				continue
			}
			p.pos = save
		}
		marker, err := p.parseCallee()
		if err != nil {
			return Top{}, err
		}
		p.skipSpace()
		if p.peek() == '[' {
			if err := p.consume("["); err != nil {
				return Top{}, err
			}
			idx, err := p.parseCapture()
			if err != nil {
				return Top{}, err
			}
			if err := p.consume("]"); err != nil {
				return Top{}, err
			}
			constraints = append(constraints, DeclConstraint{Kind: DeclEmbedMarker, Marker: marker, Index: idx})
		} else {
			constraints = append(constraints, DeclConstraint{Kind: DeclEmbedMarker, Marker: marker, Index: Arg{Kind: ArgDiscard}})
		}
		p.skipSpace()
		if p.peek() == ',' {
			_ = p.consume(",")
		}
	}
	if err := p.consume("}"); err != nil {
		return Top{}, err
	}
	return Top{Form: FormDecl, Decl: &Decl{ItemName: item, Fields: constraints}}, nil
}
