package expander

import (
	"fmt"
	"go/ast"

	"github.com/arcane-craft/go-macro/macro"
)

// ValidateSplice checks out shape against meta.Plan.
func ValidateSplice(out macro.Syntax, meta MatchMeta) error {
	if len(meta.Plan) == 0 {
		return fmt.Errorf("macro: empty splice plan")
	}
	for _, step := range meta.Plan {
		if step.Replace != nil {
			if err := validateReplaceStep(out, *step.Replace, meta); err != nil {
				return err
			}
			continue
		}
		if step.InsertAfter != nil {
			decls, err := out.ToDecls()
			if err != nil {
				// TypeSpec-only out: no generated methods to validate.
				continue
			}
			if len(decls) >= 2 {
				for _, d := range decls[1:] {
					if _, ok := d.(*ast.FuncDecl); !ok {
						return fmt.Errorf("macro: InsertAfter tail must be FuncDecl")
					}
				}
			}
			continue
		}
		return fmt.Errorf("macro: empty splice step")
	}
	return nil
}

func validateReplaceStep(out macro.Syntax, rep ReplaceInContainer, meta MatchMeta) error {
	switch rep.ContainerField {
	case ContainerBlockStmts:
		if rep.Mode != SpliceOneToMany {
			return fmt.Errorf("macro: BlockStmts requires OneToMany")
		}
		_, err := out.ToStmts()
		return err
	case ContainerAssignRhs, ContainerExprSlot:
		if rep.Mode != SpliceOneToOne {
			return fmt.Errorf("macro: expr slot requires OneToOne")
		}
		_, err := out.ToExpr()
		return err
	case ContainerReturnResults:
		switch rep.Mode {
		case SpliceReplaceAll:
			exprs, err := out.ToExprs()
			if err != nil {
				return err
			}
			if len(exprs) == 0 {
				return fmt.Errorf("macro: ReturnResults requires non-empty Exprs")
			}
			return nil
		case SpliceOneToOne:
			_, err := out.ToExpr()
			return err
		default:
			return fmt.Errorf("macro: ReturnResults requires ReplaceAll or OneToOne")
		}
	case ContainerGenDeclSpecs:
		if rep.Mode != SpliceOneToOne {
			return fmt.Errorf("macro: GenDeclSpecs requires OneToOne")
		}
		ts, err := firstTypeSpecFromOut(out)
		if err != nil {
			return err
		}
		spanTS, ok := meta.MatchedSpan.(*ast.TypeSpec)
		if !ok {
			return fmt.Errorf("macro: MatchedSpan must be TypeSpec")
		}
		if ts.Name.Name != spanTS.Name.Name {
			return fmt.Errorf("macro: TypeSpec name %q must match MatchedSpan %q", ts.Name.Name, spanTS.Name.Name)
		}
		return nil
	default:
		return fmt.Errorf("macro: unknown container field %v", rep.ContainerField)
	}
}

// Apply executes meta.Plan on file using out payload.
func Apply(file *ast.File, meta MatchMeta, out macro.Syntax) error {
	for _, step := range meta.Plan {
		if step.Replace != nil {
			if err := applyReplace(file, *step.Replace, out, meta.MatchedSpan); err != nil {
				return err
			}
			continue
		}
		if step.InsertAfter != nil {
			decls, err := out.ToDecls()
			if err == nil && len(decls) >= 2 {
				insertAfterGenDecl(file, step.InsertAfter.After, decls[1:])
			}
			continue
		}
		return fmt.Errorf("macro: empty splice step")
	}
	return nil
}

