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

### Requirement: Quote 编写指引

原 `macro/quote` 章节 MUST 替换为 **`macro.Quote` + SyntaxRules**；MUST NOT 再要求 `@kind{ }` 或 import `macro/quote`。

#### Scenario: 无 quote 子包

- **WHEN** 读者按指南实现模板化宏
- **THEN** MUST 仅 import `macro` 根包（及必要 internal 无）

### Requirement: syntax-rules 作者指引

作者指南 MUST 说明 **`SyntaxRules` / `SyntaxCase`** 为默认 Expander 形态；pattern 使用 `$`、Quote 使用 `#`；裸标识符为 literal（无 `Literals` 字段）。MUST 说明 **MatchedSpan**：pattern match 哪部分，Apply 就只替换哪部分（Call/Decl 同一规则）；out 可含比 match 更多的 stmt/decl（Try 的 `if err`、Derive 的生成 methods）。

MUST 说明 **pattern 首版子集**（design D17、macro-pattern）：

- 顶层：`CallPattern` / `StmtPattern` / `DeclPattern`
- stub **invoked name** 匹配（`Try($inner)` 匹配 `tr.Try(...)`）
- assign：`$lhs ... :=`、`$lhs ... =`、`var $lhs ... =`（首版不支持包级 `var`）；return：`return $vals ... , Try($inner)`；ExprStmt：`Try($inner);` 为 stmt 级语法糖
- Decl anchor：embed `*ast.Field`；`MacroPos` 为 embed 位置（design D18）
- 裸 `Expander`：MUST 在返回前 `site.Match(pattern)`（design D19）
- Decl struct：**顺序无关**；`Derive[$iface]` 与 `$field ...` 书写顺序不影响 match
- ellipsis 经 `Bindings.Elems`；SyntaxCase 宽 Stmt clause 应排在窄 Call clause 前

#### Scenario: Inline 示例

- **WHEN** 读者查看 Call 宏最小示例
- **THEN** MUST 展示 `SyntaxRules` 单 clause，而非手写 `go/ast`

### Requirement: 统一 Expander 与 Context

作者指南 MUST 说明 **统一** `Expander(ctx Context, site Syntax) (Syntax, error)`；**不**区分 Call/Decl 宏章节为 normative 两套签名。`Context` MUST 文档化为三字段；`MacroPos` 自 **`site.MacroPos()`** 获取。

#### Scenario: Try EnclosingResults 文档

- **WHEN** 读者查看 Try 实现指引
- **THEN** MUST 说明 error 分支 `return` 使用 `EnclosingResults` + `ZeroSyntax`，不得仅依据 assign lhs

#### Scenario: Derive MatchedSpan 与生成 methods

- **WHEN** 读者查看 Derive 实现指引
- **THEN** MUST 说明 pattern 划定 type 边界、out 含新 TypeSpec 与 **新生成** methods 即可，**不必**复制 Target 既有未 match methods；并说明与 Try「替换载荷多 stmt」的对称性

#### Scenario: Derive SyntaxCase 完整示例

- **WHEN** 读者查看 Decl 宏迁移或 Derive 实现指引
- **THEN** MUST 提供与 design.md「Derive SyntaxCase 示例」等价的完整 walkthrough：使用方源码、pattern、`deriveTransform` 读取 `$item`/`$iface`/`Elems("field")`/`site.Underlying()` embed tag、Quote 产出 `[TypeSpec', FuncDecl...]`、引擎 Plan 两步贴回、与旧 `DeclExpander` 对比表、常见错误（embed 未移除、TypeSpec 改名、method 同名冲突）

#### Scenario: Decl embed 元数据与 Underlying

- **WHEN** 读者查看 Derive / Decl 宏实现指引
- **THEN** MUST 说明 **不**恢复 `DeclContext` / `TargetMethods()`；MUST 提供旧 API → 新路径对照（`MarkerTypeArgs` → `binds.Get("iface")` + `Types().TypeOf`；`MacroTag` → `*ast.Field.Tag` + `ParseMacroTag`；`$field ...` → `binds.Elems("field")` + `Underlying()` 为 `*ast.Field`）；MUST 说明 Decl pattern 不依赖 struct 字段顺序

### Requirement: BREAKING 迁移

作者指南 MUST 含迁移表：

| 旧 API | 新 API | 过渡 |
|--------|--------|------|
| `CallExpander` | `Expander` | MAY `TargetToPlan` adapter（短期） |
| `CallExpandResult` | `Syntax` + `SplicePlan` | 同上 |
| `DeclExpander` / `DeclExpandResult` | `SyntaxCase` + `Syntax` | **无 adapter**，必须改写 |
| `DeclContext.Site()` / `MacroTag` / `MarkerTypeArgs` | `Bindings` + `Underlying()` + `ParseMacroTag` / `Types().TypeOf` | 无专用 accessor |
| `DeclContext.TargetMethods()` | 不恢复；仅生成新 methods | 读已有 methods 仅 escape |
| `macro/quote` | `macro.Quote` | 直接迁移 |

#### Scenario: 迁移入口

- **WHEN** 现有 provider 作者阅读指南
- **THEN** MUST 区分 Call adapter 期限与 Decl 强制改写说明

