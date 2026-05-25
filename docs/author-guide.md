# 宏作者指南

## 框架契约

- Provider 包：`//macro: <syntax-id>` + `XxxExpand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error)`
- **引入方式**：宏主文件必须 `import` 该 provider；`go tool macro expand` 仅对**已 import** 的包注册并展开（含本仓库官方宏库 `inline`、`try`）
- 语法桩：包级 `panic` 函数，运行时不可调用
- `ExpandResult`：`Stmts` / `Expr` / `Exprs`（`Exprs` 少用；`syntax-try` 在 `return` 语境禁止 `Exprs`）
- `Context.EnclosingFunc`：首版必选（`*ast.FuncDecl` 或 `*ast.FuncLit`）

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

生成最小单桩骨架，详见包内 README。

## 发布 checklist（对外库）

1. `go tool macro expand ./...`
2. 提交 `*_macro_gen.go`
3. `go test ./...`（无 `-tags macro`）
4. CI 可选：`git diff --exit-code` 防止 gen 漂移

## 官方宏库（可选）

本模块内维护，当作普通依赖使用：在宏主文件中 import 后即可 `go tool macro expand`，无需在 CLI 里单独登记。

- `syntax-inline`：`inline/`，表达式宏
- `syntax-try`：`try/`，多桩 `Try0`/`Try`/`Try2`/…

自研 provider（`init provider` 生成）须在同一进程内提供 `Expand` 函数指针（例如自定义 `tools` 入口调用 `expander.ExpandPackages(..., []Provider{...})`）；官方库由 expander 在检测到 import 时自动衔接。

### Try 桩族（附录）

| 桩 | k |
|----|---|
| Try0 | 0 |
| Try | 1 |
| Try2 | 2 |
| Try3 | 3 |

内外层返回列表 **error 必须在最后**。
