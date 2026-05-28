## MODIFIED Requirements

### Requirement: ExpandResult 与 Expander 签名

`macro` 包 MUST 定义：

- `SpliceTarget` 枚举，至少包含：`SpliceReplaceAssignStmt`、`SpliceReplaceAssignRHS`、`SpliceReplaceReturnStmt`、`SpliceReplaceReturnResults`、`SpliceReplaceExprStmt`、`SpliceReplaceCallExpr`
- `ExpandResult` 含 **必填** 字段 `Target SpliceTarget`，以及载荷字段 `Stmts []ast.Stmt`、`Exprs []ast.Expr`、`Expr ast.Expr`；贴回语义 **仅** 由 `Target` 与对应载荷决定，MUST NOT 再依据「哪个字段非空」隐式推断贴回方式
- `Expander func(ctx Context, call *ast.CallExpr) (ExpandResult, error)`

各 `Target` 与载荷的对应关系 MUST 为：

| `Target` | 载荷 |
|----------|------|
| `SpliceReplaceAssignStmt` | 非空 `Stmts` |
| `SpliceReplaceAssignRHS` | 非空 `Expr` |
| `SpliceReplaceReturnStmt` | 非空 `Stmts` |
| `SpliceReplaceReturnResults` | 非空 `Exprs` |
| `SpliceReplaceExprStmt` | 非空 `Stmts` |
| `SpliceReplaceCallExpr` | 非空 `Expr` |

每种 `Target` 下，其它载荷字段 MUST 为空（`Stmts`/`Exprs` 长度为 0，`Expr` 为 nil）。

Expander 函数 MUST 在**该函数**的 doc 中含 `//macro: <syntax-id>` 并符合 `Expander` 签名。语法桩 MUST 在**各桩函数**的 doc 中含 `//macro: <syntax-id>`。

`macro` 包 MUST 提供 `ValidateExpandResult(ctx Context, result ExpandResult) error`，校验 `Target`、载荷形状，以及 `Target` 是否属于 `ctx.LegalSpliceTargets()`。

#### Scenario: 表达式宏显式 Target

- **WHEN** 某宏在表达式槽展开并返回 `ExpandResult{Target: SpliceReplaceCallExpr, Expr: e}`
- **THEN** `ValidateExpandResult` MUST 通过，且引擎 MUST 仅用 `e` 替换原宏 `CallExpr`

#### Scenario: 赋值仅换 RHS

- **WHEN** 某宏在 `x := Macro()` 处返回 `ExpandResult{Target: SpliceReplaceAssignRHS, Expr: e}`
- **THEN** `ValidateExpandResult` MUST 通过，且引擎 MUST 保留 `AssignStmt.Lhs`、仅替换含 `Macro` 的 `Rhs` 元素

#### Scenario: 语句宏替换整条 assign

- **WHEN** `TryExpand` 在 `x := Try(...)` 处返回 `ExpandResult{Target: SpliceReplaceAssignStmt, Stmts: ...}`
- **THEN** `ValidateExpandResult` MUST 通过，且引擎 MUST 用 `Stmts` 替换整条 `AssignStmt`

#### Scenario: 隐式字段推断已移除

- **WHEN** 某 `Expander` 返回 `ExpandResult{Stmts: ...}` 且未设置 `Target`（零值）
- **THEN** `ValidateExpandResult` MUST 失败

### Requirement: 轻薄 AST 辅助（首版）

`macro` 包首版 MUST 提供最小辅助（`Context`、`ExpandResult`、`SpliceTarget`、`ValidateExpandResult`、`TempIdent`、定位/错误辅助）。MUST NOT 在首版要求厚重 astbuilder；后续宏增多后再抽取。

#### Scenario: 首版无 astbuilder 依赖

- **WHEN** provider 作者实现 `Expander` 并构造展开 AST
- **THEN** MUST 仅依赖 `macro` 包首版 API，且 MUST NOT 要求引入独立 astbuilder 包

## ADDED Requirements

### Requirement: Context 提供合法贴回目标

`macro.Context` MUST 提供 `LegalSpliceTargets() []SpliceTarget`，返回当前宏调用在 AST 上**结构合法**的贴回目标集合（与 `internal/expander` 锚点解析规则一致）。`CallSiteKind`（`Site()`）MAY 保留，供 provider 做语义分支，但 MUST NOT 作为贴回依据。

#### Scenario: assign 处枚举两种 Target

- **WHEN** 宏调用为 `x := Macro()` 中的 `Macro(...)`
- **THEN** `LegalSpliceTargets()` MUST 包含 `SpliceReplaceAssignRHS` 与 `SpliceReplaceAssignStmt`

#### Scenario: 表达式槽仅 CallExpr

- **WHEN** 宏调用为 `1 + Macro(2)` 中的 `Macro(2)`
- **THEN** `LegalSpliceTargets()` MUST 仅包含 `SpliceReplaceCallExpr`

### Requirement: mactest 支持贴回校验

`macro/mactest` MUST 提供对 `ValidateExpandResult` 的封装（如 `Validate(ctx, result) error`），使 provider 在单测中无需运行全链路 expand 即可验证 `Target` 与载荷是否合法。

#### Scenario: mactest 校验非法 Target

- **WHEN** 测试对 `return Macro()` 语境构造 `ExpandResult{Target: SpliceReplaceCallExpr, Expr: e}` 并调用 `mactest.Validate`
- **THEN** MUST 返回错误，且错误 MUST 表明 `SpliceReplaceCallExpr` 不在合法目标集合中
