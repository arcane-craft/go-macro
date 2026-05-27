package expander_test

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/arcane-craft/go-macro/internal/expander"
	"github.com/arcane-craft/go-macro/macro"
)

const stubProviderPath = "example.com/macprov"

func validateStubValue(t *testing.T, src string, reg *macro.Registry, wantErr bool) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "u.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	info := typecheckFileWithStubUses(t, fset, f, stubProviderPath, "MacroStub")
	err = expander.ValidateStubValueUsage(fset, f, info, reg)
	if wantErr {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "function value") {
			t.Fatalf("error %q should mention function value", err)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateStubValueArg(t *testing.T) {
	reg := setupProviderReg(t, token.NewFileSet(), stubProviderPath)
	validateStubValue(t, `package u
import mp "`+stubProviderPath+`"
func apply(func(int) int) {}
func f() { apply(mp.MacroStub) }
`, reg, true)
}

func TestValidateStubValueAssign(t *testing.T) {
	reg := setupProviderReg(t, token.NewFileSet(), stubProviderPath)
	validateStubValue(t, `package u
import mp "`+stubProviderPath+`"
func f() { fn := mp.MacroStub }
`, reg, true)
}

func TestValidateStubValueReturn(t *testing.T) {
	reg := setupProviderReg(t, token.NewFileSet(), stubProviderPath)
	validateStubValue(t, `package u
import mp "`+stubProviderPath+`"
func f() func(int) int { return mp.MacroStub }
`, reg, true)
}

func TestValidateStubValueReflect(t *testing.T) {
	reg := setupProviderReg(t, token.NewFileSet(), stubProviderPath)
	validateStubValue(t, `package u
import (
	mp "`+stubProviderPath+`"
	"reflect"
)
func f() { _ = reflect.ValueOf(mp.MacroStub) }
`, reg, true)
}

func TestValidateStubValueDeadCode(t *testing.T) {
	reg := setupProviderReg(t, token.NewFileSet(), stubProviderPath)
	validateStubValue(t, `package u
import mp "`+stubProviderPath+`"
func f() { if false { _ = mp.MacroStub } }
`, reg, true)
}

func TestValidateStubValueDirectCallOK(t *testing.T) {
	reg := setupProviderReg(t, token.NewFileSet(), stubProviderPath)
	validateStubValue(t, `package u
import mp "`+stubProviderPath+`"
func f() int { return mp.MacroStub(1) }
`, reg, false)
	validateStubValue(t, `package u
import mp "`+stubProviderPath+`"
func f() int { return (mp.MacroStub)(1) }
`, reg, false)
}

func TestValidateStubValueNestedCallOK(t *testing.T) {
	reg := setupProviderReg(t, token.NewFileSet(), stubProviderPath)
	validateStubValue(t, `package u
import mp "`+stubProviderPath+`"
func outer(int) int { return 0 }
func f() int { return outer(mp.MacroStub(1)) }
`, reg, false)
}

func TestValidateStubValueUnlinkedOK(t *testing.T) {
	reg := macro.NewRegistry()
	validateStubValue(t, `package u
import mp "`+stubProviderPath+`"
func f() { _ = mp.MacroStub }
`, reg, false)
}

func TestValidateStubValueShadowOK(t *testing.T) {
	reg := setupProviderReg(t, token.NewFileSet(), stubProviderPath)
	validateStubValue(t, `package u
func MacroStub(int) int { return 0 }
func f() int { return MacroStub(1) }
`, reg, false)
	validateStubValue(t, `package u
func MacroStub(int) int { return 0 }
func f() { _ = MacroStub }
`, reg, false)
}

func TestValidateStubValueMethodOK(t *testing.T) {
	reg := setupProviderReg(t, token.NewFileSet(), stubProviderPath)
	validateStubValue(t, `package u
type S struct{}
func (S) MacroStub(int) int { return 0 }
func f() int {
	var s S
	return s.MacroStub(1)
}
`, reg, false)
}

func TestExpandFileStubValueBlocksExpand(t *testing.T) {
	fset := token.NewFileSet()
	reg := setupProviderReg(t, fset, stubProviderPath)
	src := `package u
import mp "` + stubProviderPath + `"
func f() { _ = mp.MacroStub }
`
	f, err := parser.ParseFile(fset, "u.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	before := len(f.Decls)
	info := typecheckFileWithStubUses(t, fset, f, stubProviderPath, "MacroStub")
	engine := &expander.Engine{Registry: reg}
	imports := expander.BuildImportMap(f, "u")
	err = engine.ExpandFile(fset, f, info, nil, imports)
	if err == nil {
		t.Fatal("expected expand error")
	}
	if len(f.Decls) != before {
		t.Fatal("file should not be mutated on value usage error")
	}
}
