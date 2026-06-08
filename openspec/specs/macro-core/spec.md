# macro-core Specification

## Purpose
TBD - created by archiving change go-macro-extension. Update Purpose after archive.
## Requirements
### Requirement: 宏上下文 API

`macro` 包 MUST 提供 **`CallContext`** 接口（过程宏 / 调用点宏），至少包含：`FileSet`、**`File`（`*ast.File`）**、`types.Info`、当前 `*types.Package`、宏调用 `*ast.CallExpr`、`StubName`、`SyntaxID`、`CallSiteKind`、**`LegalSpliceTargets() []SpliceTarget`**、**`EnclosingFunc`（`*ast.FuncDecl` 或 `*ast.FuncLit`）**、生成临时标识符（`TempIdent`）、`MacroPos()`。

`EnclosingFunc` 为通用调用语境，供各 Call provider 自行解释；`macro` 包 MUST NOT 内置 Try 专用 k 规则。

过程宏 MUST NOT 使用已移除的 `Context` 类型名；文档与 API MUST 使用 `CallContext`。

#### Scenario: 展开器获取调用语境

- **WHEN** `TryExpand` 处理 `return Try(expr)`
- **THEN** `ctx.Site()` MUST 为 `SiteReturn`，且 `ctx.EnclosingFunc()` MUST 提供外层返回签名

#### Scenario: 框架不内置 Try k 规则

- **WHEN** 阅读 `macro` 包公开 API
- **THEN** MUST NOT 出现仅服务于 `Try`/`Try2` 的 k 校验函数或常量

### Requirement: ExpandResult 与 Expander 签名

`macro` 包 MUST 定义（过程宏）：

- `SpliceTarget` 枚举（名称保留），至少包含六种 `SpliceReplace*`
- **`CallExpandResult`** 含必填 `Target SpliceTarget` 与载荷 `Stmts`/`Exprs`/`Expr`
- **`CallExpander func(ctx CallContext, call *ast.CallExpr) (CallExpandResult, error)`**

贴回语义仅由 `Target` 与对应载荷决定。语法桩 MUST 在**各桩函数** doc 中含 `//macro: <syntax-id>`。Call Expander 函数 MUST 在同 syntax-id doc 下符合 `CallExpander` 签名。

`macro` 包 MUST 提供 **`ValidateCallExpandResult(ctx CallContext, result CallExpandResult) error`**。

`CallSiteKind`（`Site()`）MAY 供 provider 语义分支，MUST NOT 作为引擎贴回依据。

#### Scenario: 表达式宏显式 Target

- **WHEN** 某 Call 宏返回 `CallExpandResult{Target: SpliceReplaceCallExpr, Expr: e}`
- **THEN** `ValidateCallExpandResult` MUST 通过

#### Scenario: 隐式字段推断已移除

- **WHEN** 某 `CallExpander` 返回未设置 `Target` 的 `CallExpandResult`
- **THEN** `ValidateCallExpandResult` MUST 失败

### Requirement: 通用宏注册与查找

系统 MUST 支持：对宏主文件已 import 的 provider，解析 **函数桩** 与 **Call Expander**、**marker 类型** 与 **Decl Expander** 上的 `//macro: <syntax-id>`，并结合 link 构建注册表。

**同一 provider 包 MUST 允许多个 syntax-id**。每个 syntax-id MUST 独立映射 Call 桩集合、Decl marker 集合、可选 `CallExpander`、可选 `DeclExpander`。

Call 宏识别 MUST 使用 `(syntax-id 或 importPath, stubFuncName)` 查找 `CallExpander`（实现 MUST 支持多 syntax-id per import path）。

`internal/expander` 与 `macro/expandtool` MUST NOT 硬编码具体 provider Expander。

#### Scenario: 注册多桩名到同一 syntax-id

- **WHEN** `try` 包中 `syntax-try` 绑定 `TryExpand`，存在桩 `Try` 与 `Try2`
- **THEN** 注册表 MUST 将 `Try`、`Try2` 映射到同一 `CallExpander`

#### Scenario: 同包多 syntax-id

- **WHEN** provider 含 `derive-stringer` 与 `wire-json` 两种 marker 类型
- **THEN** 注册表 MUST 分别映射到各自 DeclExpander

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

### Requirement: Provider 纯 Expand 测试辅助

`macro/mactest` MUST 支持 Call 与 Decl 单测：`ExpandCall` / `ExpandDecl`（或等价）及 `ValidateCall` / `ValidateDecl`。

#### Scenario: TryExpand 单测

- **WHEN** `try` 包测试调用 `mactest.ExpandCall(TryExpand, snippet)`
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

`macro/expandtool` MUST 提供：

- **`RegisterCall(syntaxID string, expand CallExpander)`**（或等价，按 syntax-id 注册 Call）
- **`RegisterDecl(syntaxID string, expand DeclExpander)`**（或等价）
- **`RegisteredCall() map[string]CallExpander`**、**`RegisteredDecl() map[string]DeclExpander`**（或合并视图供 expand 使用）
- **`Run(args, linked)`** — `linked` 结构 MUST 能同时携带 Call 与 Decl 注册表

MUST NOT 仅保留无区分的 `Register(importPath, Expander)` 作为唯一入口（可保留兼容别名至实现完成，spec 以分列注册为准）。

#### Scenario: Main 使用 Registered 注册表

- **WHEN** expand_runner 已 `RegisterCall` / `RegisterDecl` 且执行 expand
- **THEN** MUST 展开已 import 且已 link 的 Call 与 Decl 宏

