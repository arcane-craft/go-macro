## MODIFIED Requirements

### Requirement: 编写宏库保留 normative 契约要点

`编写宏库` MUST 包含且保持与 `macro-core`、`macro-codegen`、`decl-macro` 一致的下列语义：

**过程宏（Call）**

- 语法桩为包级 `panic` 函数，doc 含 `//macro: <syntax-id>`
- **`CallExpander(ctx macro.CallContext, call *ast.CallExpr) (macro.CallExpandResult, error)`**
- **`CallExpandResult` MUST 设置 `Target`（`SpliceTarget`）**；说明 `ctx.LegalSpliceTargets()` 与 `mactest.ValidateCall`
- 桩须直调，不可作函数值（见 `macro-expander`）

**声明宏（Decl）**

- marker 为类型定义，类型 doc 含 `//macro: <syntax-id>`
- 使用方通过 struct **匿名嵌入** marker；可选参数仅 `` `macro:"k=v"` ``
- **`DeclExpander(ctx macro.DeclContext, site macro.DeclSite) (macro.DeclExpandResult, error)`**
- 成功时 **`Fields` 与 `Methods` 均 MUST 全量**返回；`mactest.ValidateDecl`
- marker 类型内 struct 字段仅作文档提示，引擎不读取
- Decl 作用域：仅 Target 的字段与方法；MUST NOT 生成包级 const/var、其它类型、独立测试文件

**通用**

- 同一 provider 包 **允许多个 syntax-id**；Call 与 Decl Expander **分别 link**
- 宏主文件 MUST import provider；`cmd/macro expand` 仅对已 import 且已 link 的 syntax 展开
- MUST NOT 要求 `register/` 包

#### Scenario: Call 与 Decl 签名可查

- **WHEN** 宏库作者阅读 `编写宏库`
- **THEN** MUST 能区分 `CallExpander` 与 `DeclExpander` 签名及触发方式

#### Scenario: Decl Marker 模板可查

- **WHEN** 作者实现声明宏
- **THEN** MUST 能找到无参 / `Marker[T]` / `` `macro:"..."` `` 模板说明

#### Scenario: 显式 Target 可查

- **WHEN** 作者阅读 Call 宏 ExpandResult 说明
- **THEN** MUST 能找到 `Target` 与 splice 范围对照表

### Requirement: 宏使用方节保留 codegen 要点

`宏使用方` MUST 说明：

- 宏主文件与 `*_macro_gen.go` 的 build tag 分工
- **生成侧含展开后的类型、方法与函数**（不仅函数）
- expand 顺序：引擎先 Decl 后 Call（使用者通常无感，MAY 一句带过）
- 桩直调与 Decl 嵌入规则

#### Scenario: gen 含类型

- **WHEN** 使用方阅读 `宏使用方`
- **THEN** MUST 理解 `!macro` 构建下类型定义来自 `*_macro_gen.go`
