# syntax-inline Specification

## Purpose
TBD - created by archiving change go-macro-extension. Update Purpose after archive.
## Requirements
### Requirement: Inline 语法桩

`inline` 包 MUST 提供单桩 `Inline[T any](v T) T`，函数体 MUST panic，并标注为宏桩不可直接调用。

#### Scenario: 宏源文件类型检查

- **WHEN** 用户在 macro 主文件中编写 `x := Inline(42)`
- **THEN** `Inline` 桩 MUST 使该表达式通过类型检查（与展开后 `x := 42` 一致）

### Requirement: InlineExpand 表达式展开

`inline` 包 MUST 提供 `//macro: syntax-inline` 标注的 `InlineExpand`，在 `SiteExpr` 语境 MUST 返回 `ExpandResult{Expr: <实参表达式>}`，即展开为去掉 `Inline(...)` 包装的内层表达式。

#### Scenario: SiteExpr 仅替换 CallExpr

- **WHEN** 展开引擎对 `Inline(f())` 调用且 `ctx.Site()` 为 `SiteExpr`
- **THEN** `InlineExpand` MUST 返回非空 `Expr`，且 MUST NOT 返回 `Stmts`
- **THEN** 引擎 MUST 仅用该 `Expr` 替换原 `CallExpr`

#### Scenario: 非表达式语境拒绝

- **WHEN** 调用出现在 `SiteAssign`、`SiteReturn` 或 `SiteStmt`
- **THEN** `InlineExpand` MUST 返回错误，说明 `Inline` 仅用于表达式位置

### Requirement: 可选官方宏库与引入方式

`inline` 包 MUST 作为**官方宏库**发布：框架不默认启用；使用方 MUST 在宏主文件中 import `github.com/arcane-craft/go-macro/inline` 后，`go tool macro expand` 方可展开 `Inline(...)`。

#### Scenario: 未 import 时不展开

- **WHEN** 宏主文件调用 `Inline(...)` 但未 import `inline` 包
- **THEN** 展开管线 MUST NOT 注册 `syntax-inline`

### Requirement: 与框架边界

`inline` 包 MUST NOT 依赖 `macro` 包内的 error 载荷、k 校验或 Try 专用 API；仅使用通用 `Context` 与 `ExpandResult`。

#### Scenario: 独立 syntax-id

- **WHEN** 注册表构建完成
- **THEN** `Inline` 桩 MUST 映射到 `syntax-inline` 的 `InlineExpand`，且 MUST NOT 与 `syntax-try` 共用展开器

### Requirement: mactest 单测

`InlineExpand` MUST 具备不依赖 `//go:build macro` 的 `mactest` 单测。

#### Scenario: 纯 Expand 测试

- **WHEN** 执行 `go test ./inline/...`
- **THEN** 测试 MUST 无需全链路 `macro expand` 即可通过

