package quote

// Kind is the root @kind of a Quote template.
type Kind int

const (
	KindExpr Kind = iota
	KindExprs
	KindStmts
	KindDecls
)

func (k Kind) String() string {
	switch k {
	case KindExpr:
		return "expr"
	case KindExprs:
		return "exprs"
	case KindStmts:
		return "stmts"
	case KindDecls:
		return "decls"
	default:
		return "unknown"
	}
}

func apiName(k Kind) string {
	switch k {
	case KindExpr:
		return "Expr"
	case KindExprs:
		return "Exprs"
	case KindStmts:
		return "Stmts"
	case KindDecls:
		return "Decls"
	default:
		return "Quote"
	}
}

func parseKindName(name string) (Kind, bool) {
	switch name {
	case "expr":
		return KindExpr, true
	case "exprs":
		return KindExprs, true
	case "stmts":
		return KindStmts, true
	case "decls":
		return KindDecls, true
	default:
		return 0, false
	}
}
