package quote

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// scanState walks template text, skipping strings and comments.
type scanState struct {
	s   string
	i   int
	err error
}

func (st *scanState) eof() bool {
	return st.i >= len(st.s)
}

func (st *scanState) peek() byte {
	if st.eof() {
		return 0
	}
	return st.s[st.i]
}

func (st *scanState) advance(n int) {
	st.i += n
}

func (st *scanState) skipWhitespace() {
	for st.i < len(st.s) {
		r, size := utf8.DecodeRuneInString(st.s[st.i:])
		if !unicode.IsSpace(r) {
			break
		}
		st.i += size
	}
}

func (st *scanState) skipString() {
	quote := st.peek()
	st.advance(1)
	for st.i < len(st.s) {
		c := st.s[st.i]
		if c == '\\' {
			st.advance(2)
			continue
		}
		st.advance(1)
		if c == quote {
			return
		}
	}
	st.err = errf("unterminated string")
}

func (st *scanState) skipRawString() {
	st.advance(1) // `
	for st.i < len(st.s) {
		if st.s[st.i] == '`' {
			st.advance(1)
			return
		}
		st.advance(1)
	}
	st.err = errf("unterminated raw string")
}

func (st *scanState) skipLineComment() {
	st.advance(2) // //
	for st.i < len(st.s) && st.s[st.i] != '\n' {
		st.advance(1)
	}
}

func (st *scanState) skipBlockComment() {
	st.advance(2) // /*
	depth := 1
	for st.i < len(st.s) && depth > 0 {
		if st.i+1 < len(st.s) && st.s[st.i] == '/' && st.s[st.i+1] == '*' {
			depth++
			st.advance(2)
			continue
		}
		if st.i+1 < len(st.s) && st.s[st.i] == '*' && st.s[st.i+1] == '/' {
			depth--
			st.advance(2)
			continue
		}
		st.advance(1)
	}
	if depth > 0 {
		st.err = errf("unterminated block comment")
	}
}

func (st *scanState) skipLiteral() {
	switch st.peek() {
	case '"', '\'':
		st.skipString()
	case '`':
		st.skipRawString()
	case '/':
		if st.i+1 < len(st.s) {
			switch st.s[st.i+1] {
			case '/':
				st.skipLineComment()
			case '*':
				st.skipBlockComment()
			}
		}
	}
}

func readIdent(s string, i int) (string, int) {
	start := i
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == '_' || unicode.IsLetter(r) || (i > start && unicode.IsDigit(r)) {
			i += size
			continue
		}
		break
	}
	return s[start:i], i
}

func placeholderName(hole string) string {
	return "_q_" + hole
}

func isHoleOnlyBody(body string) (hole string, ok bool) {
	st := &scanState{s: body}
	st.skipWhitespace()
	if st.eof() || st.peek() != '#' {
		return "", false
	}
	st.advance(1)
	name, end := readIdent(body, st.i)
	if name == "" {
		return "", false
	}
	st.i = end
	st.skipWhitespace()
	if st.i != len(body) {
		return "", false
	}
	return name, true
}

func holesInText(text string) ([]string, error) {
	st := &scanState{s: text}
	seen := make(map[string]bool)
	var holes []string
	for st.i < len(st.s) {
		st.skipLiteral()
		if st.err != nil {
			return nil, st.err
		}
		if st.eof() {
			break
		}
		if st.peek() == '#' {
			st.advance(1)
			name, end := readIdent(text, st.i)
			if name == "" {
				return nil, errf("invalid hole after # at offset %d", st.i)
			}
			st.i = end
			if !seen[name] {
				seen[name] = true
				holes = append(holes, name)
			}
			continue
		}
		st.advance(1)
	}
	return holes, nil
}

func replaceHoles(text string) (string, error) {
	var b strings.Builder
	st := &scanState{s: text}
	for st.i < len(st.s) {
		start := st.i
		st.skipLiteral()
		if st.err != nil {
			return "", st.err
		}
		if st.i > start {
			b.WriteString(text[start:st.i])
		}
		if st.eof() {
			break
		}
		if st.peek() == '#' {
			st.advance(1)
			name, end := readIdent(text, st.i)
			if name == "" {
				return "", errf("invalid hole after # at offset %d", st.i)
			}
			b.WriteString(placeholderName(name))
			st.i = end
			continue
		}
		b.WriteByte(st.peek())
		st.advance(1)
	}
	return b.String(), nil
}
