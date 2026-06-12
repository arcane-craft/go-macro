package quote

import "fmt"

func synthesize(kind Kind, body string) (string, error) {
	replaced, err := replaceHoles(body)
	if err != nil {
		return "", err
	}
	switch kind {
	case KindExpr:
		return replaced, nil
	case KindExprs:
		return fmt.Sprintf("package _\nfunc _() { return %s }", replaced), nil
	case KindStmts:
		return fmt.Sprintf("package _\nfunc _() { %s }", replaced), nil
	case KindDecls:
		return fmt.Sprintf("package _\n%s", replaced), nil
	default:
		return "", errf("unknown kind %v", kind)
	}
}
