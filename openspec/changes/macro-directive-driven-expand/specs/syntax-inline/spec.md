## MODIFIED Requirements

### Requirement: Inline 语法桩

`inline` 包 MUST 提供单桩 `Inline[T any](v T) T`。`Inline` 函数的 doc MUST 含 `//macro: syntax-inline`。函数体 SHOULD panic，标明宏桩不可直接调用。

#### Scenario: 宏源文件类型检查

- **WHEN** 用户在 macro 主文件中编写 `x := Inline(42)`
- **THEN** `Inline` 桩 MUST 使该表达式通过类型检查（与展开后 `x := 42` 一致）

### Requirement: InlineExpand 表达式展开

`inline` 包 MUST 提供 `InlineExpand`；其 doc MUST 含 `//macro: syntax-inline`。在 `SiteExpr` 语境 MUST 返回 `ExpandResult{Expr: <实参表达式>}`。

#### Scenario: SiteExpr 仅替换 CallExpr

- **WHEN** 展开引擎对 `Inline(f())` 调用且 `ctx.Site()` 为 `SiteExpr`
- **THEN** `InlineExpand` MUST 返回非空 `Expr`，且 MUST NOT 返回 `Stmts`
- **THEN** 引擎 MUST 仅用该 `Expr` 替换原 `CallExpr`

#### Scenario: 非表达式语境拒绝

- **WHEN** 调用出现在 `SiteAssign`、`SiteReturn` 或 `SiteStmt`
- **THEN** `InlineExpand` MUST 返回错误，说明 `Inline` 仅用于表达式位置

### Requirement: 可选官方宏库与引入方式

`inline` 包 MUST 在 `go-macro-contrib` 仓库内发布。使用方 MUST 在宏主文件中 import 该路径，且 MUST 通过 `cmd/macro expand` 生成 link 后方可展开 `Inline(...)`。

#### Scenario: 未 import 时不展开

- **WHEN** 宏主文件调用 `Inline(...)` 但未 import `github.com/arcane-craft/go-macro-contrib/inline`
- **THEN** 展开管线 MUST NOT 注册 `syntax-inline`

#### Scenario: import 后 expand 生成 link

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro-contrib/inline` 并执行 `cmd/macro expand`
- **THEN** MUST 注册并展开 `syntax-inline`
