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

`编写宏库` MUST 包含且保持与 `macro-core`、`macro-codegen`、`macro-directive` 一致的下列语义（正文 MAY 使用自然语言，不必逐字保留 RFC 2119 英文关键词或设计文档内部代号如「方案 C」「首版」）：

- 每个语法桩与 Expander 的 doc MUST 含 `//macro: <syntax-id>`；同一 syntax-id 下多桩共享一个 Expander
- Provider：`Expand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error)` 签名约定
- 宏主文件 MUST import provider；`cmd/macro expand` 仅对已 import 且已生成 link 的 provider 展开
- MUST NOT 要求宏作者维护 `register/` 包
- 语法桩为包级 `panic` 函数，运行时不可调用
- **宏主文件中，已 link 注册的语法桩 MUST 仅以 `pkg.Stub(...)`（或 dot-import 下 `Stub(...)`）直接调用；MUST NOT 将桩作为函数值传递、赋值、返回或传入 `reflect.ValueOf` / `reflect.TypeOf`；违反时 expand MUST 失败（见 `macro-expander`）**
- **`ExpandResult` MUST 设置显式 `Target`（`SpliceTarget`）**；MUST 说明各 `Target` 替换的 AST 范围及对应载荷字段（`Stmts` / `Expr` / `Exprs`）；MUST 说明 `ctx.LegalSpliceTargets()` 与 `mactest.Validate`（或等价）用于单测校验；MAY 保留「调用处语境」表作阅读辅助，但 MUST NOT 暗示仅凭填哪个字段即可贴回
- MUST 说明 `SpliceReplaceAssignRHS`：保留赋值左侧，仅替换含宏调用的右侧表达式
- 展开时 `Context` MUST 提供 `EnclosingFunc()`（`*ast.FuncDecl` 或 `*ast.FuncLit`），供 provider 读取外层函数语境
- `init provider` 文档入口为 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>`
- 纯 Expand 单测使用 `macro/mactest` 的示例（含 `Validate`）

#### Scenario: 契约与 macro-directive 对齐

- **WHEN** 读者对照 `macro-directive` 与 `macro-core` 要求
- **THEN** author-guide `编写宏库` MUST 说明 per-function `//macro:`，且 MUST NOT 将 `register/` 列为作者职责

#### Scenario: 值用法约束可查

- **WHEN** 宏库作者或宏使用方阅读 `编写宏库` 或 `宏使用方`
- **THEN** MUST 能找到「桩须直调、不可作函数值」的说明，且 MUST 指向 expand 期报错行为

#### Scenario: 显式 Target 可查

- **WHEN** 宏库作者阅读 `编写宏库` 中 ExpandResult 说明
- **THEN** MUST 能找到 `Target` 与「替换整条语句 / 仅 RHS / 仅 CallExpr」的对照，且 MUST NOT 仅以「宏写在哪里 → 填 Stmts 或 Expr」隐式规则作为唯一说明

### Requirement: 宏使用方节保留 codegen 要点

`宏使用方` MUST 说明（语义与 `macro-codegen` 一致，文案 MAY 简化）：

- 宏主文件与 `*_macro_gen.go` 的 build tag 分工
- `go:generate` 调用 `cmd/macro expand`（或等价）生成侧文件
- 发布前在带/不带 `macro` tag 下测试的建议
- **已 link 的语法桩在宏主文件中 MUST 直接调用；将桩作为参数、变量或反射对象会导致 expand 失败**

#### Scenario: 使用方知悉 expand 约束

- **WHEN** 仅使用已有宏库的读者阅读 `宏使用方`
- **THEN** MUST 能理解 `import` 宏库后应写 `try.Try(...)` 一类直调，且 MUST NOT 依赖将 `try.Try` 当作普通函数值

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

