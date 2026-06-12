package macro

import (
	"fmt"

	"github.com/arcane-craft/go-macro/macro/pattern"
)

// Clause is one syntax-rule case.
type Clause struct {
	Pattern   string
	Template  string
	Fender    func(Context, Syntax, Bindings) error
	Transform func(Context, Syntax, Bindings) (Syntax, error)
}

// SyntaxRules returns an Expander from pattern-template clauses.
func SyntaxRules(clauses ...Clause) Expander {
	validateClauses(clauses)
	return func(ctx Context, site Syntax) (Syntax, error) {
		return runSyntaxCase(ctx, site, clauses, false)
	}
}

// SyntaxCase returns an Expander with optional fender and transform.
func SyntaxCase(clauses ...Clause) Expander {
	validateClauses(clauses)
	return func(ctx Context, site Syntax) (Syntax, error) {
		return runSyntaxCase(ctx, site, clauses, true)
	}
}

func validateClauses(clauses []Clause) {
	for i, cl := range clauses {
		if _, err := pattern.Parse(cl.Pattern); err != nil {
			panic(fmt.Sprintf("macro: invalid pattern in clause %d: %v", i, err))
		}
	}
}

func runSyntaxCase(ctx Context, site Syntax, clauses []Clause, allowTransform bool) (Syntax, error) {
	var lastErr error
	for _, cl := range clauses {
		clearSiteMeta(site)
		binds, err := site.Match(cl.Pattern)
		if err != nil {
			lastErr = err
			continue
		}
		if cl.Fender != nil {
			if err := cl.Fender(ctx, site, binds); err != nil {
				clearSiteMeta(site)
				lastErr = err
				continue
			}
		}
		if cl.Transform != nil {
			if !allowTransform {
				return nil, fmt.Errorf("macro: Transform requires SyntaxCase")
			}
			return cl.Transform(ctx, site, binds)
		}
		if cl.Template == "" {
			clearSiteMeta(site)
			lastErr = fmt.Errorf("macro: clause missing Template and Transform")
			continue
		}
		out, err := Quote(cl.Template, quoteBinds(binds, cl.Pattern))
		if err != nil {
			clearSiteMeta(site)
			lastErr = err
			continue
		}
		return out, nil
	}
	if lastErr != nil {
		return nil, fmt.Errorf("macro: no matching syntax rule: %v", lastErr)
	}
	return nil, fmt.Errorf("macro: no matching syntax rule")
}

func clearSiteMeta(site Syntax) {
	if ms, ok := site.(MetaSlot); ok {
		ms.ClearExpansionMeta()
	}
}

func quoteBinds(binds Bindings, patternSrc string) map[string]Syntax {
	names, err := pattern.CaptureNames(patternSrc)
	if err != nil {
		return nil
	}
	m := make(map[string]Syntax, len(names))
	for _, name := range names {
		if elems, ok := binds.Elems(name); ok && len(elems) > 0 {
			if len(elems) == 1 {
				m[name] = elems[0]
			} else {
				m[name] = WrapSyntaxList(elems)
			}
			continue
		}
		if v, ok := binds.Get(name); ok {
			m[name] = v
		}
	}
	return m
}
