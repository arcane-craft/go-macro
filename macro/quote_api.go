package macro

import (
	"strings"

	"github.com/arcane-craft/go-macro/internal/quote"
)

// Quote expands a template with # holes filled from binds. Output shape is determined by To*.
func Quote(tpl string, binds map[string]Syntax) (Syntax, error) {
	args := make(map[string]any, len(binds))
	for k, v := range binds {
		if nodes, ok := QuoteElems(v); ok {
			args[k] = nodes
			continue
		}
		args[k] = v.Underlying()
	}
	tpl = strings.TrimSpace(tpl)
	if tpl == "" {
		return nil, quoteErr("empty template")
	}
	kind := inferQuoteKind(tpl)
	switch kind {
	case quote.KindExpr:
		e, err := quote.Expr(tpl, args)
		if err != nil {
			return nil, err
		}
		return WrapExpr(e), nil
	case quote.KindExprs:
		es, err := quote.Exprs(tpl, args)
		if err != nil {
			return nil, err
		}
		return WrapExprs(es), nil
	case quote.KindStmts:
		ss, err := quote.Stmts(tpl, args)
		if err != nil {
			return nil, err
		}
		return WrapStmts(ss), nil
	case quote.KindDecls:
		ds, err := quote.Decls(tpl, args)
		if err != nil {
			return nil, err
		}
		return WrapDecls(ds), nil
	default:
		return nil, quoteErr("cannot infer template kind")
	}
}

func inferQuoteKind(tpl string) quote.Kind {
	trim := strings.TrimSpace(tpl)
	switch {
	case strings.HasPrefix(trim, "type "), strings.HasPrefix(trim, "func "), strings.HasPrefix(trim, "var "), strings.HasPrefix(trim, "const "), strings.HasPrefix(trim, "import "):
		return quote.KindDecls
	case strings.HasPrefix(trim, "return "), strings.HasPrefix(trim, "if "), strings.HasPrefix(trim, "for "),
		strings.HasPrefix(trim, "switch "), strings.Contains(trim, ":="), strings.HasSuffix(trim, ";"):
		return quote.KindStmts
	default:
		return quote.KindExpr
	}
}

func quoteErr(msg string) error {
	return &quoteError{msg: msg}
}

type quoteError struct{ msg string }

func (e *quoteError) Error() string { return "macro: " + e.msg }
