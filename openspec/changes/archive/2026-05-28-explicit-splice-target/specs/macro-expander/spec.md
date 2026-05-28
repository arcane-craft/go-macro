## MODIFIED Requirements

### Requirement: ExpandResult 贴回（splice）

引擎 MUST 按 `ExpandResult.Target` 与对应载荷贴回 AST。引擎 MUST 在贴回前调用与 `macro.ValidateExpandResult` 等价的校验（可直接调用 `macro` 包函数）。引擎 MUST NOT 依据 `ctx.Site()` 选择贴回字段；`Site()` 仅用于 provider 语义与错误提示。

| `Target` | 行为 |
|----------|------|
| `SpliceReplaceAssignStmt` | 替换整条 `AssignStmt` 为 `Stmts` |
| `SpliceReplaceAssignRHS` | 在 enclosing `AssignStmt.Rhs` 中定位宏 `CallExpr` 所在槽位，仅用 `Expr` 替换该槽位；`Lhs` MUST 不变 |
| `SpliceReplaceReturnStmt` | 替换整条 `ReturnStmt` 为 `Stmts` |
| `SpliceReplaceReturnResults` | 仅将 `ReturnStmt.Results` 设为 `Exprs` |
| `SpliceReplaceExprStmt` | 替换整条 `ExprStmt` 为 `Stmts` |
| `SpliceReplaceCallExpr` | 在表达式槽中用 `Expr` 替换宏 `CallExpr`（含 `BinaryExpr`、`CallExpr` 参数、`CompositeLit` 等父节点，与现有 `replaceCallExpr` 行为一致） |

若 `Target` 不在当前调用处的结构合法集合内，或载荷与 `Target` 不匹配，MUST 报错。错误信息 MUST 含文件名与行号，并 SHOULD 列出 `LegalSpliceTargets()` 允许的目标名称。引擎 MUST NOT 对任何 `syntax-id` 硬编码 splice 分支。

#### Scenario: assign 仅换 RHS

- **WHEN** 某 `Expander` 对 `x := Macro()` 返回 `Target: SpliceReplaceAssignRHS` 与 `Expr: e`
- **THEN** 引擎 MUST 保留 `x` 在 `Lhs` 中，且 `Rhs` 中含宏的项 MUST 变为 `e`

#### Scenario: return 语境替换整条 return

- **WHEN** 某 `Expander`（如 contrib 仓 `TryExpand`）对 `return Macro(...)` 返回 `Target: SpliceReplaceReturnStmt` 与非空 `Stmts`
- **THEN** 引擎 MUST 用 `Stmts` 替换整条 `return` 语句

#### Scenario: return 语境仅换 Results

- **WHEN** 某 `Expander` 对 `return Macro(...)` 返回 `Target: SpliceReplaceReturnResults` 与非空 `Exprs`
- **THEN** 引擎 MUST 仅替换 `ReturnStmt.Results`，且 MUST NOT 替换整条 `ReturnStmt` 为语句块

#### Scenario: 表达式宏替换 CallExpr

- **WHEN** 某 `Expander` 对表达式槽宏返回 `Target: SpliceReplaceCallExpr` 与 `Expr: x`
- **THEN** 引擎 MUST 仅用 `x` 替换原宏 `CallExpr`

#### Scenario: Target 与锚点不一致

- **WHEN** 宏位于 `return Macro()` 但 `Expander` 返回 `Target: SpliceReplaceCallExpr`
- **THEN** expand MUST 失败，且 MUST NOT 写回部分展开结果

#### Scenario: 载荷与 Target 不匹配

- **WHEN** `Expander` 返回 `Target: SpliceReplaceAssignRHS` 且 `Expr` 为 nil
- **THEN** expand MUST 失败

## ADDED Requirements

### Requirement: 结构锚点解析

展开引擎 MUST 能从 `*ast.CallExpr` 解析宏调用的结构锚点，并据此计算合法 `SpliceTarget` 集合（与 `macro.Context.LegalSpliceTargets()` 一致）：

- 若 `call` 等于某 `AssignStmt.Rhs` 元素（剥除 `ParenExpr` 后比较），合法目标 MUST 包含 `SpliceReplaceAssignRHS` 与 `SpliceReplaceAssignStmt`
- 若 `call` 等于某 `ReturnStmt.Results` 元素，合法目标 MUST 包含 `SpliceReplaceReturnResults` 与 `SpliceReplaceReturnStmt`
- 若 `call` 等于某 `ExprStmt.X`，合法目标 MUST 包含 `SpliceReplaceExprStmt`
- 否则，合法目标 MUST 为 `SpliceReplaceCallExpr` 仅此一项

#### Scenario: 嵌套表达式仍为 CallExpr 目标

- **WHEN** 源码为 `f(1 + Macro(2))`
- **THEN** 合法目标 MUST 为 `SpliceReplaceCallExpr` only

#### Scenario: assign 双目标

- **WHEN** 源码为 `a, b := Macro()`
- **THEN** 合法目标 MUST 包含 `SpliceReplaceAssignRHS` 与 `SpliceReplaceAssignStmt`

### Requirement: ApplyExpandResult 不依赖 CallSiteKind

`ApplyExpandResult`（或等价 splice 入口）MUST 仅接受 `file`、`call`、`ExpandResult`；MUST NOT 将 `CallSiteKind` 作为贴回分支条件。

#### Scenario: Site 与 Target 解耦

- **WHEN** `ctx.Site()` 为 `SiteAssign` 但 `ExpandResult.Target` 为 `SpliceReplaceCallExpr` 且宏在 assign RHS
- **THEN** splice MUST 失败（即使 `Site` 为 assign 语境）
