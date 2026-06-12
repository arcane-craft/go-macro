package expander

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/arcane-craft/go-macro/macro"
	"github.com/arcane-craft/go-macro/macro/pattern"
)

func matchPattern(site *siteSyntax, top pattern.Top) (macro.Bindings, MatchMeta, error) {
	switch top.Form {
	case pattern.FormCall:
		return matchCallPattern(site, *top.Call)
	case pattern.FormStmt:
		return matchStmtPattern(site, *top.Stmt)
	case pattern.FormDecl:
		return matchDeclPattern(site, *top.Decl)
	default:
		return nil, MatchMeta{}, fmt.Errorf("macro: unknown pattern form")
	}
}

func matchCallPattern(site *siteSyntax, pat pattern.Call) (macro.Bindings, MatchMeta, error) {
	call, ok := site.anchor.(*ast.CallExpr)
	if !ok {
		return nil, MatchMeta{}, fmt.Errorf("macro: CallPattern requires Call anchor")
	}
	binds := newBindings()
	if err := matchCallNode(call, pat, binds, call); err != nil {
		return nil, MatchMeta{}, err
	}
	plan, err := planForCallParent(site.file, call)
	if err != nil {
		return nil, MatchMeta{}, err
	}
	return binds, MatchMeta{
		Bindings:    binds,
		MatchedSpan: call,
		Plan:        plan,
		MatchRoot:   MatchRootCall,
	}, nil
}

func matchStmtPattern(site *siteSyntax, pat pattern.Stmt) (macro.Bindings, MatchMeta, error) {
	call, ok := site.anchor.(*ast.CallExpr)
	if !ok {
		return nil, MatchMeta{}, fmt.Errorf("macro: StmtPattern requires Call anchor")
	}
	binds := newBindings()
	var matched ast.Stmt
	var err error

	switch pat.Kind {
	case pattern.StmtAssignDefine:
		matched, err = matchAssignStmt(site.file, call, token.DEFINE, pat, binds)
	case pattern.StmtAssignPlain:
		matched, err = matchAssignStmt(site.file, call, token.ASSIGN, pat, binds)
	case pattern.StmtVar:
		matched, err = matchVarStmt(site.file, call, pat, binds)
	case pattern.StmtReturn:
		matched, err = matchReturnStmt(site.file, call, pat, binds)
	case pattern.StmtExprSemi:
		matched, err = matchExprStmtSemi(site.file, call, pat.Call, binds)
	default:
		return nil, MatchMeta{}, fmt.Errorf("macro: unknown stmt pattern kind")
	}
	if err != nil {
		return nil, MatchMeta{}, err
	}
	block, idx, ok := findStmtInBlock(site.file, matched)
	if !ok {
		return nil, MatchMeta{}, fmt.Errorf("macro: cannot locate matched stmt in block")
	}
	plan := []SpliceStep{{
		Replace: &ReplaceInContainer{
			Parent:         block,
			ContainerField: ContainerBlockStmts,
			Index:          idx,
			Mode:           SpliceOneToMany,
		},
	}}
	return binds, MatchMeta{
		Bindings:    binds,
		MatchedSpan: matched,
		Plan:        plan,
		MatchRoot:   MatchRootStmt,
	}, nil
}

