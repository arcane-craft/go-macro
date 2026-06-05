package quote

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

type parsedAST struct {
	fset  *token.FileSet
	file  *ast.File
	kind  Kind
	expr  ast.Expr
	exprs []ast.Expr
	stmts []ast.Stmt
	decls []ast.Decl
}

func parseSynthesized(kind Kind, src string) (*parsedAST, error) {
	fset := token.NewFileSet()
	switch kind {
	case KindExpr:
		expr, err := parser.ParseExpr(src)
		if err != nil {
			return nil, errf("parse expr: %v", err)
		}
		return &parsedAST{fset: fset, kind: kind, expr: expr}, nil
	default:
		file, err := parser.ParseFile(fset, "quote.go", src, parser.ParseComments)
		if err != nil {
			return nil, errf("parse: %v", err)
		}
		out := &parsedAST{fset: fset, file: file, kind: kind}
		switch kind {
		case KindExprs:
			if len(file.Decls) != 1 {
				return nil, errf("expected one func decl for @exprs")
			}
			fn, ok := file.Decls[0].(*ast.FuncDecl)
			if !ok || fn.Body == nil || len(fn.Body.List) != 1 {
				return nil, errf("expected return stmt for @exprs")
			}
			ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
			if !ok || len(ret.Results) == 0 {
				return nil, errf("expected non-empty return for @exprs")
			}
			out.exprs = ret.Results
		case KindStmts:
			if len(file.Decls) != 1 {
				return nil, errf("expected one func decl for @stmts")
			}
			fn, ok := file.Decls[0].(*ast.FuncDecl)
			if !ok || fn.Body == nil || len(fn.Body.List) == 0 {
				return nil, errf("expected non-empty stmt list for @stmts")
			}
			out.stmts = fn.Body.List
		case KindDecls:
			if len(file.Decls) == 0 {
				return nil, errf("expected non-empty decl list for @decls")
			}
			out.decls = file.Decls
		}
		return out, nil
	}
}

func formatParsedFile(pt *parsedAST) (string, error) {
	if pt.file == nil {
		return "", errf("no file to format")
	}
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, pt.fset, pt.file); err != nil {
		return "", err
	}
	return buf.String(), nil
}
