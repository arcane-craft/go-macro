package macro

import (
	"fmt"
	"go/ast"
)

// ValidateDeclExpandResult checks that a successful DeclExpandResult is complete.
func ValidateDeclExpandResult(ctx DeclContext, result DeclExpandResult) error {
	if ctx == nil {
		return fmt.Errorf("macro: nil DeclContext")
	}
	return validateDeclExpandResult(ctx.Site().Target, result)
}

func validateDeclExpandResult(target *ast.TypeSpec, result DeclExpandResult) error {
	if target == nil {
		return fmt.Errorf("macro: nil Target")
	}
	if result.Fields == nil {
		return fmt.Errorf("macro: DeclExpandResult.Fields is required on success")
	}
	if result.Methods == nil {
		return fmt.Errorf("macro: DeclExpandResult.Methods is required on success")
	}
	targetName := target.Name.Name
	for _, fn := range result.Methods {
		if fn == nil {
			return fmt.Errorf("macro: DeclExpandResult.Methods contains nil")
		}
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			return fmt.Errorf("macro: method %s missing receiver", fn.Name.Name)
		}
		if recvNameForType(fn.Recv.List[0].Type) != targetName {
			return fmt.Errorf("macro: method %s receiver must be %s", fn.Name.Name, targetName)
		}
	}
	return nil
}
