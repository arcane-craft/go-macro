## MODIFIED Requirements

### Requirement: 轻薄 AST 辅助（首版）

`macro` 包 MUST 提供最小辅助：`CallContext`、`CallExpandResult`、`DeclContext`、`DeclExpandResult`、`SpliceTarget`、`ValidateCallExpandResult`、`ValidateDeclExpandResult`、`TempIdent`、定位/错误辅助。

框架 MUST 在 **`macro/quote` 可选子包** 提供模板化 AST 组装（见 `macro-quote` 规范）。provider 实现 Call 或 Decl Expander时：

- MUST NOT 被**强制** import `macro/quote` 或任何独立 astbuilder 包；
- MAY import `macro/quote` 以用 `@kind{ }` / `#` 模板构造展开结果。

#### Scenario: 首版无 astbuilder 依赖

- **WHEN** provider 实现 Call 或 Decl Expander 且选择不 import `macro/quote`
- **THEN** MUST 可仅依赖 `macro` 根包公开 API 完成实现

#### Scenario: 可选 quote 子包

- **WHEN** provider import `github.com/arcane-craft/go-macro/macro/quote` 并使用 `quote.Stmts` 等 API
- **THEN** MUST NOT 要求同时 import 其它 astbuilder 包
