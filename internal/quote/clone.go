package quote

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

func cloneExpr(e ast.Expr) ast.Expr {
	if e == nil {
		return nil
	}
	var buf bytes.Buffer
	fset := token.NewFileSet()
	_ = printer.Fprint(&buf, fset, e)
	out, err := parser.ParseExpr(buf.String())
	if err != nil {
		panic("quote: clone expr: " + err.Error())
	}
	return out
}

func cloneExprs(exprs []ast.Expr) []ast.Expr {
	if exprs == nil {
		return nil
	}
	out := make([]ast.Expr, len(exprs))
	for i, e := range exprs {
		out[i] = cloneExpr(e)
	}
	return out
}

func cloneStmt(s ast.Stmt) ast.Stmt {
	if s == nil {
		return nil
	}
	src := "package _\nfunc _() { " + stmtString(s) + " }"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "quote.go", src, parser.ParseComments)
	if err != nil {
		panic("quote: clone stmt: " + err.Error())
	}
	fn := file.Decls[0].(*ast.FuncDecl)
	return fn.Body.List[0]
}

func cloneStmts(stmts []ast.Stmt) []ast.Stmt {
	if stmts == nil {
		return nil
	}
	var body bytes.Buffer
	for i, s := range stmts {
		if i > 0 {
			body.WriteByte('\n')
		}
		body.WriteString(stmtString(s))
	}
	src := "package _\nfunc _() {\n" + body.String() + "\n}"
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "quote.go", src, parser.ParseComments)
	if err != nil {
		panic("quote: clone stmts: " + err.Error())
	}
	fn := file.Decls[0].(*ast.FuncDecl)
	return fn.Body.List
}

func cloneDecl(d ast.Decl) ast.Decl {
	if d == nil {
		return nil
	}
	src := "package _\n" + declString(d)
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "quote.go", src, parser.ParseComments)
	if err != nil {
		panic("quote: clone decl: " + err.Error())
	}
	return file.Decls[0]
}

func cloneDecls(decls []ast.Decl) []ast.Decl {
	if decls == nil {
		return nil
	}
	var body bytes.Buffer
	for i, d := range decls {
		if i > 0 {
			body.WriteByte('\n')
		}
		body.WriteString(declString(d))
	}
	src := "package _\n" + body.String()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "quote.go", src, parser.ParseComments)
	if err != nil {
		panic("quote: clone decls: " + err.Error())
	}
	return file.Decls
}

func stmtString(s ast.Stmt) string {
	fset := token.NewFileSet()
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, s)
	return buf.String()
}

func declString(d ast.Decl) string {
	fset := token.NewFileSet()
	var buf bytes.Buffer
	_ = printer.Fprint(&buf, fset, d)
	return buf.String()
}

func cloneNode(n ast.Node) (ast.Node, error) {
	switch x := n.(type) {
	case ast.Expr:
		return cloneExpr(x), nil
	case ast.Stmt:
		return cloneStmt(x), nil
	case ast.Decl:
		return cloneDecl(x), nil
	default:
		return nil, errBadBinding("", "unsupported ast.Node kind")
	}
}