func matchDeclPattern(site *siteSyntax, pat pattern.Decl) (macro.Bindings, MatchMeta, error) {
	embed, ok := site.anchor.(*ast.Field)
	if !ok {
		return nil, MatchMeta{}, fmt.Errorf("macro: DeclPattern requires Field anchor")
	}
	ts, gen, err := enclosingTypeSpec(site.file, embed)
	if err != nil {
		return nil, MatchMeta{}, err
	}
	binds := newBindings()
	binds.singles[pat.ItemName] = macro.WrapNode(ts)

	st, ok := ts.Type.(*ast.StructType)
	if !ok || st.Fields == nil {
		return nil, MatchMeta{}, fmt.Errorf("macro: DeclPattern requires struct type")
	}

	var embedPat *pattern.DeclConstraint
	var fieldPat *pattern.DeclConstraint
	for i := range pat.Fields {
		c := &pat.Fields[i]
		switch c.Kind {
		case pattern.DeclEmbedMarker:
			if embedPat != nil {
				return nil, MatchMeta{}, fmt.Errorf("macro: multiple embed markers in pattern")
			}
			embedPat = c
		case pattern.DeclFieldEllipsis:
			if fieldPat != nil {
				return nil, MatchMeta{}, fmt.Errorf("macro: multiple field ellipsis in pattern")
			}
			fieldPat = c
		}
	}
	if embedPat == nil {
		return nil, MatchMeta{}, fmt.Errorf("macro: DeclPattern requires embed marker")
	}

	var matchedEmbed *ast.Field
	var namedFields []*ast.Field
	for _, f := range st.Fields.List {
		if f.Names == nil {
			if err := matchEmbedField(f, *embedPat, binds); err == nil {
				if matchedEmbed != nil {
					return nil, MatchMeta{}, fmt.Errorf("macro: multiple embed markers in struct")
				}
				matchedEmbed = f
			}
			continue
		}
		namedFields = append(namedFields, f)
	}
	if matchedEmbed == nil {
		return nil, MatchMeta{}, fmt.Errorf("macro: no embed marker in struct")
	}
	if matchedEmbed != embed {
		return nil, MatchMeta{}, fmt.Errorf("macro: embed anchor does not match pattern marker")
	}

	if fieldPat != nil {
		elems := make([]macro.Syntax, len(namedFields))
		for i, f := range namedFields {
			elems[i] = macro.WrapNode(f)
		}
		binds.lists[fieldPat.Field] = elems
	} else if len(namedFields) > 0 {
		return nil, MatchMeta{}, fmt.Errorf("macro: struct has named fields but pattern has no $field ...")
	}

	specIdx := -1
	for i, spec := range gen.Specs {
		if spec == ts {
			specIdx = i
			break
		}
	}
	if specIdx < 0 {
		return nil, MatchMeta{}, fmt.Errorf("macro: cannot find TypeSpec in GenDecl")
	}

	plan := []SpliceStep{
		{
			Replace: &ReplaceInContainer{
				Parent:         gen,
				ContainerField: ContainerGenDeclSpecs,
				Index:          specIdx,
				Mode:           SpliceOneToOne,
			},
		},
		{
			InsertAfter: &InsertAfterInFileDecls{After: gen},
		},
	}
	return binds, MatchMeta{
		Bindings:    binds,
		MatchedSpan: ts,
		Plan:        plan,
		MatchRoot:   MatchRootDecl,
	}, nil
}

func matchCallNode(call *ast.CallExpr, pat pattern.Call, binds *bindings, anchor *ast.CallExpr) error {
	if call != anchor {
		return fmt.Errorf("macro: call anchor mismatch")
	}
	if !calleeMatches(call.Fun, pat.Callee) {
		return fmt.Errorf("macro: callee mismatch")
	}
	if len(pat.Args) != len(call.Args) {
		return fmt.Errorf("macro: argument count mismatch")
	}
	for i, argPat := range pat.Args {
		if argPat.Kind == pattern.ArgDiscard {
			continue
		}
		binds.singles[argPat.Name] = macro.WrapExpr(call.Args[i])
	}
	return nil
}

func calleeMatches(fun ast.Expr, pat pattern.Callee) bool {
	switch f := fun.(type) {
	case *ast.Ident:
		if pat.Selector != "" {
			return false
		}
		return f.Name == pat.Ident
	case *ast.SelectorExpr:
		if pat.Selector != "" {
			if sel, ok := f.X.(*ast.Ident); ok {
				return sel.Name == pat.Ident && f.Sel.Name == pat.Selector
			}
			return false
		}
		return f.Sel.Name == pat.Ident
	default:
		return false
	}
}

func invokedName(fun ast.Expr) string {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	default:
		return ""
	}
}

