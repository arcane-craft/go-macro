## MODIFIED Requirements

### Requirement: 通用展开器分发

对每个识别到的宏站点（Call 调用点或 Decl 嵌入点），引擎 MUST **`ResolveSite` → `site Syntax`**，构造 **`Context`**，调用 **`Expander(ctx, site) (Syntax, error)`**。

引擎 MUST NOT 将 normative 分发建立在 `CallExpander(ctx, call *ast.CallExpr)` 与 `DeclExpander(ctx, DeclSite)` 分列 API 上（**Call** adapter 期 MAY 内部转换；**Decl** MUST NOT adapter，须 native `Expander`）。

#### Scenario: 按 syntax-id 分发统一 Expander

- **WHEN** 识别到 `Inline(...)` 且注册 `syntax-inline` Expander
- **THEN** 引擎 MUST 调用该 Expander 且传入 `site Syntax`

#### Scenario: Decl 站点统一 Expander

- **WHEN** 识别到嵌入 `Derive[T]` marker
- **THEN** 引擎 MUST 以 embed `*ast.Field` 为 anchor 调用 `ResolveSite` 构造 `site Syntax`，并调用同一 Expander 类型

### Requirement: ExpandResult 贴回（splice）

引擎 MUST 在 `Expander(ctx, site)` 返回 `out` 后，从 **`site` 内部 meta 槽**读取 **`MatchMeta`**（见 design D15、macro-syntax site 内部槽），再：

1. 调用 **`ValidateSplice(out, meta)`**
2. 调用 **`Apply(file, meta, out)`**
3. 调用 **`StampStmtPos(site.MacroPos(), ...)`**（适用时）

`MatchMeta` MUST NOT 由公开 `Expander` 返回值携带；`out Syntax` MUST NOT 携带 `Plan`。meta 槽为空时 MUST 在步骤 1 之前失败。

MUST NOT 要求 Expander 返回 `CallExpandResult` / `DeclExpandResult`（**Call** adapter 除外；**Decl** MUST NOT adapter）。MUST NOT 调用 normative **`InferTarget`**。Apply MUST 按 `meta.Plan` 执行；out 节点数 MAY 大于 MatchedSpan（Try 多 stmt、Derive 多 decl）。

#### Scenario: 推断 AssignStmt 贴回

- **WHEN** Try clause 为 stmt 级 pattern、MatchedSpan 为 AssignStmt、out 为多条 Stmts
- **THEN** Apply MUST 以 out 的 stmts 替换该 AssignStmt

#### Scenario: Derive 贴回不触碰未 match methods

- **WHEN** Derive clause match `type $item struct { ... }`、out 含 TypeSpec' 与生成 method
- **THEN** Apply MUST 仅处理 MatchedSpan 与 out 载荷；文件中未 match 的既有 methods MUST 保留

### Requirement: ApplyExpandResult 不依赖 CallSiteKind

贴回 MUST 依据 **`meta.Plan`**、**`MatchedSpan`** 与 `out Syntax` 形状；MUST NOT 将 `CallSiteKind` 作为分支条件（`CallSiteKind` MAY 从 author API 移除）。

#### Scenario: out 与 Plan 不兼容

- **WHEN** `meta.Plan` 要求 `ToExpr()` 但 `out` 无法 `ToExpr()`
- **THEN** `ValidateSplice` MUST 失败

## ADDED Requirements

### Requirement: ResolveSite

引擎 MUST 在 `internal/expander` 提供 `ResolveSite(file, anchor) Syntax`，在每轮 expand 前对 **当前 AST** 调用。anchor 类型（design D18）：

- **Call**：`*ast.CallExpr`
- **Decl**：embed `*ast.Field`

每轮 MUST 构造**空** meta 槽的 `site`。同句多宏 MUST 每轮重新 Resolve。

#### Scenario: splice 后重新 Resolve

- **WHEN** 同函数内两颗宏先后展开
- **THEN** 第二轮 ResolveSite MUST 反映第一轮 splice 后的 stmt 形状

#### Scenario: Decl ResolveSite anchor

- **GIVEN** `type T struct { provider.Derive[I]; X int }`
- **WHEN** 引擎展开 Derive
- **THEN** MUST 以 Derive 的 embed `*ast.Field` 为 anchor 调用 `ResolveSite`；`site.MacroPos()` MUST 为 embed 位置

### Requirement: MatchMeta 从 site 读取

每轮 expand 流水线 MUST 遵循：`ResolveSite` 构造**空** meta 槽的 `site` → `Expander(ctx, site) (Syntax, error)` → `MatchMetaFromSite(site)` → `ValidateSplice` → `Apply`。normative 路径下 meta 由 `SyntaxRules` / `SyntaxCase` 或手写 `Expander` 内 `site.Match` 写入（design D19）；Call adapter 路径由 `TargetToPlan` 写入。

#### Scenario: SyntaxCase 端到端 meta 来源

- **GIVEN** `SyntaxCase` 某 clause match 成功且 `Transform` 返回 `out`
- **WHEN** `Expander` 返回
- **THEN** 引擎 MUST 从同一 `site` 读取含该 clause `Plan` 的 `MatchMeta`；MUST NOT 要求 `Transform` 返回或设置 `MatchMeta`

#### Scenario: Call adapter 写入 meta 槽

- **GIVEN** Call adapter 包装旧 `CallExpander` 且 `TargetToPlan` 成功
- **WHEN** adapter 版 `Expander` 返回 `out`
- **THEN** meta 槽 MUST 已含 `TargetToPlan` 产出的 `MatchMeta`；引擎 MUST 与 normative 路径使用相同读取与 `ValidateSplice` / `Apply` 流程

## REMOVED Requirements

### Requirement: 分列 CallExpander 与 DeclExpander 分发（normative）

**Reason**: 统一 Expander 与 site Syntax。

**Migration**: **Call** MAY 经 `TargetToPlan` adapter；**Decl** MUST 改写 native `Expander`（无 adapter）。
