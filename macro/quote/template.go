package quote

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ParsedTemplate holds a parsed Quote template root and body.
type ParsedTemplate struct {
	Kind  Kind
	Body  string
	Holes []string
}

// resolveTemplate parses tpl for a typed API (Expr/Exprs/Stmts/Decls).
// When tpl has no @kind{ } wrapper, the entire string is treated as template body
// and kind comes from the API. Quote uses parseTemplate and always requires @kind{ }.
func resolveTemplate(tpl string, kind Kind) (*ParsedTemplate, error) {
	if !hasRootKindWrapper(tpl) {
		holes, err := holesInText(tpl)
		if err != nil {
			return nil, err
		}
		return &ParsedTemplate{Kind: kind, Body: tpl, Holes: holes}, nil
	}
	pt, err := parseTemplate(tpl)
	if err != nil {
		return nil, err
	}
	if pt.Kind != kind {
		return nil, errKindMismatch(pt.Kind, apiName(kind))
	}
	return pt, nil
}

func hasRootKindWrapper(tpl string) bool {
	st := &scanState{s: tpl}
	st.skipWhitespace()
	if st.eof() || st.peek() != '@' {
		return false
	}
	st.advance(1)

	kindStart := st.i
	for st.i < len(st.s) {
		r, size := utf8.DecodeRuneInString(st.s[st.i:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			st.i += size
			continue
		}
		break
	}
	kindName := tpl[kindStart:st.i]
	if _, ok := parseKindName(kindName); !ok {
		return false
	}
	st.skipWhitespace()
	return !st.eof() && st.peek() == '{'
}

func parseTemplate(tpl string) (*ParsedTemplate, error) {
	st := &scanState{s: tpl}
	st.skipWhitespace()
	if st.eof() {
		return nil, errf("missing root @kind{ }")
	}
	if st.peek() != '@' {
		return nil, errf("missing root @kind{ }")
	}
	st.advance(1)

	kindStart := st.i
	for st.i < len(st.s) {
		r, size := utf8.DecodeRuneInString(st.s[st.i:])
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			st.i += size
			continue
		}
		break
	}
	kindName := tpl[kindStart:st.i]
	kind, ok := parseKindName(kindName)
	if !ok {
		return nil, errf("unknown root kind %q", kindName)
	}

	st.skipWhitespace()
	if st.eof() || st.peek() != '{' {
		return nil, errf("root @%s must use braces { }", kind)
	}
	st.advance(1)

	bodyStart := st.i
	depth := 1
	for st.i < len(st.s) && depth > 0 {
		st.skipLiteral()
		if st.err != nil {
			return nil, st.err
		}
		if st.eof() {
			return nil, errf("unclosed root brace for @%s", kind)
		}
		switch st.peek() {
		case '{':
			depth++
			st.advance(1)
		case '}':
			depth--
			if depth == 0 {
				body := tpl[bodyStart:st.i]
				st.advance(1)
				rest := strings.TrimSpace(tpl[st.i:])
				if rest != "" {
					return nil, errf("unexpected content after root @%s{ }", kind)
				}
				holes, err := holesInText(body)
				if err != nil {
					return nil, err
				}
				return &ParsedTemplate{Kind: kind, Body: body, Holes: holes}, nil
			}
			st.advance(1)
		default:
			st.advance(1)
		}
	}
	return nil, errf("unclosed root brace for @%s", kind)
}
