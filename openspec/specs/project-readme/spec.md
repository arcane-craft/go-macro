# project-readme Specification

## Purpose
TBD - created by archiving change improve-readme-structure. Update Purpose after archive.
## Requirements
### Requirement: README 必备章节顺序

根目录 `README.md` MUST 按以下顺序包含顶级章节（`##`），标题文案 MAY 微调但语义 MUST 等价：

1. 项目标题与 L0 概览（标题下首段，无需单独 `##`）— 一句话，不含角色分工与实现细节
2. `阅读指引`（或等价标题）— 说明不同读者应跳转的章节或文档
3. `项目说明`（或等价标题）— 工作原理概要，位于 `快速上手` 之前
4. `快速上手` — 先准备项目内 expand，再在使用宏的文件中接 generate，直至提交 gen 与日常构建
5. `命令` — 项目内 expand 与 `cmd/macro` 脚手架的表格或列表
6. `文档` — 链至 `docs/author-guide.md` 与 `examples/` 等
7. `编辑器` 或 `gopls` — 说明 `-tags=macro` 的原因及 `buildFlags` 配置
8. `参考` 或分拆的等价章节 — 官方宏库（contrib）、模块路径、本地联调

`快速上手` MUST 出现在 `参考` 类章节之前。

#### Scenario: 新读者沿主路径阅读

- **WHEN** 读者从 README 顶部顺序阅读至 `快速上手` 结束
- **THEN** MUST 能在不阅读 contrib 细节的情况下理解如何创建 expand、接 generate 并完成首次 expand 与 `go build`

#### Scenario: 章节顺序校验

- **WHEN** 维护者检查 README 的 `##` 标题列表
- **THEN** `快速上手` 的行号 MUST 小于任一包含「模块路径」或「contrib」语义的 `##` 或 `###` 参考节标题的行号

### Requirement: L0 概览一句话

README 在 `# go-macro` 标题后 MUST 仅包含一句简短概述（如「Go 过程宏框架」），MUST NOT 在 L0 段落中描述 expand 入口、gen 提交策略、作者/使用方分工或 `-tags macro` 等细节。

#### Scenario: 首屏简洁

- **WHEN** 读者仅阅读标题下首段
- **THEN** 该段 MUST 不超过两句话，且细节 MUST 位于独立的 `项目说明`（或等价）章节

### Requirement: 项目说明承载原理概要

README MUST 包含 `项目说明`（或等价标题）章节，用通俗语言概括过程宏如何工作（编写宏调用 → 展开生成 `*_macro_gen.go` → 日常构建使用生成代码；宏库与使用方项目的大致分工）。该章节 MUST 位于 `快速上手` 之前，且 SHOULD 避免实现级细节（如具体 API 名）堆砌。

#### Scenario: 理解项目定位

- **WHEN** 读者阅读标题下首段与 `项目说明`
- **THEN** MUST 能回答「此项目解决什么问题」而无需阅读 `快速上手`

### Requirement: 快速上手保留 normative 要点

`快速上手` MUST 包含且保持与 `macro-codegen`、`macro-repo-layout` 一致的下列语义（README 正文 MAY 使用自然语言，不必逐字保留 RFC 2119 英文关键词）：

- **首要步骤**：在使用宏的项目内创建 `cmd/macroexpand`（blank import 所需 `register`，调用 `expandtool.Main()`）
- 以 `examples/cmd/macroexpand` 与 `examples/readfile` 作为**对照示例**（非宏使用方默认长期命令）
- 使用宏的文件（带 `//go:build macro`）中的 `//go:generate` 指向**本项目** expand，RECOMMENDED 为 `go run ./cmd/macroexpand .`（或模块等价路径）
- 须在使用宏的文件中 import 所用宏库
- 对外发布的库 SHOULD 提交 `*_macro_gen.go`
- 日常 `go build` / `go test` 使用生成侧代码，不依赖读者长期开启 `-tags macro`

#### Scenario: 快速上手与 codegen spec 对齐

- **WHEN** 读者对照 `macro-codegen` 中关于 README 与 expand 入口的 MUST
- **THEN** README `快速上手` MUST 在 generate 之前说明项目内 `cmd/macroexpand`，并包含对外提交 gen 的表述

#### Scenario: examples 仅作对照

