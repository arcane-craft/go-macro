package macro

import (
	"fmt"
	"go/ast"
	"strings"
)

// SpliceTarget names which AST node an expansion replaces.
type SpliceTarget int

const (
	spliceTargetInvalid SpliceTarget = iota // zero value = unset
	SpliceReplaceAssignStmt
	SpliceReplaceAssignRHS
	SpliceReplaceReturnStmt
	SpliceReplaceReturnResults
	SpliceReplaceExprStmt
	SpliceReplaceCallExpr
)

var spliceTargetNames = [...]string{
	"(unset)",
	"SpliceReplaceAssignStmt",
	"SpliceReplaceAssignRHS",
	"SpliceReplaceReturnStmt",
	"SpliceReplaceReturnResults",
	"SpliceReplaceExprStmt",
	"SpliceReplaceCallExpr",
}

// String returns the target name for error messages.
func (t SpliceTarget) String() string {
	if int(t) >= 0 && int(t) < len(spliceTargetNames) {
		return spliceTargetNames[t]
	}
	return fmt.Sprintf("SpliceTarget(%d)", t)
}

// LegalSpliceTargetsForCall returns splice targets structurally valid at call.
func LegalSpliceTargetsForCall(file *ast.File, call *ast.CallExpr) []SpliceTarget {
	if file == nil || call == nil {
		return nil
	}
	if callInAssignRHS(file, call) {
		return []SpliceTarget{SpliceReplaceAssignRHS, SpliceReplaceAssignStmt}
	}
	if callInReturnResults(file, call) {
		return []SpliceTarget{SpliceReplaceReturnResults, SpliceReplaceReturnStmt}
	}
	if callIsExprStmt(file, call) {
		return []SpliceTarget{SpliceReplaceExprStmt, SpliceReplaceCallExpr}
	}
	return []SpliceTarget{SpliceReplaceCallExpr}
}

// ValidateExpandResult checks Target, payload shape, and structural legality.
func ValidateExpandResult(ctx Context, result ExpandResult) error {
	if ctx == nil {
		return fmt.Errorf("macro: nil Context")
	}
	return ValidateExpandResultForCall(ctx.File(), ctx.Call(), result)
}

// ValidateExpandResultForCall validates without a full Context (used by the engine).
func ValidateExpandResultForCall(file *ast.File, call *ast.CallExpr, result ExpandResult) error {
	return validateExpandResult(file, call, result)
}

func validateExpandResult(file *ast.File, call *ast.CallExpr, result ExpandResult) error {
	if err := validatePayload(result); err != nil {
		return err
	}
	legal := LegalSpliceTargetsForCall(file, call)
	if !containsSpliceTarget(legal, result.Target) {
		return fmt.Errorf("macro: %s invalid at this call site; legal targets: %s",
			result.Target, formatSpliceTargets(legal))
	}
	return nil
}

func validatePayload(result ExpandResult) error {
	if result.Target == spliceTargetInvalid {
		return fmt.Errorf("macro: ExpandResult.Target is required")
	}
	switch result.Target {
	case SpliceReplaceAssignStmt, SpliceReplaceReturnStmt, SpliceReplaceExprStmt:
		if len(result.Stmts) == 0 {
			return fmt.Errorf("macro: %s requires non-empty Stmts", result.Target)
		}
		if len(result.Exprs) > 0 || result.Expr != nil {
			return fmt.Errorf("macro: %s must not set Expr or Exprs", result.Target)
		}
	case SpliceReplaceAssignRHS, SpliceReplaceCallExpr:
		if result.Expr == nil {
			return fmt.Errorf("macro: %s requires Expr", result.Target)
		}
		if len(result.Stmts) > 0 || len(result.Exprs) > 0 {
			return fmt.Errorf("macro: %s must not set Stmts or Exprs", result.Target)
		}
	case SpliceReplaceReturnResults:
		if len(result.Exprs) == 0 {
			return fmt.Errorf("macro: %s requires non-empty Exprs", result.Target)
		}
		if len(result.Stmts) > 0 || result.Expr != nil {
			return fmt.Errorf("macro: %s must not set Stmts or Expr", result.Target)
		}
	default:
		return fmt.Errorf("macro: unknown ExpandResult.Target %v", result.Target)
	}
	return nil
}

func containsSpliceTarget(list []SpliceTarget, t SpliceTarget) bool {
	for _, x := range list {
		if x == t {
			return true
		}
	}
	return false
}

func formatSpliceTargets(list []SpliceTarget) string {
	if len(list) == 0 {
		return "(none)"
	}
	names := make([]string, len(list))
	for i, t := range list {
		names[i] = t.String()
	}
	return strings.Join(names, ", ")
}

func callInAssignRHS(file *ast.File, call *ast.CallExpr) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, rhs := range assign.Rhs {
			if rhs == call || unwrapParenExpr(rhs) == call {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func callInReturnResults(file *ast.File, call *ast.CallExpr) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, r := range ret.Results {
			if r == call || unwrapParenExpr(r) == call {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

func callIsExprStmt(file *ast.File, call *ast.CallExpr) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		stmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}
		if exprContainsCall(stmt.X, call) {
			found = true
			return false
		}
		return true
	})
	return found
}

func exprContainsCall(expr ast.Expr, call *ast.CallExpr) bool {
	if expr == nil {
		return false
	}
	if expr == call || unwrapParenExpr(expr) == call {
		return true
	}
	contained := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if n == call {
			contained = true
			return false
		}
		return true
	})
	return contained
}

func unwrapParenExpr(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}
