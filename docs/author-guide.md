# 宏作者指南

本文面向**编写宏库**的开发者：你要定义「宏长什么样、展开后变成什么」。若你只是在项目里**使用**已有宏库，请直接看 [README 快速上手](../README.md#快速上手)。

## 阅读指引

| 你是谁 | 从这里开始 |
| ------ | ---------- |
| **编写**宏库 | [用脚手架初始化](#用脚手架初始化-provider) → [框架契约](#框架契约) |
| 在工程里**使用**宏 | [README 快速上手](../README.md#快速上手)，或 [宏使用方](#宏使用方) |
| 官方宏库、本地联调 | [参考](#参考) |

## 角色分工

go-macro 把三件事分开：框架负责展开流水线，宏作者负责定义语法与展开逻辑，使用方在业务代码里调用宏。

| 角色 | 做什么 |
| ---- | ------ |
| **框架** | 提供 `cmd/macro expand` 与 `macro/expandtool`；按 import 自动生成 `.gomacro/expand_runner` 并完成展开 |
| **宏作者** | 编写 Call 语法桩 / Decl marker 与对应 `XxxExpand`；在桩或 marker 类型 doc 上标注 `//macro: <syntax-id>` |
| **宏使用方** | 在宏主文件 import 宏库，用 `go:generate` 调用 `cmd/macro expand` |

`[examples/](../examples/)` 演示宏使用方如何接入，可参考 [examples/readfile](../examples/readfile/readfile.go)。

## 编写宏库

### 用脚手架初始化 provider

在项目根目录执行：

```bash
go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac
```

脚手架会生成最小 provider 骨架：一个语法桩、`MacroExpand` 占位实现，以及 `expand_test.go` 模板。桩与 Expander 的 doc 里已带好 `//macro:`，你只需补全展开逻辑。

### 框架契约

框架支持 Call / Decl 两类站点，统一为 **syntax-rules** 模型。同一 provider 包可注册多个 `syntax-id`；推荐 `expandtool.Register(syntaxID, expander)` 注册统一 `Expander`。展开时引擎**先 Decl、后 Call**。

| 轨道 | 触发方式 | 推荐 Expander |
| ---- | -------- | ------------- |
| **Call** | 函数体内 `Stub(...)` | `macro.SyntaxCase` / `macro.SyntaxRules` |
| **Decl** | struct **匿名嵌入** marker | `macro.SyntaxCase` / `macro.SyntaxRules` |

统一签名：`func(ctx Context, site Syntax) (Syntax, error)`。`Context` 仅含 `FileSet()`、`Types()`、`TempIdent()`；位置用 `site.MacroPos()`；外层签名用 `EnclosingSignature` / `EnclosingResults`。

**Pattern 首版子集**（`site.Match`）：`Try($inner)`、`$lhs ... := Try($inner)`、`var $lhs ... =`、`return $vals ... ,`、`Try($inner);`、`type $item struct { Derive[$iface] $field ... }`。贴回由 Match 产出 `Plan`，经 `ValidateSplice` + `Apply` 执行。

**Decl embed 元数据**（旧 `DeclSite` 对照）：

| 旧 API | 新路径 |
| ------ | ------ |
| `site.Target` / 类型名 | `binds.Get("item")` |
| 字段列表 | `binds.Elems("field")`，每项 `Underlying()` 为 `*ast.Field`（含 `Tag`） |
| 类型实参 | `binds.Get("iface")` + `ctx.Types().TypeOf(expr)` |
| `MacroTag` | `site.Underlying().(*ast.Field).Tag` + `ParseMacroTag` |
| `TargetMethods()` | 不恢复；未 match 的 methods 自动保留 |

Quote 使用 `macro.Quote(template, map[string]Syntax)`（`#` 洞）；无需 import 独立子包。

#### 标注 `//macro:`

在以下位置的 doc 里写 `//macro: <syntax-id>`：

- Call 语法桩
- Call Expander
- Decl marker 类型
- Decl Expander

同一 syntax-id 下可以有多个 Call 桩名，但至多一个 Call Expander 与一个 Decl Expander。

**Call 桩示例：**

```go
//macro: syntax-mine
// Mine 是语法桩，运行时不要直接调用。
func Mine[T any](v T) T {
    panic("Mine is a macro stub")
}
```

**Call Expander 示例：**

```go
//macro: syntax-mine
var MineExpander = macro.SyntaxRules(macro.Clause{
    Pattern:   "Mine($v)",
    Template:  "#v",
})
```

#### Call 宏：语法桩约定

编写 Call 桩时，请遵循这些约定：

- 桩是**包级函数**（无 receiver）
- doc 里写 `//macro: <syntax-id>`
- 函数体建议 `panic(...)`，避免运行时被误调用
- 使用方须**直接调用**语法桩：`pkg.Stub(...)`（或 dot-import 下的 `Stub(...)`）
- 桩只能作为调用表达式出现；若把桩当作函数值传递、赋值、返回，或传给 `reflect.ValueOf` / `reflect.TypeOf`，`expand` 会报错

#### Call 宏：Match 与贴回

展开 Call 宏时，pattern match 决定 **MatchedSpan**（引擎替换哪段 AST）与 **Plan**（如何贴回）。你返回的 `Syntax` 形状须与 Plan 一致，经 `ValidateSplice` 校验后由引擎 `Apply`。

读取外层函数签名用 `macro.EnclosingSignature(ctx, site)` / `macro.EnclosingResults(ctx, site)`，勿依赖 `*ast.FuncDecl`。

#### Decl 宏：marker 与 Expander

Decl marker 是空 struct（或带类型参数），`//macro:` 写在**类型** doc 上。使用方在源代码里**匿名嵌入** marker：

```go
//macro: derive
type Derive[T any] struct{}
```

```go
import "fmt"

type Item struct {
    mypkg.Derive[fmt.Stringer]
    Name string
}
```

常见 marker 写法：

| 模板 | 示例 | 说明 |
| ---- | ---- | ---- |
| 无参 | `type Marker struct{}` | 最简形式 |
| 类型参数 | `type Marker[T any] struct{}` | 泛型 marker，按类型参数区分展开 |
| struct tag | `` type Marker struct{ `macro:"k=v"` } `` | 可选参数通过 tag 传入 |
| 组合 | `` type Marker[T any] struct{ `macro:"opt=1"` } `` | 类型参数与 tag 可同时使用 |

tag 里的键值由你从嵌入字段的 struct tag 读取；marker 类型体内的 struct 字段仅作文档提示，引擎不会读取。

Decl Expander 使用与 Call 相同的 `Expander` 签名；pattern 通常匹配 `type $item struct { Marker $field ... }` 等形式。返回的 `Syntax` 经 `ToDecls()` 提供新 TypeSpec 与**新生成** methods；未 match 的既有 methods 自动保留。

Decl 宏只作用于 Target 的字段与方法；不要生成包级 const/var、其它类型或独立测试文件。

完整示例可参考 contrib：[derive](https://github.com/arcane-craft/go-macro-contrib/tree/master/derive)、[wirejson](https://github.com/arcane-craft/go-macro-contrib/tree/master/wirejson)。

单测可以用 `mactest.Expand`（Decl）或 `mactest.ExpandSyntax`（Call）。

### 模板化 AST（Quote）

`macro.Quote(template, map[string]Syntax)` 用接近最终 Go 的模板组装 AST，不必手写大量 `go/ast` 结构体。简单宏仍可只用手写 AST，**不强制**使用 Quote。

模板直接写 body（无 `@kind{ }`）；形状由 `Syntax.ToExpr` / `ToStmts` / `ToDecls` 等决定。洞名为 `#name`，绑定值为 `macro.Syntax`（或 `QuoteElems` 列表）。

Call 宏在 `ToStmts()` 成功后，**必须**调用 `macro.StampStmtPos(site.MacroPos(), stmts)`，以便生成代码的 `//line` 指向宏调用处。

```go
out, err := macro.Quote(`res, err := #call
if err != nil { return #zero, err }
return res, nil`, map[string]macro.Syntax{
    "call": macro.WrapExpr(callExpr),
    "zero": macro.WrapExpr(ast.NewIdent("0")),
})
if err != nil {
    return nil, err
}
stmts, _ := out.ToStmts()
macro.StampStmtPos(site.MacroPos(), stmts)
return out, nil
```

#### 调用方如何找到你的宏

使用方在宏主文件里 `import` 你的 provider 包即可。`cmd/macro expand` 只会展开**已 import** 且带 `//macro:` Expander 的包，并自动完成 link。你不需要再维护 `register/` 包。

### 纯 Expand 单测

你不必先跑完整 expand 流水线，可以直接用 `macro/mactest` 测展开逻辑：

```go
out, err := mactest.ExpandSyntax(MineExpander, "Mine", "syntax-mine", `
func Mine[T any](v T) T { panic("stub") }
func f() int { return 1 + Mine(2) }
`)
if err != nil {
    t.Fatal(err)
}
```

## 宏使用方

操作步骤以 [README 快速上手](../README.md#快速上手) 为准。本节补充文件布局与发布习惯。

### 两份源码如何配合

你会维护一对互斥的源文件：

| 文件 | 作用 |
| ---- | ---- |
| `foo.go`（宏主文件） | 你手写宏调用；带 `//go:build macro` |
| `foo_macro_gen.go`（生成文件） | expand 生成；带 `//go:build !macro` |

工具不会修改宏主文件上的 build tag。生成代码带 `//line foo.go:N`，报错行号会指回主文件。

日常 `go build`、`go test` 用生成侧即可，一般不必长期加 `-tags macro`。生成文件里不仅有展开后的函数，还包含 Decl 宏展开后的**类型定义与方法**——日常构建时，这些定义来自 `*_macro_gen.go`。

### 如何写宏调用

对已 link 的宏库，请在宏主文件里**直接调用**语法桩，例如 `try.Try(...)` 或 `inline.Inline(...)`。Decl 宏则通过 struct **匿名嵌入** marker 类型触发。

下面这类写法会导致 `expand` 失败并给出文件行号：

```go
f(try.Try)        // 把桩当参数
fn := try.Try     // 把桩赋值给变量
```

### expand 入口

在宏主文件写一行即可：

```go
//go:generate go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .
```

`expand` 会在模块根目录生成 `.gomacro/expand_runner/`（建议加入 `.gitignore`），根据宏主文件的 import 自动 link provider，再写回 `*_macro_gen.go`。

对照示例：[examples/readfile](../examples/readfile/readfile.go)。

### 发布前建议

1. 跑一遍 expand（`go generate ./...` 或 `go run .../cmd/macro@latest expand ./...`）
2. 提交更新后的 `*_macro_gen.go`
3. 执行 `go test ./...`（不加 `-tags macro`）
4. （可选）在 CI 里于 test 前跑 expand，并用 `git diff --exit-code` 检查生成文件是否忘提交

## 参考

### 官方宏库

[go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) 提供官方宏库。版本兼容与详细用法见其 [README](https://github.com/arcane-craft/go-macro-contrib/blob/master/README.md)。

| syntax-id | 模块路径 | 类型 |
| --------- | -------- | ---- |
| `inline` | `github.com/arcane-craft/go-macro-contrib/inline` | Call |
| `try` | `github.com/arcane-craft/go-macro-contrib/try` | Call |
| `derive` | `github.com/arcane-craft/go-macro-contrib/derive` | Decl |
| `wire-json` | `github.com/arcane-craft/go-macro-contrib/wirejson` | Decl |

宏主文件 import 对应包后，执行 `cmd/macro expand` 即可。

### 本地联调

**contrib** — clone 到 `../go-macro-contrib`（与 `go-macro` 同级），本地可加：

```go
replace github.com/arcane-craft/go-macro-contrib => ../go-macro-contrib
```

**go.work** — 根目录可选用 `go.work`（`use` 为 `.` 与 `./examples`）。

**测试** — 根目录 `GOWORK=off go test ./...`；`examples/` 下 `go test ./...`。

**contrib 与核心并行开发** — 可在 `go-macro-contrib/go.mod` 加：

```go
replace github.com/arcane-craft/go-macro => ../go-macro
```

上述 `replace` 一般只用于本地开发，不要提交到仓库。

### 消费其它第三方宏库

在宏主文件 import 第三方 provider，执行 `cmd/macro expand` 即可；框架会根据 `//macro:` 自动发现 Expander。第三方作者用 `init provider` 脚手架即可，无需再提供 `register` 包。
