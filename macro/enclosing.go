package macro

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// FileCarrier is implemented by engine-constructed site Syntax values.
type FileCarrier interface {
	Syntax
	ExpansionFile() *ast.File
}

// EnclosingSignature returns the types.Signature of the function enclosing site.
func EnclosingSignature(ctx Context, site Syntax) (*types.Signature, error) {
	fn, err := enclosingFuncObject(ctx, site)
	if err != nil {
		return nil, err
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return nil, fmt.Errorf("macro: enclosing func is not a Signature")
	}
	return sig, nil
}

// EnclosingResults returns the result tuple of the enclosing function.
func EnclosingResults(ctx Context, site Syntax) (*types.Tuple, error) {
	sig, err := EnclosingSignature(ctx, site)
	if err != nil {
		return nil, err
	}
	return sig.Results(), nil
}

// ZeroSyntax returns a Syntax literal zero value for typ.
func ZeroSyntax(ctx Context, typ types.Type) (Syntax, error) {
	if typ == nil {
		return nil, fmt.Errorf("macro: nil type")
	}
	switch u := typ.Underlying().(type) {
	case *types.Basic:
		switch u.Kind() {
		case types.Bool:
			return WrapExpr(&ast.Ident{Name: "false"}), nil
		case types.String:
			return WrapExpr(&ast.BasicLit{Kind: token.STRING, Value: `""`}), nil
		case types.Int, types.Int8, types.Int16, types.Int32, types.Int64,
			types.Uint, types.Uint8, types.Uint16, types.Uint32, types.Uint64,
			types.Uintptr, types.Float32, types.Float64, types.Complex64, types.Complex128:
			return WrapExpr(&ast.BasicLit{Kind: token.INT, Value: "0"}), nil
		case types.UnsafePointer:
			return WrapExpr(ast.NewIdent("nil")), nil
		default:
			return WrapExpr(ast.NewIdent("nil")), nil
		}
	case *types.Pointer, *types.Slice, *types.Map, *types.Chan, *types.Interface, *types.Signature:
		return WrapExpr(ast.NewIdent("nil")), nil
	case *types.Struct, *types.Array:
		var expr ast.Expr
		if ctx != nil && ctx.FileSet() != nil {
			expr = ast.NewIdent("0") // fallback; complex zeros need composite lit
		}
		return WrapExpr(expr), nil
	default:
		return WrapExpr(ast.NewIdent("nil")), nil
	}
}

func enclosingFuncObject(ctx Context, site Syntax) (*types.Func, error) {
	if ctx == nil || ctx.Types() == nil {
		return nil, fmt.Errorf("macro: Context.Types required")
	}
	fc, ok := site.(FileCarrier)
	if !ok {
		return nil, fmt.Errorf("macro: site does not carry file context")
	}
	file := fc.ExpansionFile()
	pos := site.MacroPos()
	if !pos.IsValid() {
		return nil, fmt.Errorf("macro: invalid MacroPos")
	}
	fnNode := enclosingFuncNode(file, pos)
	if fnNode == nil {
		return nil, fmt.Errorf("macro: no enclosing function")
	}
	var name *ast.Ident
	switch f := fnNode.(type) {
	case *ast.FuncDecl:
		name = f.Name
	case *ast.FuncLit:
		// FuncLit has no name; use Types scope on body
		if tv, ok := ctx.Types().Types[f]; ok && tv.Type != nil {
			if sig, ok := tv.Type.(*types.Signature); ok {
				return types.NewFunc(token.NoPos, nil, "_", sig), nil
			}
		}
		return nil, fmt.Errorf("macro: cannot resolve FuncLit signature")
	default:
		return nil, fmt.Errorf("macro: unexpected enclosing node %T", fnNode)
	}
	obj := ctx.Types().ObjectOf(name)
	if obj == nil {
		return nil, fmt.Errorf("macro: cannot resolve enclosing func object")
	}
	fn, ok := obj.(*types.Func)
	if !ok {
		return nil, fmt.Errorf("macro: enclosing object is not *types.Func")
	}
	return fn, nil
}

func enclosingFuncNode(file *ast.File, pos token.Pos) ast.Node {
	var enc ast.Node
	best := -1
	ast.Inspect(file, func(n ast.Node) bool {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			if fn.Body != nil && pos >= fn.Pos() && pos <= fn.End() {
				span := int(fn.End() - fn.Pos())
				if span > best {
					best = span
					enc = fn
				}
			}
		case *ast.FuncLit:
			if fn.Body != nil && pos >= fn.Pos() && pos <= fn.End() {
				span := int(fn.End() - fn.Pos())
				if span > best {
					best = span
					enc = fn
				}
			}
		}
		return true
	})
	return enc
}
