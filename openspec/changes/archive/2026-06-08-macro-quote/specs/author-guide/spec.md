## ADDED Requirements

### Requirement: 编写宏库 Quote 子节

`docs/author-guide.md` 的 `编写宏库` MUST 含 `###` 或等价子节「模板化 AST（Quote）」或并入「框架契约」，说明：

- `macro/quote` 为**可选**子包；与手写 AST 的关系
- 四种 kind 与 `Expr`/`Exprs`/`Stmts`/`Decls` 对应；typed API 直接写 body，不必再包 `@kind{ }`；`quote.Quote` 仍需 `@expr{ }` 等显式根
- `#name` 填洞；绑定类型（string ident、`ast.Expr`、`[]ast.Stmt`、嵌套 Quote 等）
- 产出与 `CallExpandResult` / `DeclExpandResult` 及 `SpliceTarget` 的对应表
- Expander 内在 `quote.Stmts` 等之后 MUST 调用 `macro.StampStmtPos(ctx.MacroPos(), ...)`（Call）或等价行号策略
- 模板注释会保留至产出 AST 的说明（一句即可）

MUST NOT 要求所有 provider 使用 Quote。

#### Scenario: 作者查 Quote 根 kind

- **WHEN** 宏作者阅读 `编写宏库` Quote 子节
- **THEN** MUST 能找到四种 kind 与 `Expr`/`Exprs`/`Stmts`/`Decls` 的对应，以及 typed API 可直接写 body 的说明

#### Scenario: 作者查贴回衔接

- **WHEN** 宏作者阅读 Quote 子节
- **THEN** MUST 能找到 `@stmts` → `CallExpandResult.Stmts` + 显式 `Target` 的示例或表格