- **WHEN** 读者阅读 README `快速上手` 全文
- **THEN** MUST 能找到指向 `examples/cmd/macroexpand` 或 `examples/readfile` 的对照说明，且 MUST NOT 将 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand` 作为宏使用方项目的唯一或默认推荐命令

### Requirement: README 命令节与 cmd/macro 调用方式

README `命令` 节 MUST 列出 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>`（或等价表述）作为编写宏库时的脚手架入口。README MUST NOT 将 `go tool macro` 作为文档中的主要调用方式。

#### Scenario: 脚手架命令与文档一致

- **WHEN** 读者在 README `命令` 节查找如何初始化 provider
- **THEN** MUST 看到 `go run .../cmd/macro@latest init provider` 而非仅以 `go tool macro` 作为唯一示例

### Requirement: gopls 配置说明原因

README `gopls`（或等价）节 MUST 说明使用宏的源文件常带 `//go:build macro`，以及为何建议为 gopls 配置 `buildFlags: ["-tags=macro"]`（使编辑器分析宏版本源码）。

#### Scenario: 理解 gopls 配置

- **WHEN** 读者仅阅读 README `gopls` 节
- **THEN** MUST 能理解配置目的而不必查阅 author-guide

### Requirement: 参考内容与主路径分离

contrib 本地 `replace` 说明、双 module 测试命令（如 `GOWORK=off`）MUST 位于 `快速上手` 之后的参考类章节（`## 参考` 或其子标题 `###`），MUST NOT 插入快速上手编号步骤中间。

#### Scenario: 主路径无参考信息打断

- **WHEN** 读者执行 `快速上手` 中的步骤
- **THEN** MUST 能在连续阅读主路径步骤的过程中完成「准备 expand → 接 generate → 运行展开」，中间 MUST NOT 插入非操作性的长篇参考段落

### Requirement: README 不含迁移与兼容说明

README MUST NOT 包含历史模块路径迁移、旧 `contrib/` 路径升级、BREAKING 对照表或任何面向向后兼容的升级指引（含阅读指引中的迁移链接）。

#### Scenario: 无迁移相关内容

- **WHEN** 读者搜索 README 全文
- **THEN** MUST NOT 出现 `go-macro/contrib/`、路径对照迁移表或引导读者为兼容旧路径而阅读的专节

### Requirement: 与作者指南的职责边界

README MUST 链至 `docs/author-guide.md` 作为详细说明入口。README MUST NOT 复制 author-guide 级别的长契约正文（如完整 provider 实现说明、宏位置与返回字段表、主文件/生成文件 build tag 细节）。

`docs/author-guide.md` MUST 遵循 `author-guide` spec 的信息架构（含 `阅读指引`、角色分工、编写/使用方分节），并在 `阅读指引` 或首段链回 README 供宏使用方跳转。

#### Scenario: 深度文档跳转

- **WHEN** 读者需要 ignore/tag 或 provider 实现细节
- **THEN** README `文档` 节 MUST 提供指向 `docs/author-guide.md` 的链接

#### Scenario: 使用方从 author-guide 回到 README

- **WHEN** 仅想在使用方项目接 expand 的读者打开 author-guide
- **THEN** MUST 能在 `阅读指引` 或文档首段找到指向 README 快速上手的链接，且 MUST NOT 必须先阅读完整 provider 契约

### Requirement: 禁止无读者价值的 meta 节

README MUST NOT 包含仅描述 spec 对齐过程、且对终端用户无操作指引的独立章节（例如「术语澄清（无行为变更）」类尾注）。

#### Scenario: 无 spec 过程尾注

- **WHEN** 读者阅读 README 全部 `##` 标题
- **THEN** MUST NOT 存在以「无行为变更」或等价语义的 spec 对齐说明作为唯一目的的章节

### Requirement: README 面向用户的叙述风格

README 正文 SHOULD 使用自然中文说明步骤与原理；MAY 在 spec 校验需要处保留英文 RFC 2119 关键词，但 MUST NOT 以 spec/plan 文体为主（例如大量「不是 xxx」「MUST NOT 做 yyy」的边界划分类段落）。

#### Scenario: 可读性

- **WHEN** 新读者阅读 README 全文
- **THEN** MUST 能将其理解为项目使用指引，而非 OpenSpec 变更说明

