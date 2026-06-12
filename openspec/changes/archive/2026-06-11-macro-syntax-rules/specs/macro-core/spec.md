## MODIFIED Requirements

### Requirement: 宏上下文 API

`macro` 包 MUST 提供 **`Context`** 接口，且 **仅** 包含：`FileSet()`、**`Types() *types.Info`**、**`TempIdent(prefix string) *ast.Ident`**。

`Context` MUST NOT 包含 `File()`、`*ast.CallExpr` / `Call()`、`StubName`、`SyntaxID`、`CallSiteKind` / `Site()`、`LegalSpliceTargets()`、`EnclosingFunc()`、`MacroPos()`、`*types.Package`（除非后续 ADDED 明确恢复）。

宏位置 MUST 由 **`site Syntax` 的 `MacroPos()`** 提供。外层函数签名 MUST 通过 **`EnclosingSignature` / `EnclosingResults`**（见 macro-enclosing-signature）获取，MUST NOT 通过 `EnclosingFunc()` AST API。

`macro` 包 MUST NOT 内置 Try 专用 k 规则或 Try 专用 Quote 洞注入。

#### Scenario: 极简 Context

- **WHEN** provider 实现 `Expander` 且仅需 typecheck
- **THEN** MUST 可仅使用 `ctx.FileSet()`、`ctx.Types()`、`ctx.TempIdent()`

#### Scenario: 框架不内置 Try k 规则

- **WHEN** 阅读 `macro` 包公开 API
- **THEN** MUST NOT 出现仅服务于 `Try`/`Try2` 的 k 校验函数或常量

### Requirement: ExpandResult 与 Expander 签名

`macro` 包 MUST 定义：

- **`Expander func(ctx Context, site Syntax) (Syntax, error)`** — **统一**宏展开签名；MUST NOT 再区分 `CallExpander` / `DeclExpander` 为 normative 作者 API

语法桩 MUST 在 doc 中含 `//macro: <syntax-id>`。Expander MUST 在同 syntax-id doc 下注册。

引擎 MUST 通过 **`ValidateSplice(out, meta)` + `Apply(file, meta, out)`** 贴回（`meta` 含 **MatchedSpan** 与 Match 时确定的 **`Plan []SpliceStep`**；无 normative `InferTarget`）；Expander MUST NOT 返回 `CallExpandResult` / `DeclExpandResult` 作为 normative 作者 API。

**Migration**：MAY 提供短期 **Call-only** adapter：`CallExpander` → `Expander`；adapter MUST 将 `CallExpandResult.Target` 编译为 `[]SpliceStep`（`TargetToPlan`）。**MUST NOT** 提供 `DeclExpander` / `DeclExpandResult` adapter（方案 C）。

Provider MAY 实现手写 `Expander`（非 `SyntaxRules` / `SyntaxCase`），但 MUST 在返回 `out` 前对 `site` 调用 `Match(pattern)` 写入 meta 槽（design D19）；**MUST NOT** 暴露 `SetMatchMeta` 给 provider。

#### Scenario: SyntaxRules Expander

- **WHEN** provider 注册 `SyntaxRules(clause...)` 为 Expander
- **THEN** expand MUST 成功且贴回由 Match 产出的 `Plan` + `Apply` 完成

#### Scenario: 旧 CallExpandResult 作者 API deprecated

- **WHEN** provider 直接返回 `CallExpandResult`
- **THEN** MUST 经 adapter 过渡；最终 MUST 迁移为 `Syntax`

### Requirement: 通用宏注册与查找

系统 MUST 支持：对宏主文件已 import 的 provider，解析 **函数桩**、**marker 类型** 与 **Expander** 上的 `//macro: <syntax-id>`，并结合 link 构建注册表。

**同一 provider 包 MUST 允许多个 syntax-id**。每个 syntax-id MUST 映射桩集合与 **单一 `Expander`**（MAY 由 `SyntaxCase` 内多 clause 覆盖 Call/Decl 形态）。

