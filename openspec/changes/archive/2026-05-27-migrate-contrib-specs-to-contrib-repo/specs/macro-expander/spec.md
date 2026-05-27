## ADDED Requirements

### Requirement: Provider 语义以外置规范为准

`TryExpand`、`InlineExpand` 的载荷校验、Site 禁止规则、展开语句形态等 provider 级语义 MUST 以 `go-macro-contrib` 仓库内 `syntax-try`、`syntax-inline` OpenSpec 为准。本规范仅定义展开引擎的识别、分发、`ExpandResult` 贴回与 link/import 边界。

#### Scenario: 修改 Try 展开语义

- **WHEN** 维护者需变更 `return Try` 的错误路径语句形态
- **THEN** MUST 修改 contrib 仓 `syntax-try` spec 及 `try` 实现，而非在本 spec 中新增 Try 专用引擎分支

## MODIFIED Requirements

### Requirement: ExpandResult 贴回（splice）

引擎 MUST 按 `ctx.Site()` 与 `ExpandResult` 字段贴回 AST：

| Site | 字段 | 行为 |
|------|------|------|
| `SiteAssign` | `Stmts` | 替换整条 `AssignStmt` |
| `SiteReturn` | `Stmts` | 替换整条 `ReturnStmt` |
| `SiteStmt` | `Stmts` | 替换 `ExprStmt` |
| `SiteExpr` | `Expr` | 仅替换 `CallExpr` |
| `SiteReturn` | `Exprs` | 仅替换 `ReturnStmt` 的 Results 列表（首版保留；contrib 仓 `syntax-try` 规范规定 `TryExpand` 在 `SiteReturn` 不得使用 `Exprs`） |

若字段与 Site 不匹配，MUST 报错。引擎 MUST NOT 对任何 `syntax-id` 硬编码分支；贴回规则仅依赖上表。

#### Scenario: return 语境使用 Stmts（引擎行为）

- **WHEN** 某 `Expander`（如 contrib 仓的 `TryExpand`）对 `SiteReturn` 返回非空 `Stmts`
- **THEN** 引擎 MUST 用 `Stmts` 替换整条 `return` 语句

#### Scenario: 表达式宏替换 CallExpr

- **WHEN** 某 `Expander` 在 `SiteExpr` 返回 `ExpandResult{Expr: x}`
- **THEN** 引擎 MUST 仅用 `x` 替换原宏 `CallExpr`
