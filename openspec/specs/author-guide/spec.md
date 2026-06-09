# author-guide Specification

## Purpose

定义 `docs/author-guide.md` 的信息架构与必备内容，使其面向宏库作者与人类读者可扫读、可分层深入，并与根目录 `README.md` 职责分离。不改变框架运行时行为。
## Requirements
### Requirement: 作者指南必备章节顺序

`docs/author-guide.md` MUST 按以下顺序包含顶级章节（`##`），标题文案 MAY 微调但语义 MUST 等价：

1. `阅读指引`（或等价标题）— 说明宏库作者、宏使用方、仅需查参考的读者应跳转的章节或 README
2. `角色分工` — 框架、provider、使用方职责说明（仅需列出各角色职责，MAY 使用两列表格）；明确 `examples` 为示例调用方项目
3. `编写宏库` — `init provider` 脚手架、框架契约、宏出现位置与 `ExpandResult` 字段对应、纯 Expand 单测（可用 `###` 子节；子节顺序 MAY 将脚手架置于契约之前以利于上手）
4. `宏使用方` — 主文件 + 生成文件的 build tag 布局（与 `macro-codegen` 方案 C 语义一致）、expand 入口、发布前建议（可用 `###` 子节）
5. `参考`（或等价标题）— 官方 contrib、本地联调、消费第三方宏库

`编写宏库` MUST 出现在 `参考` 类章节之前。

#### Scenario: 宏作者沿主路径阅读

- **WHEN** 读者从 author-guide 顶部顺序阅读至 `编写宏库` 结束
- **THEN** MUST 能在不阅读 contrib 本地联调细节的情况下理解 provider 契约、如何 `init provider` 以及如何编写 mactest 单测

#### Scenario: 章节顺序校验

- **WHEN** 维护者检查 author-guide 的 `##` 标题列表
- **THEN** `编写宏库` 的行号 MUST 小于任一包含「本地联调」语义的参考节 `##` 或 `###` 标题的行号

### Requirement: 阅读指引区分读者类型

`阅读指引` MUST 用表格或等价短列表说明至少两类读者去向：

- 编写宏库（provider）→ `编写宏库` 各节（MAY 指向脚手架子节作为首选入口）
- 在项目里使用宏 → 链至 README 快速上手或 author-guide 内 `宏使用方` 节

#### Scenario: 使用方不被误导

- **WHEN** 仅想在使用方项目接 expand 的读者阅读 `阅读指引`
- **THEN** MUST 能找到指向 README 或 `宏使用方` 的明确跳转，且 MUST NOT 必须先阅读 provider 契约长段

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

### Requirement: 参考内容与主路径分离

contrib 模块路径、本地 `replace` / `go.work`、双 module 测试命令（如 `GOWORK=off`）MUST 位于 `编写宏库` 与 `宏使用方` 之后的参考类章节（`## 参考` 或其 `###` 子标题），MUST NOT 插入 `编写宏库` 或 `宏使用方` 的编号/步骤中间。

`参考` 节中关于官方宏库（inline/try）的版本兼容与使用说明 MUST 链至 `go-macro-contrib` 仓库 README。author-guide MUST NOT 向终端读者提及 OpenSpec、`openspec/` 目录或内部规范工作流。

#### Scenario: 主路径无参考信息打断

- **WHEN** 读者连续阅读 `编写宏库` 各子节（含 `init provider` 与 mactest）
- **THEN** 中间 MUST NOT 插入非操作性的长篇本地联调段落

#### Scenario: 官方宏库文档外链

- **WHEN** 读者在 `参考` 节查找 Try/Inline 或 contrib 版本兼容说明
- **THEN** MUST 能找到指向 `github.com/arcane-craft/go-macro-contrib` README 的链接，且 MUST NOT 出现 OpenSpec 或 `openspec/` 路径

### Requirement: 禁止无读者价值的 meta 节

