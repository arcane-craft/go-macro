## ADDED Requirements

### Requirement: SyntaxRules 纯 pattern-template

系统 MUST 提供 `SyntaxRules(clauses ...Clause) Expander`。每个 `Clause` MUST 含 `Pattern` 与 `Template` 字符串。MUST NOT 含 `Literals` 字段。match 成功时 MUST 将 **MatchMeta**（含 **MatchedSpan** 与 **Plan**）写入 **`site` 内部 meta 槽**（见 macro-syntax）；`Template` 产出为 `out`（节点数 MAY 大于 MatchedSpan）。公开 `Expander` 返回值 MUST 仅为 `out`。

#### Scenario: Inline 单 clause

- **WHEN** Clause 为 `Pattern: Inline($v)`, `Template: #v`
- **THEN** Expander 对匹配 site MUST 返回等价于 `v` 的 Syntax

#### Scenario: Try stmt 级替换载荷扩展

- **WHEN** Clause 为 `Pattern: $lhs ... := Try($inner)`，`Template` 展开为多条 stmt（含 `if err`）
- **THEN** MatchedSpan MUST 为 AssignStmt，out MUST 可作为该 span 的多 stmt 替换载荷

### Requirement: SyntaxCase fender 与 transform

系统 MUST 提供 `SyntaxCase(clauses ...Clause) Expander`。`Clause` MAY 含 `Fender func(Context, Syntax, Bindings) error` 与/或 `Transform func(Context, Syntax, Bindings) (Syntax, error)`（与 `Template` 二选一）。`Clause` MAY 含 **`Plan []SpliceStep` override**（非零时替代 pattern 默认 Plan；首版 MAY 仅 adapter 使用）。

#### Scenario: fender 失败试下一 clause

- **WHEN** 第一 clause match 成功但 fender 返回 error，第二 clause match 成功
- **THEN** MUST 使用第二 clause 结果

#### Scenario: fender 失败清空 meta 槽

- **WHEN** 第一 clause match 成功但 fender 返回 error
- **THEN** MUST 清空 `site` meta 槽后再尝试下一 clause；不得将失败 clause 的 `Plan` 传入 `ValidateSplice`

### Requirement: Clause 顺序与错误

Pattern **解析**失败 MUST 在加载/注册时 fatal。Runtime **match**失败 MUST 尝试下一 clause。全部失败 MUST 返回 `no matching syntax rule`（或等价 error）。

更宽 **StmtPattern** clause SHOULD 排在更窄 **CallPattern** clause 之前（如先 `$lhs ... := Try($inner)`，再 `Try($inner)`），以免窄 pattern 抢占宽语义。

#### Scenario: 宽 pattern 优先

- **GIVEN** `SyntaxCase` clause 1 为 `$lhs ... := Try($inner)`，clause 2 为 `Try($inner)`，site 为 `x, err := Try(f())`
- **WHEN** Expander 执行
- **THEN** MUST 使用 clause 1（stmt 级 MatchedSpan），MUST NOT 仅用 clause 2 的 call 级 Plan

#### Scenario: 无匹配 clause

- **WHEN** 所有 clause match 均失败
- **THEN** Expander MUST 返回非 nil error

#### Scenario: pattern 解析失败

- **WHEN** Clause.Pattern 非法且 Expander 注册时校验
- **THEN** MUST fatal 错误，不得 silent 忽略

### Requirement: 不提供 syntax-id 特化 built-in

`macro` 根包 MUST NOT 提供仅服务于单一 stub（如 Try）的框架级 Quote 洞注入或 Transform built-in（如 `ErrorReturn`、`ZerosFor`）。Try 等 provider MUST 使用通用 API（如 `EnclosingResults`、`ZeroSyntax`）在 Transform 内组合。

#### Scenario: 框架无 Try 专用 API

- **WHEN** 阅读 `macro` 包公开 API 列表
- **THEN** MUST NOT 出现 `TryExpand` 专用 helper 或 `ZerosForTry` 等名称
