package macro

import (
	"fmt"
	"go/token"
)

// ErrorAt reports an expansion error at the given position.
func ErrorAt(fset *token.FileSet, pos token.Pos, format string, args ...any) error {
	if !pos.IsValid() {
		return fmt.Errorf(format, args...)
	}
	posn := fset.Position(pos)
	return fmt.Errorf("%s:%d:%d: %s", posn.Filename, posn.Line, posn.Column, fmt.Sprintf(format, args...))
}