func matchAssignStmt(file *ast.File, anchor *ast.CallExpr, tok token.Token, pat pattern.Stmt, binds *bindings) (ast.Stmt, error) {
	var matched *ast.AssignStmt
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || assign.Tok != tok {
			return true
		}
		for _, rhs := range assign.Rhs {
			if rhs == anchor || unwrapParen(rhs) == anchor {
				matched = assign
				return false
			}
		}
		return true
	})
	if matched == nil {
		return nil, fmt.Errorf("macro: assign stmt not found")
	}
	elems := make([]macro.Syntax, len(matched.Lhs))
	for i, lhs := range matched.Lhs {
		elems[i] = macro.WrapExpr(lhs)
	}
	binds.lists[pat.LHS] = elems
	if err := matchCallNode(anchor, pat.Call, binds, anchor); err != nil {
		return nil, err
	}
	return matched, nil
}

func matchVarStmt(file *ast.File, anchor *ast.CallExpr, pat pattern.Stmt, binds *bindings) (ast.Stmt, error) {
	var matched *ast.DeclStmt
	ast.Inspect(file, func(n ast.Node) bool {
		ds, ok := n.(*ast.DeclStmt)
		if !ok {
			return true
		}
		gd, ok := ds.Decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			return true
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok || vs.Values == nil {
				continue
			}
			for _, v := range vs.Values {
				if v == anchor || unwrapParen(v) == anchor {
					matched = ds
					return false
				}
			}
		}
		return true
	})
	if matched == nil {
		return nil, fmt.Errorf("macro: var stmt not found")
	}
	gd := matched.Decl.(*ast.GenDecl)
	var names []macro.Syntax
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, n := range vs.Names {
			names = append(names, macro.WrapExpr(n))
		}
	}
	binds.lists[pat.LHS] = names
	if err := matchCallNode(anchor, pat.Call, binds, anchor); err != nil {
		return nil, err
	}
	return matched, nil
}

func matchReturnStmt(file *ast.File, anchor *ast.CallExpr, pat pattern.Stmt, binds *bindings) (ast.Stmt, error) {
	var matched *ast.ReturnStmt
	ast.Inspect(file, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for i, r := range ret.Results {
			if r == anchor || unwrapParen(r) == anchor {
				matched = ret
				if i > 0 {
					prefix := make([]macro.Syntax, i)
					for j := 0; j < i; j++ {
						prefix[j] = macro.WrapExpr(ret.Results[j])
					}
					binds.lists[pat.LHS] = prefix
				} else {
					binds.lists[pat.LHS] = []macro.Syntax{}
				}
				return false
			}
		}
		return true
	})
	if matched == nil {
		return nil, fmt.Errorf("macro: return stmt not found")
	}
	if err := matchCallNode(anchor, pat.Call, binds, anchor); err != nil {
		return nil, err
	}
	return matched, nil
}

func matchExprStmtSemi(file *ast.File, anchor *ast.CallExpr, pat pattern.Call, binds *bindings) (ast.Stmt, error) {
	var matched *ast.ExprStmt
	ast.Inspect(file, func(n ast.Node) bool {
		es, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}
		if es.X == anchor || unwrapParen(es.X) == anchor {
			matched = es
			return false
		}
		return true
	})
	if matched == nil {
		return nil, fmt.Errorf("macro: ExprStmt not found")
	}
	if err := matchCallNode(anchor, pat, binds, anchor); err != nil {
		return nil, err
	}
	return matched, nil
}

func matchEmbedField(field *ast.Field, pat pattern.DeclConstraint, binds *bindings) error {
	if pat.Index.Kind == pattern.ArgDiscard {
		baseName := typeExprInvokedName(field.Type)
		if pat.Marker.Selector != "" {
			if !typeExprMatchesSelector(field.Type, pat.Marker) {
				return fmt.Errorf("macro: embed marker mismatch")
			}
		} else if baseName != pat.Marker.Ident {
			return fmt.Errorf("macro: embed marker name mismatch")
		}
		return nil
	}
	typ, ok := field.Type.(*ast.IndexExpr)
	if !ok {
		return fmt.Errorf("macro: embed is not index expression")
	}
	baseName := typeExprInvokedName(typ.X)
	if pat.Marker.Selector != "" {
		if !typeExprMatchesSelector(typ.X, pat.Marker) {
			return fmt.Errorf("macro: embed marker mismatch")
		}
	} else if baseName != pat.Marker.Ident {
		return fmt.Errorf("macro: embed marker name mismatch")
	}
	if pat.Index.Kind == pattern.ArgCapture {
		binds.singles[pat.Index.Name] = macro.WrapExpr(typ.Index)
	}
	return nil
}

