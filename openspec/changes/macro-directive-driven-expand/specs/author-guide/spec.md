## MODIFIED Requirements

### Requirement: 编写宏库保留 normative 契约要点

`编写宏库` MUST 包含且保持与 `macro-core`、`macro-codegen`、`macro-directive` 一致的下列语义：

- 每个语法桩函数与每个 Expander 函数的 doc MUST 含 `//macro: <syntax-id>`；同一 syntax-id 下多桩共享一个 Expander
- `Expand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error)` 签名约定
- 宏主文件 MUST import provider；expand 仅对已 import 且由 `cmd/macro expand` 生成 link 的 provider 展开
- 语法桩 RECOMMENDED 包级 `panic`，运行时不可调用
- MUST NOT 要求宏作者维护 `register/` 包
- `ExpandResult` 的 `Stmts` / `Expr` / `Exprs` 及宏出现位置与返回字段的对应关系
- `Context.EnclosingFunc()`
- `init provider` 入口为 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>`
- 纯 Expand 单测使用 `macro/mactest`

#### Scenario: 契约与 macro-directive 对齐

- **WHEN** 读者对照 `macro-directive` 与 `macro-core` 要求
- **THEN** author-guide `编写宏库` MUST 说明 per-function `//macro:` 且 MUST NOT 将 `register/` 列为作者职责

### Requirement: 宏使用方节保留 codegen 要点

`宏使用方` MUST 包含且保持与 `macro-codegen` 一致的下列语义：

- 主文件 `//go:build macro`、生成侧 `//go:build !macro`、工具不修改主文件 build tag、生成代码含 `//line`
- expand 入口 RECOMMENDED 为 `go run github.com/arcane-craft/go-macro/cmd/macro@latest expand`（或 generate 一行）；MUST NOT 要求自建 `cmd/macroexpand` 或 blank import `contrib/register`
- 发布前：expand、`go test`（无 `-tags macro`）、提交 `*_macro_gen.go`、可选 CI diff

#### Scenario: expand 入口说明更新

- **WHEN** 读者阅读 `宏使用方` 中 expand 入口说明
- **THEN** MUST 理解默认使用 `cmd/macro expand`，且 MUST NOT 将手写 `register` 列为使用方步骤

## MODIFIED Requirements

### Requirement: 角色分工（文档同步）

`角色分工` 表中宏作者职责 MUST 列出 per-function `//macro:`、语法桩、`Expand`；MUST NOT 列出 `register/`。框架职责 MUST 包含 `cmd/macro expand` 与 link 生成。

#### Scenario: 角色表无 register

- **WHEN** 读者阅读 `角色分工`
- **THEN** MUST NOT 将 `register/` 作为宏作者必做项