`docs/author-guide.md` MUST NOT 包含仅描述 spec 对齐过程、且对终端读者无操作指引的独立章节（例如「术语澄清（无行为变更）」类尾注）。

#### Scenario: 无 spec 过程尾注

- **WHEN** 读者阅读 author-guide 全部 `##` 标题
- **THEN** MUST NOT 存在以「无行为变更」或等价语义的 spec 对齐说明作为唯一目的的章节

### Requirement: 作者指南面向人类的叙述风格

author-guide 正文 SHOULD 使用自然中文说明步骤与契约；MAY 在 spec 校验需要处保留英文 RFC 2119 关键词，但 MUST NOT 以 spec/plan 文体为主（例如大量「不是 xxx」「MUST NOT 做 yyy」的边界划分类段落堆砌于首屏，或要求读者理解「方案 A/B/C」等内部设计代号）。

角色分工 SHOULD 仅描述各角色职责，SHOULD NOT 以「不负责」列作为读者必读结构。

#### Scenario: 可读性

- **WHEN** 新宏作者阅读 author-guide 全文
- **THEN** MUST 能将其理解为编写与发布宏库的指引，而非 OpenSpec 变更说明

### Requirement: 与 README 的职责边界

author-guide MUST 在 `阅读指引` 或首段链至根目录 `README.md`（快速上手）。author-guide MUST NOT 复制 README 级别的逐步 expand 教程全文；README MUST NOT 复制 author-guide 级别的 provider 契约与宏位置/返回字段长段（见 `project-readme`）。

#### Scenario: 深度文档互链

- **WHEN** 读者需要 ignore/tag 或 provider 实现细节
- **THEN** README `文档` 节 MUST 链至 author-guide；author-guide MUST 链回 README 供使用方跳转

### Requirement: 声明宏作者指南子节

`编写宏库` MUST 含 `###` 或等价子节「声明宏（Decl）」或并入「框架契约」，说明：

- Marker 四种模板（无参、`[T]`、tag、组合）
- `DeclExpandResult` 全量 `Fields`/`Methods` 义务
- Contract 与 Wire 首期场景示例指向 contrib：`github.com/arcane-craft/go-macro-contrib/derive`、`.../wirejson`

#### Scenario: 全量 Methods 义务可查

- **WHEN** 作者阅读 Decl 契约
- **THEN** MUST 明确「成功返回须包含 Target 全部方法，漏方法即丢失」

### Requirement: 编写宏库 Quote 子节

`docs/author-guide.md` 的 `编写宏库` MUST 含 `###` 或等价子节「模板化 AST（Quote）」或并入「框架契约」，说明：

- `macro/quote` 为**可选**子包；与手写 AST 的关系
- 四种 kind 与 `Expr`/`Exprs`/`Stmts`/`Decls` 对应；typed API 直接写 body，不必再包 `@kind{ }`；`quote.Quote` 仍需 `@expr{ }` 等显式根
- `#name` 填洞；绑定类型（string ident、`ast.Expr`、`[]ast.Stmt`、嵌套 Quote 等）
- 产出与 `CallExpandResult` / `DeclExpandResult` 及 `SpliceTarget` 的对应表
- Expander 内在 `quote.Stmts` 等之后 MUST 调用 `macro.StampStmtPos(ctx.MacroPos(), ...)`（Call）或等价行号策略
- 模板注释会保留至产出 AST 的说明（一句即可）

MUST NOT 要求所有 provider 使用 Quote。

#### Scenario: 作者查 Quote 根 kind

- **WHEN** 宏作者阅读 `编写宏库` Quote 子节
- **THEN** MUST 能找到四种 kind 与 `Expr`/`Exprs`/`Stmts`/`Decls` 的对应，以及 typed API 可直接写 body 的说明

#### Scenario: 作者查贴回衔接

- **WHEN** 宏作者阅读 Quote 子节
- **THEN** MUST 能找到 `@stmts` → `CallExpandResult.Stmts` + 显式 `Target` 的示例或表格

