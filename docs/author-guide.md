# 宏作者指南

## 框架契约

- Provider 包：`//macro: <syntax-id>` + `XxxExpand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error)`
- **引入方式**：宏主文件必须 `import` 该 provider；expand 工具仅对**已 import 且已在 expand 二进制中 link** 的包注册并展开
- 语法桩：包级 `panic` 函数，运行时不可调用
- `ExpandResult`：`Stmts` / `Expr` / `Exprs`（`Exprs` 少用；`syntax-try` 在 `return` 语境禁止 `Exprs`）
- `Context.EnclosingFunc`：首版必选（`*ast.FuncDecl` 或 `*ast.FuncLit`）

## 角色分工

| 角色 | 负责 | 不负责 |
|------|------|--------|
| 框架 | `macro/expandtool`、`examples/cmd/macroexpand`、`contrib/register` | — |
| 宏作者（provider） | stubs、`Expand`、`//macro:`、`register/register.go`（脚手架） | expand main、`tools/macroexpand`、手写 linked map |
| 宏使用方 | import 宏库、一行 generate | 默认无需项目内 expand 代码 |

## 调用语境（Site）

| Site | 字段 |
|------|------|
| 赋值 `:=` | `Stmts` |
| `return` | `Stmts`（或罕见 `Exprs`） |
| 语句 `Try0(...);` | `Stmts` |
| 表达式 | `Expr` |

## 纯 Expand 单测

使用 `macro/mactest`：

```go
result, err := mactest.Expand(MyExpand, "MyStub", "syntax-mine", `
func MyStub[T any](v T) T { panic("stub") }
func f() int { return 1 + MyStub(2) }
`)
```

## 使用方文件（方案 C）

- 主文件：`foo.go`，用户维护 `//go:build macro`（可与 `linux` 等合并）
- 生成侧：`foo_macro_gen.go`，工具写入 `//go:build !macro ...`
- 工具 **不** 修改主文件 build tag
- 生成代码含 `//line foo.go:N` 指向宏主文件

## init provider

```bash
go tool macro init provider mymac
```

生成最小单桩骨架与 `register/register.go`（`init` 内 `expandtool.Register`）。宏使用方展开用：

```go
//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .
```

## 发布 checklist（对外库）

1. `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand ./...`
2. 提交 `*_macro_gen.go`
3. `go test ./...`（无 `-tags macro`）
4. CI 可选：`git diff --exit-code` 防止 gen 漂移

## 官方宏库（可选）

`contrib` 子 module：`contrib/inline`、`contrib/try`。宏主文件 import 对应路径后，使用 `examples/cmd/macroexpand`（已 blank import `contrib/register`）或自建等价 cmd 即可展开。

- `syntax-inline`：`contrib/inline/`
- `syntax-try`：`contrib/try/`

## 附录：消费第三方宏库

当除 contrib 外还需其它带 `register` 子包的宏库时，使用方 MAY 复制 `examples/cmd/macroexpand` 到项目内，**仅**追加 blank import 该库的 `register` 包并仍调用 `expandtool.Main()`。无需手写 `linked` map，也无需 `tools/macroexpand`。

第三方宏作者 MUST 提供 `register` 子包（`go tool macro init provider` 已生成），**不必**维护 expand 二进制。

### Try 桩族（附录）

| 桩 | k |
|----|---|
| Try0 | 0 |
| Try | 1 |
| Try2 | 2 |
| Try3 | 3 |

内外层返回列表 **error 必须在最后**。