func applyReplace(file *ast.File, rep ReplaceInContainer, out macro.Syntax, matched ast.Node) error {
	switch rep.ContainerField {
	case ContainerBlockStmts:
		stmts, err := out.ToStmts()
		if err != nil {
			return err
		}
		block, ok := rep.Parent.(*ast.BlockStmt)
		if !ok {
			return fmt.Errorf("macro: BlockStmts parent must be BlockStmt")
		}
		if rep.Index < 0 || rep.Index >= len(block.List) {
			return fmt.Errorf("macro: BlockStmts index out of range")
		}
		old := block.List[rep.Index]
		prefix := append([]ast.Stmt(nil), block.List[:rep.Index]...)
		block.List = append(append(prefix, stmts...), block.List[rep.Index+1:]...)
		_ = old
		return nil
	case ContainerAssignRhs:
		expr, err := out.ToExpr()
		if err != nil {
			return err
		}
		call, ok := matched.(*ast.CallExpr)
		if !ok {
			return fmt.Errorf("macro: AssignRhs requires CallExpr MatchedSpan")
		}
		assign, ok := rep.Parent.(*ast.AssignStmt)
		if !ok {
			return fmt.Errorf("macro: AssignRhs parent must be AssignStmt")
		}
		if rep.Index < 0 || rep.Index >= len(assign.Rhs) {
			return fmt.Errorf("macro: AssignRhs index out of range")
		}
		assign.Rhs[rep.Index] = replaceCallInExpr(assign.Rhs[rep.Index], call, expr)
		return nil
	case ContainerReturnResults:
		ret, ok := rep.Parent.(*ast.ReturnStmt)
		if !ok {
			return fmt.Errorf("macro: ReturnResults parent must be ReturnStmt")
		}
		switch rep.Mode {
		case SpliceOneToOne:
			expr, err := out.ToExpr()
			if err != nil {
				return err
			}
			call, ok := matched.(*ast.CallExpr)
			if !ok {
				return fmt.Errorf("macro: ReturnResults OneToOne requires CallExpr MatchedSpan")
			}
			if rep.Index < 0 || rep.Index >= len(ret.Results) {
				return fmt.Errorf("macro: ReturnResults index out of range")
			}
			ret.Results[rep.Index] = replaceCallInExpr(ret.Results[rep.Index], call, expr)
			return nil
		case SpliceReplaceAll:
			exprs, err := out.ToExprs()
			if err != nil {
				return err
			}
			ret.Results = exprs
			return nil
		default:
			return fmt.Errorf("macro: ReturnResults requires ReplaceAll or OneToOne")
		}
	case ContainerExprSlot:
		expr, err := out.ToExpr()
		if err != nil {
			return err
		}
		call, ok := matched.(*ast.CallExpr)
		if !ok {
			return fmt.Errorf("macro: ExprSlot requires CallExpr MatchedSpan")
		}
		es, ok := rep.Parent.(*ast.ExprStmt)
		if !ok {
			return fmt.Errorf("macro: ExprSlot parent must be ExprStmt")
		}
		es.X = replaceCallInExpr(es.X, call, expr)
		return nil
	case ContainerGenDeclSpecs:
		ts, err := firstTypeSpecFromOut(out)
		if err != nil {
			return err
		}
		gen, ok := rep.Parent.(*ast.GenDecl)
		if !ok {
			return fmt.Errorf("macro: GenDeclSpecs parent must be GenDecl")
		}
		if rep.Index < 0 || rep.Index >= len(gen.Specs) {
			return fmt.Errorf("macro: GenDeclSpecs index out of range")
		}
		gen.Specs[rep.Index] = ts
		return nil
	default:
		return fmt.Errorf("macro: unknown container field %v", rep.ContainerField)
	}
}

func firstTypeSpecFromOut(out macro.Syntax) (*ast.TypeSpec, error) {
	if ts, ok := out.Underlying().(*ast.TypeSpec); ok {
		return ts, nil
	}
	decls, err := out.ToDecls()
	if err != nil {
		return nil, err
	}
	for _, d := range decls {
		if gd, ok := d.(*ast.GenDecl); ok {
			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					return ts, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("macro: out has no TypeSpec")
}

func insertAfterGenDecl(file *ast.File, after *ast.GenDecl, decls []ast.Decl) {
	idx := -1
	for i, d := range file.Decls {
		if d == after {
			idx = i
			break
		}
	}
	if idx < 0 {
		file.Decls = append(file.Decls, decls...)
		return
	}
	prefix := append([]ast.Decl(nil), file.Decls[:idx+1]...)
	file.Decls = append(append(prefix, decls...), file.Decls[idx+1:]...)
}
