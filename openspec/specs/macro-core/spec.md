# macro-core Specification

## Purpose
TBD - created by archiving change go-macro-extension. Update Purpose after archive.
## Requirements
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

### Requirement: Provider 纯 Expand 测试辅助

`macro/mactest` MUST 支持统一 Expander 单测：`Expand(ctx, site)` 或包装 `ExpandCall`/`ExpandDecl` 调用新 API，并提供 Validate 等价能力。

#### Scenario: SyntaxRules 单测

- **WHEN** 测试 `SyntaxRules(Inline...)` snippet
- **THEN** MUST 无需 macro tag 即可 `go test`

### Requirement: 合法贴回目标枚举

`LegalSpliceTargetsForCall`（及 `CallContext.LegalSpliceTargets()`）MUST 与 `internal/expander` 锚点解析规则一致（规则不变）。

#### Scenario: assign 处枚举两种 Target

- **WHEN** 宏调用为 `x := Macro()` 中的 `Macro(...)`
- **THEN** `LegalSpliceTargets()` MUST 包含 `SpliceReplaceAssignRHS` 与 `SpliceReplaceAssignStmt`

### Requirement: init provider 脚手架

`github.com/arcane-craft/go-macro/cmd/macro` MUST 提供 `init provider` 子命令，生成**最小** provider 目录：含 `//macro:`、`Expand` 占位、**单个** panic 语法桩及 `expand_test.go`（mactest 模板）；MUST NOT 默认生成 Try 式多桩族模板。

用户文档 RECOMMENDED 通过 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>` 调用该子命令（`go tool macro` MAY 在已安装 tool 的环境下使用，但 MUST NOT 作为唯一文档入口）。

#### Scenario: 初始化新 provider

- **WHEN** 用户执行 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac`
- **THEN** MUST 创建可编译的 provider 骨架且文档指向作者指南

### Requirement: 语法桩运行时防护

provider 包内的语法桩（包级 panic 函数）在运行时 MUST panic，并标明宏名与不可直接调用。

#### Scenario: 直接调用语法桩

- **WHEN** 运行时代码直接调用 provider 包级语法桩（非经 expand 写回）
- **THEN** MUST panic 并提示不可直接调用

### Requirement: expandtool 展开入口

`macro/expandtool` MUST 提供 **`Register(syntaxID string, expand Expander)`**（或等价统一注册）。MAY 保留 **`RegisterCall`** 为 Call adapter 别名；**MUST NOT** 提供 **`RegisterDecl`** adapter 别名（Decl 须 native `Expander`）。

#### Scenario: 统一注册表 expand

- **WHEN** expand_runner 已 Register 多个 syntax-id Expander
- **THEN** MUST 展开已 import 且已 link 的所有宏站点

