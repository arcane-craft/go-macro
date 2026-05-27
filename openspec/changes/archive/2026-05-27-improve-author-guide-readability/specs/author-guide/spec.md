# author-guide Specification

## Purpose

定义 `docs/author-guide.md` 的信息架构与必备内容，使其面向宏库作者与人类读者可扫读、可分层深入，并与根目录 `README.md` 职责分离。不改变框架运行时行为。

## ADDED Requirements

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

`编写宏库` MUST 包含且保持与 `macro-core`、`macro-codegen` 一致的下列语义（正文 MAY 使用自然语言，不必逐字保留 RFC 2119 英文关键词或设计文档内部代号如「方案 C」「首版」）：

- Provider：`//macro: <syntax-id>` 与 `Expand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error)` 签名约定
- 宏主文件 MUST import provider；expand 工具仅对已 import 且已在 expand 二进制中 link 的包注册并展开
- 语法桩为包级 `panic` 函数，运行时不可调用
- `ExpandResult` 的 `Stmts` / `Expr` / `Exprs` 及宏出现位置与返回字段的对应关系（MAY 以表格呈现，标题 MAY 不使用「Site」术语）
- 展开时 `Context` MUST 提供 `EnclosingFunc()`（`*ast.FuncDecl` 或 `*ast.FuncLit`），供 provider 读取外层函数语境
- `init provider` 文档入口为 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>`
- 纯 Expand 单测使用 `macro/mactest` 的示例

#### Scenario: 契约与 macro-core 对齐

- **WHEN** 读者对照 `macro-core` 中 provider 与 init provider 要求
- **THEN** author-guide `编写宏库` MUST 包含上述签名、ExpandResult、EnclosingFunc 与 mactest 说明

### Requirement: 宏使用方节保留 codegen 要点

`宏使用方` MUST 包含且保持与 `macro-codegen`、`macro-repo-layout` 一致的下列语义（正文 MAY 使用「主文件 + 生成文件」等自然语言，不必出现「方案 C」字样）：

- 主文件 `//go:build macro`、生成侧 `//go:build !macro`、工具不修改主文件 build tag、生成代码含 `//line` 指向宏主文件
- expand 入口由宏使用方项目承载（blank import `register` + `expandtool.Main()`）；`examples/cmd/macroexpand` 为推荐参考实现，非框架内置唯一工具
- 对外发布建议：expand、`go test`（无 `-tags macro`）、提交 `*_macro_gen.go`、可选 CI `git diff --exit-code`

#### Scenario: expand 入口与 examples 定位清晰

- **WHEN** 读者阅读 `宏使用方` 中 expand 入口说明
- **THEN** MUST 理解 expand 入口由使用宏的项目自建，且 MAY 对照 `examples/cmd/macroexpand` 作为示例

### Requirement: 参考内容与主路径分离

contrib 模块路径、本地 `replace` / `go.work`、双 module 测试命令（如 `GOWORK=off`）MUST 位于 `编写宏库` 与 `宏使用方` 之后的参考类章节（`## 参考` 或其 `###` 子标题），MUST NOT 插入 `编写宏库` 或 `宏使用方` 的编号/步骤中间。

#### Scenario: 主路径无参考信息打断

- **WHEN** 读者连续阅读 `编写宏库` 各子节（含 `init provider` 与 mactest）
- **THEN** 中间 MUST NOT 插入非操作性的长篇本地联调段落

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
