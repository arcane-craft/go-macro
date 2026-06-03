# 宏作者指南

面向**编写宏库**的开发者。若你只是在项目里**使用**已有宏库，请直接看 [README 快速上手](../README.md#快速上手)。

## 阅读指引


| 你是谁         | 从这里开始                                            |
| ----------- | ------------------------------------------------ |
| **编写**宏库    | [用脚手架初始化](#用脚手架初始化-provider) → [框架契约](#框架契约)     |
| 在工程里**使用**宏 | [README 快速上手](../README.md#快速上手)，或 [宏使用方](#宏使用方) |
| 官方宏库、本地联调   | [参考](#参考)                                        |


## 角色分工


| 角色       | 做什么                                                                                   |
| -------- | ------------------------------------------------------------------------------------- |
| **框架**   | 提供 `cmd/macro expand`、`macro/expandtool`；按 import 自动生成 `.gomacro/expand_runner` 并完成展开 |
| **宏作者**  | 写 Call 语法桩 / Decl marker 与对应 `XxxExpand`；在桩或 marker 类型 doc 上标注 `//macro: <syntax-id>` |
| **宏使用方** | 在宏主文件 import 宏库，用 `go:generate` 调用 `cmd/macro expand`                                 |


`[examples/](../examples/)` 演示宏使用方如何接入，可参考 `[examples/readfile](../examples/readfile/readfile.go)`。

## 编写宏库

### 用脚手架初始化 provider

在项目根目录执行：

```bash
go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac
```

脚手架会生成最小 provider 骨架：一个语法桩、`MacroExpand` 占位实现，以及 `expand_test.go` 模板。桩与 Expander 的 doc 里已带好 `//macro:`，你只需补全展开逻辑。

### 框架契约

框架支持两条宏轨道。同一 provider 包可以注册**多个** `syntax-id`；`expandtool.RegisterCall` / `RegisterDecl` 按 syntax-id 分别 link。展开顺序为：**先 Decl，后 Call**。


| 轨道            | 触发方式                      | Expander 签名                                                           |
| ------------- | ------------------------- | --------------------------------------------------------------------- |
| **Call（过程宏）** | 函数体内 `Stub(...)`          | `func(ctx CallContext, call *ast.CallExpr) (CallExpandResult, error)` |
| **Decl（声明宏）** | struct **匿名嵌入** marker 类型 | `func(ctx DeclContext, site DeclSite) (DeclExpandResult, error)`      |


#### 标注 `//macro:`

你需要在以下位置的 doc 里写 `//macro: <syntax-id>`：

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
func MineExpand(ctx macro.CallContext, call *ast.CallExpr) (macro.CallExpandResult, error) {
    // ...
}
```

#### Call 宏：语法桩约定

编写 Call 桩时，请遵循这些约定：

- 桩是**包级函数**（无 receiver）
- doc 里写 `//macro: <syntax-id>`
- 函数体建议 `panic(...)`，避免运行时被误调用
- 使用过程宏，请**直接调用**语法桩：`pkg.Stub(...)`（或 dot-import 下的 `Stub(...)`）
- 桩只能作为调用表达式出现；若你把桩当作函数值传递、赋值、返回，或传给 `reflect.ValueOf` / `reflect.TypeOf`，`expand` 会报错

#### Call 宏：设置 `Target` 并返回结果

展开 Call 宏时，你需要在 `CallExpandResult` 里设置 `Target`（`macro.SpliceTarget`），告诉引擎要替换哪段 AST。引擎按 `Target` 贴回，不会根据你填了 `Stmts` 还是 `Expr` 来猜测替换范围。

**先按调用位置选 Target**（完整列表见 `ctx.LegalSpliceTargets()`）：


| 宏写在哪里           | 常用 `Target`                                                           |
| --------------- | --------------------------------------------------------------------- |
| 赋值右侧 `:=` / `=` | `SpliceReplaceAssignRHS`（只换右侧，保留左侧）或 `SpliceReplaceAssignStmt`（换整条赋值） |
| `return` 里      | `SpliceReplaceReturnResults` 或 `SpliceReplaceReturnStmt`              |
| 单独一条语句          | `SpliceReplaceExprStmt`                                               |
| 表达式里            | `SpliceReplaceCallExpr`                                               |


**各 Target 对应的替换范围与载荷：**


| `Target`                     | 替换范围              | 你需要提供的载荷   |
| ---------------------------- | ----------------- | ---------- |
| `SpliceReplaceAssignStmt`    | 整条赋值语句            | 非空 `Stmts` |
| `SpliceReplaceAssignRHS`     | 仅赋值右侧含宏的那一项       | 非空 `Expr`  |
| `SpliceReplaceReturnStmt`    | 整条 `return` 语句    | 非空 `Stmts` |
| `SpliceReplaceReturnResults` | 仅 `return` 的返回值列表 | 非空 `Exprs` |
| `SpliceReplaceExprStmt`      | 整条表达式语句           | 非空 `Stmts` |
| `SpliceReplaceCallExpr`      | 仅宏 `CallExpr`     | 非空 `Expr`  |


单测时，展开后可以调用 `mactest.ValidateCall(ctx, result)`，检查 `Target` 与载荷是否与调用处一致。

#### Call 宏：读取外层函数语境

展开时，你可以通过 `ctx.EnclosingFunc()` 获取包住本次宏调用的函数（`*ast.FuncDecl` 或 `*ast.FuncLit`），用来读取外层 `return` 签名等信息。

#### Decl 宏：marker 与 Expander

Decl marker 是空 struct（或带类型参数），`//macro:` 写在**类型** doc 上：

```go
//macro: derive-stringer
type DeriveStringer struct{}
```

使用方在源代码里匿名嵌入 marker：

```go
type Item struct {
    mypkg.DeriveStringer
    Name string
}
```

Decl Expander 签名：

```go
func DeriveStringerExpand(ctx macro.DeclContext, site macro.DeclSite) (macro.DeclExpandResult, error)
```

展开成功时，你需要返回**全量** `Fields`（不含嵌入桩）与**全量** `Methods`（receiver 为 Target 的全部方法，含生成与未改动的）。可以用 `ctx.TargetMethods()` 复制现有方法后再修改。

可选参数通过嵌入字段 ``macro:"k=v"`` 传递；请从 struct tag 读取，而不要读取 marker 类型内的 struct 字段。

单测可以用 `mactest.ExpandDecl` / `mactest.ValidateDecl`。

#### 调用方如何找到你的宏

使用方在宏主文件里 `import` 你的 provider 包即可。`cmd/macro expand` 只会展开**已 import** 且带 `//macro:` Expander 的包，并自动完成 link。你不需要再维护 `register/` 包。

### 纯 Expand 单测

你不必先跑完整 expand 流水线，可以直接用 `macro/mactest` 测展开逻辑：

```go
result, err := mactest.ExpandCall(MineExpand, "Mine", "syntax-mine", `
func Mine[T any](v T) T { panic("stub") }
func f() int { return 1 + Mine(2) }
`)
if err != nil {
    t.Fatal(err)
}
// 可选：在同一 snippet 上构造 ctx 后调用 mactest.ValidateCall(ctx, result)
```

## 宏使用方

操作步骤以 [README 快速上手](../README.md#快速上手) 为准。本节补充文件布局与发布习惯。

### 两份源码如何配合

你会维护一对互斥的源文件：


| 文件                       | 作用                              |
| ------------------------ | ------------------------------- |
| `foo.go`（宏主文件）           | 你手写宏调用；带 `//go:build macro`     |
| `foo_macro_gen.go`（生成文件） | expand 生成；带 `//go:build !macro` |


工具不会修改宏主文件上的 build tag。生成代码带 `//line foo.go:N`，报错行号会指回主文件。

日常 `go build`、`go test` 用生成侧即可，一般不必长期加 `-tags macro`。

### 如何写宏调用

对已 link 的宏库，请在宏主文件里**直接调用**语法桩，例如 `try.Try(...)` 或 `inline.Inline(...)`。

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

[go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) 提供官方宏库：


| syntax-id         | 模块路径                                                                                    | 类型   |
| ----------------- | --------------------------------------------------------------------------------------- | ---- |
| `syntax-inline`   | `github.com/arcane-craft/go-macro-contrib/inline`                                       | Call |
| `syntax-try`      | `github.com/arcane-craft/go-macro-contrib/try`                                          | Call |
| `derive-stringer` | [go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) · `derivestringer` | Decl |
| `wire-json`       | [go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) · `wirejson`       | Decl |


宏主文件 import 对应包后，执行 `cmd/macro expand` 即可。contrib 中 `TryExpand` / `InlineExpand` 已使用 `CallContext` / `CallExpandResult`。

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