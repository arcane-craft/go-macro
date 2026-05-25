//go:build macro

//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .

package readfile

import (
	"fmt"
	"os"

	"github.com/arcane-craft/go-macro/contrib/try"
)

// ReadFile reads hello.txt using the Try macro.
func ReadFile() (bs []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("readfile: %w", err)
		}
	}()
	file := try.Try(os.Open("hello.txt"))
	defer file.Close()
	return os.ReadFile("hello.txt")
}
