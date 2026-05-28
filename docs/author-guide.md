# 宏作者指南

面向**编写宏库**的开发者。若你只是在项目里**使用**已有宏库，请直接看 [README 快速上手](../README.md#快速上手)。

## 阅读指引

| 你是谁 | 从这里开始 |
|--------|------------|
| **编写**宏库 | [用脚手架初始化](#用脚手架初始化-provider) → [框架契约](#框架契约) |
| 在工程里**使用**宏 | [README 快速上手](../README.md#快速上手)，或下文 [宏使用方](#宏使用方) |
| 官方宏库、本地联调 | [参考](#参考) |

## 角色分工

| 角色 | 做什么 |
|------|--------|
| **框架** | 提供 `cmd/macro expand`、`macro/expandtool`；按 import 自动生成 `.gomacro/expand_runner` 并完成展开 |
| **宏作者** | 写语法桩与 `XxxExpand`；在函数 doc 上标注 `//macro: <syntax-id>` |
| **宏使用方** | 在宏主文件 import 宏库，用 `go:generate` 调用 `cmd/macro expand` |

[`examples/`](../examples/) 演示宏使用方如何接入，可参考 [`examples/readfile`](../examples/readfile/readfile.go)。

## 编写宏库

### 用脚手架初始化 provider

在项目根目录执行：

```bash
go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac
```

会生成最小 provider 骨架：一个语法桩、`MacroExpand` 占位实现，以及 `expand_test.go` 模板。桩与 Expander 的 doc 里已带好 `//macro:`，你只需补全展开逻辑。

### 框架契约

#### 1. 用 `//macro:` 标注桩和 Expander

每个**语法桩**和**展开函数**的 doc 注释里都要写 `//macro: <syntax-id>`。同一 `syntax-id` 下可以有多个桩名，但只能有一个 Expander。

桩函数示例：

```go
//macro: syntax-mine
// Mine 是语法桩，运行时不要直接调用。
func Mine[T any](v T) T {
    panic("Mine is a macro stub")
}
```

Expander 示例：

```go
//macro: syntax-mine
func MineExpand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error) {
    // ...
}
```

#### 2. Expander 签名

展开函数必须是：

```go
func XxxExpand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error)
```

#### 3. 调用方如何找到你的宏

使用方在宏主文件里 `import` 你的 provider 包。`cmd/macro expand` 只会展开**已 import** 且带 `//macro:` Expander 的包，并自动完成 link。

#### 4. 语法桩约定

- 桩必须是**包级函数**（无 receiver）
- doc 里要有 `//macro:`
- 建议函数体 `panic(...)`，避免运行时被误调用
- 在宏主文件中，**已 link 注册的**语法桩只能写成直调：`pkg.Stub(...)`（或 dot-import 下的 `Stub(...)`）。不要把桩当作普通函数值传递、赋值、返回，也不要传给 `reflect.ValueOf` / `reflect.TypeOf`；违反时 `expand` 会报错

#### 5. ExpandResult：显式贴回目标（Target）

`ExpandResult` **必须**设置 `Target`（`macro.SpliceTarget`），指明要替换的 AST 范围；引擎按 `Target` 贴回，**不再**根据「填了哪个字段」隐式推断。

| `Target` | 替换范围 | 载荷 |
|----------|----------|------|
| `SpliceReplaceAssignStmt` | 整条赋值语句 | 非空 `Stmts` |
| `SpliceReplaceAssignRHS` | 仅赋值右侧含宏的那一项（**保留左侧**） | 非空 `Expr` |
| `SpliceReplaceReturnStmt` | 整条 `return` 语句 | 非空 `Stmts` |
| `SpliceReplaceReturnResults` | 仅 `return` 的返回值列表 | 非空 `Exprs` |
| `SpliceReplaceExprStmt` | 整条表达式语句 | 非空 `Stmts` |
| `SpliceReplaceCallExpr` | 仅宏 `CallExpr`（表达式槽） | 非空 `Expr` |

调用处语境（便于选 Target，**不以 `ctx.Site()` 单独决定贴回**）：

| 宏写在哪里 | 可选 `Target`（见 `ctx.LegalSpliceTargets()`） |
|------------|--------------------------------------------------|
| 赋值右侧 `:=` / `=` | `SpliceReplaceAssignRHS`、`SpliceReplaceAssignStmt` |
| `return` 里 | `SpliceReplaceReturnResults`、`SpliceReplaceReturnStmt` |
| 单独一条语句 | `SpliceReplaceExprStmt`（亦可 `SpliceReplaceCallExpr` 只换调用） |
| 表达式里 | `SpliceReplaceCallExpr` |

单测建议：展开后用 `mactest.Validate(ctx, result)` 校验 `Target` 与载荷是否与调用处一致。

#### 6. 外层函数语境

展开时可通过 `ctx.EnclosingFunc()` 获取包住本次宏调用的函数（`*ast.FuncDecl` 或 `*ast.FuncLit`），用于读取外层 `return` 签名等语境。

### 纯 Expand 单测

不必先跑完整 expand 流水线，可直接用 `macro/mactest` 测展开逻辑：

```go
result, err := mactest.Expand(MineExpand, "Mine", "syntax-mine", `
func Mine[T any](v T) T { panic("stub") }
func f() int { return 1 + Mine(2) }
`)
if err != nil {
    t.Fatal(err)
}
// 可选：在同一 snippet 上构造 ctx 后调用 mactest.Validate(ctx, result)
```

## 宏使用方

操作步骤以 [README 快速上手](../README.md#快速上手) 为准。本节补充文件布局与发布习惯。

### 两份源码如何配合

你会维护一对互斥的源文件：

| 文件 | 作用 |
|------|------|
| `foo.go`（宏主文件） | 你手写宏调用；带 `//go:build macro` |
| `foo_macro_gen.go`（生成文件） | expand 生成；带 `//go:build !macro` |

工具**不会**修改宏主文件上的 build tag。生成代码带 `//line foo.go:N`，报错行号会指回主文件。

日常 `go build`、`go test` 用生成侧即可，一般不必长期加 `-tags macro`。

### 如何写宏调用

对已 link 的宏库，在宏主文件里应**直接调用**语法桩，例如 `try.Try(...)` 或 `inline.Inline(...)`。不要写 `f(try.Try)`、`fn := try.Try` 等把桩当作函数值的用法，`cmd/macro expand` 会失败并给出文件行号。

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

[go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) 提供官方宏库：

| syntax-id | 模块路径 |
|-----------|----------|
| `syntax-inline` | `github.com/arcane-craft/go-macro-contrib/inline` |
| `syntax-try` | `github.com/arcane-craft/go-macro-contrib/try` |

宏主文件 import 对应包后，执行 `cmd/macro expand` 即可。版本兼容与本地联调见 [go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) README。

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

一般不提交上述 `replace`。

### 消费其它第三方宏库

在宏主文件 import 第三方 provider，执行 `cmd/macro expand` 即可；框架会根据 `//macro:` 自动发现 Expander。第三方作者用 `init provider` 脚手架即可，无需再提供 `register` 包。
