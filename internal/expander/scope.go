package expander

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
)

// FuncAtPos returns the innermost *types.Func enclosing pos in file.
// Does not expose *ast.FuncDecl to callers.
func FuncAtPos(file *ast.File, info *types.Info, pos token.Pos) (*types.Func, error) {
	if file == nil || info == nil {
		return nil, fmt.Errorf("expander: file and types.Info required")
	}
	if !pos.IsValid() {
		return nil, fmt.Errorf("expander: invalid position")
	}
	fnNode := enclosingFuncNodeAt(file, pos)
	if fnNode == nil {
		return nil, fmt.Errorf("expander: no enclosing function")
	}
	switch f := fnNode.(type) {
	case *ast.FuncDecl:
		obj := info.ObjectOf(f.Name)
		if obj == nil {
			return nil, fmt.Errorf("expander: cannot resolve func object")
		}
		fn, ok := obj.(*types.Func)
		if !ok {
			return nil, fmt.Errorf("expander: enclosing object is not *types.Func")
		}
		return fn, nil
	case *ast.FuncLit:
		if tv, ok := info.Types[f]; ok && tv.Type != nil {
			if sig, ok := tv.Type.(*types.Signature); ok {
				return types.NewFunc(token.NoPos, nil, "_", sig), nil
			}
		}
		return nil, fmt.Errorf("expander: cannot resolve FuncLit signature")
	default:
		return nil, fmt.Errorf("expander: unexpected enclosing node %T", fnNode)
	}
}

func enclosingFuncNodeAt(file *ast.File, pos token.Pos) ast.Node {
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