注册表 MUST NOT 将 Call Expander 与 Decl Expander 作为 normative 分列类型（adapter 期 MAY 保留内部分发）。

#### Scenario: 同包多 syntax-id

- **WHEN** provider 含 `syntax-try` 与 `syntax-derive` 两个 Expander
- **THEN** 注册表 MUST 分别映射

### Requirement: 轻薄 AST 辅助（首版）

`macro` 包 MUST 提供：`Context`、`Syntax`、`Bindings`（含 `Get` 与 `Elems`）、`Quote`、`SyntaxRules`、`SyntaxCase`、`EnclosingSignature` / `EnclosingResults` / `ZeroSyntax`、`ErrorAt`、定位辅助。`MatchMeta` / `SpliceStep` MUST 为引擎内部类型（定义于 `internal/expander`，**非** provider 公开 API）；贴回 meta 经 **`site` 内部槽**传递（design D15），不得要求作者构造或返回 `MatchMeta`。

模板化 AST MUST 通过 **`macro.Quote` + `SyntaxRules`** 完成；**MUST NOT** 要求 import 独立 `macro/quote` 子包（见 macro-template REMOVED）。

#### Scenario: SyntaxRules 实现 Inline

- **WHEN** provider 使用 `SyntaxRules` 且未手写 `go/ast`
- **THEN** MUST 可 `go test` 且 expand 行为正确

## REMOVED Requirements

### Requirement: CallContext 与 CallExpander 分列 API

**Reason**: 统一为 `Context` + `Expander(ctx, site Syntax) (Syntax, error)` 与 syntax-rules 模型。

**Migration**: 使用 adapter 将 `CallExpander(ctx, call)` 包装为新 Expander；见 author-guide。

### Requirement: DeclContext 与 DeclExpander 分列 API（macro-core 侧）

**Reason**: 统一 Expander；Decl 站点由 `site Syntax` 表达。

**Migration**: Derive 等 MUST 改写 `SyntaxCase`；**无 Decl adapter**；Call 见 `TargetToPlan`。

### Requirement: CallExpandResult 作者必填 Target

**Reason**: 贴回计划在 Match 时确定；见 macro-splice-apply。

**Migration**: Transform 返回 `Syntax`；引擎使用 `meta.Plan` + `ValidateSplice` + `Apply`。

## MODIFIED Requirements

### Requirement: Provider 纯 Expand 测试辅助

`macro/mactest` MUST 支持统一 Expander 单测：`Expand(ctx, site)` 或包装 `ExpandCall`/`ExpandDecl` 调用新 API，并提供 Validate 等价能力。

#### Scenario: SyntaxRules 单测

- **WHEN** 测试 `SyntaxRules(Inline...)` snippet
- **THEN** MUST 无需 macro tag 即可 `go test`

### Requirement: expandtool 展开入口

`macro/expandtool` MUST 提供 **`Register(syntaxID string, expand Expander)`**（或等价统一注册）。MAY 保留 **`RegisterCall`** 为 Call adapter 别名；**MUST NOT** 提供 **`RegisterDecl`** adapter 别名（Decl 须 native `Expander`）。

#### Scenario: 统一注册表 expand

- **WHEN** expand_runner 已 Register 多个 syntax-id Expander
- **THEN** MUST 展开已 import 且已 link 的所有宏站点

## REMOVED Requirements

### Requirement: SpliceTarget 与 LegalSpliceTargets 作为 normative 作者 API

**Reason**: 贴回第一性概念改为 `SplicePlan` / `SpliceStep`；`SpliceTarget` MAY 保留于 adapter 内部（`TargetToPlan`）。

**Migration**: 删除作者必填 `Target`；见 macro-splice-apply 与 author-guide。

### Requirement: LegalSpliceTargetsForCall 作为作者 API

**Reason**: 贴回计划在 Match 时确定；作者 MAY 不调用。

**Migration**: 删除 `CallContext.LegalSpliceTargets()`；文档说明贴回由 `SplicePlan` 执行。
