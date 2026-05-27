# 宏作者指南

面向**编写宏库**的开发者：如何搭骨架、实现展开、写单测。若你只是在项目里**使用**已有宏库，请直接看 [README 快速上手](../README.md#快速上手)。

## 阅读指引

| 你是谁 | 从这里开始 |
|--------|------------|
| **编写**宏库 | [用脚手架初始化](#用脚手架初始化-provider) → [框架契约](#框架契约) |
| 在工程里**使用**宏 | [README 快速上手](../README.md#快速上手)，或 [宏使用方](#宏使用方) |
| 官方宏库、本地联调 | [参考](#参考) |

## 角色分工

| 角色 | 职责 |
|------|------|
| 框架 | 提供 `macro/expandtool`（Register / Run / Main） |
| 宏作者 | 语法桩、`Expand`、`//macro:`、`register/`（脚手架会生成） |
| 宏使用方 | import 宏库；在项目内自建 `cmd/macroexpand` 触发展开 |

本仓库 [`examples`](../examples/) 演示宏使用方如何接入，可参考 [`examples/cmd/macroexpand`](../examples/cmd/macroexpand/main.go)。

## 编写宏库

### 用脚手架初始化 provider

在项目根目录执行：

```bash
go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac
```

CLI 会生成最小单桩骨架和 `register/register.go`（在 `init` 里调用 `expandtool.Register`）。你接着实现 `Expand` 与语法桩即可。

若他人要使用你的宏库，请让对方在其项目内自建 `cmd/macroexpand`（见 [README 快速上手](../README.md#快速上手)），不要长期依赖本仓库的 `examples/cmd/macroexpand` 路径。

### 框架契约

1. **语法标识**：在 provider 包注释里写 `//macro: <syntax-id>`（例如 `syntax-mine`）。
2. **展开函数**：实现  
   `XxxExpand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error)`。
3. **引入与注册**：调用方源码须 `import` 你的 provider；expand 只会处理**已 import、且 expand 二进制已通过 `register` 登记**的宏库。
4. **语法桩**：提供包级 `panic` 占位函数，供类型检查与 IDE 使用；正常运行时不会走到这些桩。
5. **展开结果** `ExpandResult`：
   - `Stmts`：一条或多条语句；
   - `Expr`：单个表达式；
   - `Exprs`：多个表达式（少用；在 `return` 位置尤其要谨慎）。
6. **外层函数**：展开时可通过 `ctx.EnclosingFunc()` 拿到包住本次调用的函数（`*ast.FuncDecl` 或 `*ast.FuncLit`）。若生成代码需要参考外层 `return` 签名，从这里读取即可。

### 宏出现的位置与返回字段

| 宏写在哪里 | 通常返回 |
|------------|----------|
| 赋值右侧 `:=` | `Stmts` |
| `return` 里 | `Stmts`（少数情况用 `Exprs`） |
| 单独一条语句 | `Stmts` |
| 表达式里 | `Expr` |

### 纯 Expand 单测

你不必先接完整 expand 流水线，可直接用 `macro/mactest` 验证展开逻辑：

```go
result, err := mactest.Expand(MyExpand, "MyStub", "syntax-mine", `
func MyStub[T any](v T) T { panic("stub") }
func f() int { return 1 + MyStub(2) }
`)
```

## 宏使用方

操作步骤以 [README 快速上手](../README.md#快速上手) 为准；下面是文件布局与入口要点。

### build tag 与生成文件

你会维护两份互斥的源码，这样日常编译走展开结果，编辑时仍可写带宏的「主文件」：

- **宏主文件** `foo.go`：你手写宏调用，并加上 `//go:build macro`（可与 `linux` 等 tag 写在一起）。
- **生成文件** `foo_macro_gen.go`：expand 工具生成，带 `//go:build !macro`；工具不会改你主文件上的 build tag。
- 生成代码带 `//line foo.go:N`，报错行号会指回主文件。

因此日常 `go build` / `go test` 用生成侧即可，一般不必长期加 `-tags macro`。

### expand 入口

使用宏的项目在 `cmd/macroexpand/main.go` 里 blank import 各宏库的 `register`，并调用 `expandtool.Main()`。

对照示例：[examples/cmd/macroexpand](../examples/cmd/macroexpand/main.go)、[examples/readfile](../examples/readfile/readfile.go)（含 `go:generate`）。

### 发布前建议

库会被他人 `import`，且仓库里已有宏调用与 `*_macro_gen.go` 时，建议：

1. 跑一遍 expand（例如 `go run ./cmd/macroexpand ./...`）。
2. 提交更新后的 `*_macro_gen.go`。
3. 执行 `go test ./...`（不加 `-tags macro`）。
4. （可选）CI 里用 `git diff --exit-code` 检查生成文件是否忘提交。

## 参考

### 官方宏库

[go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) 提供例如：

| syntax-id | 模块路径 |
|-----------|----------|
| `syntax-inline` | `github.com/arcane-craft/go-macro-contrib/inline` |
| `syntax-try` | `github.com/arcane-craft/go-macro-contrib/try` |

在源码里 import 对应包，并在 `cmd/macroexpand` 里 blank import `go-macro-contrib/register` 后即可展开。

### 本地联调

- **contrib**：将 `go-macro-contrib` clone 到 `../go-macro-contrib`（与 `go-macro` 同级）。并行开发时可在本地加  
  `replace github.com/arcane-craft/go-macro-contrib => ../go-macro-contrib`（通常不必提交）。
- **go.work**：根目录可选用 `go.work`（`use` 为 `.` 与 `./examples`），不配置也能正常用框架。
- **测试**：根目录 `GOWORK=off go test ./...`；`examples/` 下 `go test ./...`。
- **联调 contrib 与核心**：可在 `go-macro-contrib/go.mod` 加  
  `replace github.com/arcane-craft/go-macro => ../go-macro`（一般不提交）。

### 消费其它第三方宏库

除 contrib 外，你还可以在自建的 `cmd/macroexpand` 里多写一行 blank import 其它宏库的 `register`，仍调用 `expandtool.Main()`。第三方宏库作者用 `init provider` 脚手架即可带上 `register` 包。