func typeExprInvokedName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return e.Sel.Name
	default:
		return ""
	}
}

func typeExprMatchesSelector(expr ast.Expr, pat pattern.Callee) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	left, ok := sel.X.(*ast.Ident)
	return ok && left.Name == pat.Ident && sel.Sel.Name == pat.Selector
}

func planForCallParent(file *ast.File, call *ast.CallExpr) ([]SpliceStep, error) {
	if _, stmt, _, rhsIdx, ok := findAssignRHSContainingCall(file, call); ok {
		return []SpliceStep{{
			Replace: &ReplaceInContainer{
				Parent:         stmt,
				ContainerField: ContainerAssignRhs,
				Index:          rhsIdx,
				Mode:           SpliceOneToOne,
			},
		}}, nil
	}
	if callInReturnResults(file, call) {
		_, stmt, ok := findReturnStmt(file, call)
		if !ok {
			return nil, fmt.Errorf("macro: cannot plan ReturnResults")
		}
		return []SpliceStep{{
			Replace: &ReplaceInContainer{
				Parent:         stmt,
				ContainerField: ContainerReturnResults,
				Mode:           SpliceReplaceAll,
			},
		}}, nil
	}
	if _, stmt, ri, ok := findReturnContainingCall(file, call); ok {
		return []SpliceStep{{
			Replace: &ReplaceInContainer{
				Parent:         stmt,
				ContainerField: ContainerReturnResults,
				Index:          ri,
				Mode:           SpliceOneToOne,
			},
		}}, nil
	}
	if _, stmt, ok := findExprStmtContainingCall(file, call); ok {
		return []SpliceStep{{
			Replace: &ReplaceInContainer{
				Parent:         stmt,
				ContainerField: ContainerExprSlot,
				Mode:           SpliceOneToOne,
			},
		}}, nil
	}
	return nil, fmt.Errorf("macro: unsupported call parent context")
}

func enclosingTypeSpec(file *ast.File, embed *ast.Field) (*ast.TypeSpec, *ast.GenDecl, error) {
	var ts *ast.TypeSpec
	var gen *ast.GenDecl
	ast.Inspect(file, func(n ast.Node) bool {
		gd, ok := n.(*ast.GenDecl)
		if !ok || gd.Tok != token.TYPE {
			return true
		}
		for _, spec := range gd.Specs {
			tsp, ok := spec.(*ast.TypeSpec)
			if !ok || tsp.Type == nil {
				continue
			}
			st, ok := tsp.Type.(*ast.StructType)
			if !ok || st.Fields == nil {
				continue
			}
			for _, f := range st.Fields.List {
				if f == embed {
					ts = tsp
					gen = gd
					return false
				}
			}
		}
		return true
	})
	if ts == nil {
		return nil, nil, fmt.Errorf("macro: cannot find enclosing TypeSpec for embed")
	}
	return ts, gen, nil
}

func findStmtInBlock(file *ast.File, stmt ast.Stmt) (*ast.BlockStmt, int, bool) {
	var block *ast.BlockStmt
	idx := -1
	ast.Inspect(file, func(n ast.Node) bool {
		b, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for i, s := range b.List {
			if s == stmt {
				block = b
				idx = i
				return false
			}
		}
		return true
	})
	return block, idx, block != nil && idx >= 0
}

func findReturnStmt(file *ast.File, call *ast.CallExpr) (*ast.BlockStmt, *ast.ReturnStmt, bool) {
	var block *ast.BlockStmt
	var ret *ast.ReturnStmt
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		b, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for _, s := range b.List {
			r, ok := s.(*ast.ReturnStmt)
			if !ok {
				continue
			}
			for _, res := range r.Results {
				if res == call || unwrapParen(res) == call {
					block = b
					ret = r
					found = true
					return false
				}
			}
		}
		return true
	})
	return block, ret, found
}

func callInReturnResults(file *ast.File, call *ast.CallExpr) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, r := range ret.Results {
			if r == call || unwrapParen(r) == call {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func exprContainsCall(expr ast.Expr, call *ast.CallExpr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok && c == call {
			found = true
			return false
		}
		return true
	})
	return found
}

func findAssignRHSContainingCall(file *ast.File, call *ast.CallExpr) (*ast.BlockStmt, *ast.AssignStmt, int, int, bool) {
	var block *ast.BlockStmt
	var assign *ast.AssignStmt
	var stmtIdx, rhsIdx int
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		b, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for si, s := range b.List {
			a, ok := s.(*ast.AssignStmt)
			if !ok {
				continue
			}
			for ri, rhs := range a.Rhs {
				if exprContainsCall(rhs, call) {
					block = b
					assign = a
					stmtIdx = si
					rhsIdx = ri
					found = true
					return false
				}
			}
		}
		return true
	})
	return block, assign, stmtIdx, rhsIdx, found
}

func findReturnContainingCall(file *ast.File, call *ast.CallExpr) (*ast.BlockStmt, *ast.ReturnStmt, int, bool) {
	var block *ast.BlockStmt
	var ret *ast.ReturnStmt
	var resultIdx int
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		b, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for _, s := range b.List {
			r, ok := s.(*ast.ReturnStmt)
			if !ok {
				continue
			}
			for i, res := range r.Results {
				if exprContainsCall(res, call) {
					block = b
					ret = r
					resultIdx = i
					found = true
					return false
				}
			}
		}
		return true
	})
	return block, ret, resultIdx, found
}

func findExprStmtContainingCall(file *ast.File, call *ast.CallExpr) (*ast.BlockStmt, *ast.ExprStmt, bool) {
	var block *ast.BlockStmt
	var es *ast.ExprStmt
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		b, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}
		for _, s := range b.List {
			xs, ok := s.(*ast.ExprStmt)
			if !ok {
				continue
			}
			if exprContainsCall(xs.X, call) {
				block = b
				es = xs
				found = true
				return false
			}
		}
		return true
	})
	return block, es, found
}

func replaceCallInExpr(root ast.Expr, call *ast.CallExpr, repl ast.Expr) ast.Expr {
	if root == nil {
		return repl
	}
	if root == call || unwrapParen(root) == call {
		return repl
	}
	switch x := root.(type) {
	case *ast.BinaryExpr:
		return &ast.BinaryExpr{X: replaceCallInExpr(x.X, call, repl), Op: x.Op, Y: replaceCallInExpr(x.Y, call, repl)}
	case *ast.UnaryExpr:
		return &ast.UnaryExpr{Op: x.Op, X: replaceCallInExpr(x.X, call, repl)}
	case *ast.ParenExpr:
		return &ast.ParenExpr{X: replaceCallInExpr(x.X, call, repl)}
	case *ast.CallExpr:
		out := &ast.CallExpr{Fun: replaceCallInExpr(x.Fun, call, repl)}
		for _, a := range x.Args {
			out.Args = append(out.Args, replaceCallInExpr(a, call, repl))
		}
		return out
	case *ast.SelectorExpr:
		return &ast.SelectorExpr{X: replaceCallInExpr(x.X, call, repl), Sel: x.Sel}
	default:
		return root
	}
}

type bindings struct {
	singles map[string]macro.Syntax
	lists   map[string][]macro.Syntax
}

func newBindings() *bindings {
	return &bindings{
		singles: make(map[string]macro.Syntax),
		lists:   make(map[string][]macro.Syntax),
	}
}

func (b *bindings) Get(name string) (macro.Syntax, bool) {
	v, ok := b.singles[name]
	return v, ok
}

func (b *bindings) Elems(name string) ([]macro.Syntax, bool) {
	v, ok := b.lists[name]
	if !ok {
		return nil, false
	}
	return v, true
}
