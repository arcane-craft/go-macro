## Context

`README.md` 涵盖框架简介、五步快速上手、命令表、gopls、文档链接、contrib 说明、模块路径与「术语澄清」尾注。内容与 `macro-codegen`、`macro-repo-layout` 等 spec 已对齐，但呈现上存在：

- **主路径被打断**：读者在步骤 1–5 之后需跳过 contrib、模块路径等参考信息才能回到「我只想把宏跑起来」。
- **层级扁平**：`##` 下混有操作指南、参考表与 meta 说明，缺少「先做什么 / 后查什么」的分层。
- **尾注易误解**：「术语澄清（无行为变更）」对新人无帮助，宜并入设计说明或删除。

约束：文档须继续满足 `macro-codegen` 等对 README 的 MUST（如对外库提交 gen、使用方承载 expand、RECOMMENDED generate 一行）；中文为主，保留 RFC 2119 英文关键词。

## Goals / Non-Goals

**Goals:**

- 定义并落实固定的 README 信息架构（见 `project-readme` spec）。
- 让读者在首屏获得：项目是什么、适合谁、最小可行步骤、去哪看详细契约。
- 将本地 `replace`、双 module 测试等放入「参考」层，不删减事实；**不**在 README 保留历史 `contrib/` 的 BREAKING 迁移表（见 CHANGELOG）。

**Non-Goals:**

- 不重写 `docs/author-guide.md` 正文（仅确保链接有效）。
- 不新增功能、不改 `go.mod`、不调整 OpenSpec 运行时 spec。
- 不引入多语言 README。

## Decisions

### 1. 采用「三层阅读」结构

| 层级 | 章节（建议 `##` 标题） | 内容 |
|------|------------------------|------|
| L0 概览 | （标题下首段） | 一句话：项目是什么 |
| L0.5 说明 | 项目说明 | 工作原理、角色分工、gen / expand / import 等细节 |
| L1 主路径 | 快速上手、日常构建 | 编号步骤、generate、expand、提交 gen、`go build` |
| L2 工具与导航 | 命令、gopls、文档与示例 | 表格与短配置 |
| L3 参考 | 官方宏库、模块与仓库 | contrib、路径表、`replace` 联调 |

**理由**：符合「扫读 → 动手 → 深挖」；与 `go-macro-extension` 设计里「根 README 面向快速上手，详细契约以作者指南为准」一致。

**备选**：维持现状仅润色措辞 —— 拒绝，无法解决结构问题。

### 2. 新增「阅读指引」短节（可选 `## 阅读指引`）

用 3 行说明三类读者去向：

- 使用宏的应用开发者 → 快速上手 + 命令
- 宏库作者 → `docs/author-guide.md`
**理由**：降低误读「本 README = 完整规范」的预期；不描述历史路径迁移或兼容升级。

### 3. 删除或合并「术语澄清（无行为变更）」

该段为 spec 对齐过程的 meta 说明，对终端用户无操作价值。若需保留工程上下文，可移至 `design.md` 或 CHANGELOG，不留在 README。

### 4. MUST/RECOMMENDED 保留策略

- 快速上手与「提交 gen」等 normative 句保留英文 **MUST** / **RECOMMENDED**。
- 路径列表为描述性内容，不必强行加 RFC 关键词。

### 5. 与 author-guide 的分工

| 文档 | 职责 |
|------|------|
| README | 安装级步骤、默认命令、链接入口 |
| author-guide | 宏契约、ignore/tag、provider 实现细节 |

README 中「文档」节仅保留链接与一行说明，不复制 author-guide 长段。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 重组后外部博客/.issue 引用的锚点失效 | 尽量保留原 `##` 标题文案（如「快速上手」「命令」）；若改名则在 CHANGELOG 记一笔 |
| 附录过长导致仍难扫读 | L3 使用子标题 `###`，参考信息保持精简 |
| 与 spec 字面不一致 | apply 阶段对照 `macro-codegen`、`macro-repo-layout` 必读句做 diff 核对 |

## Migration Plan

1. 按 `tasks.md` 起草新 README 大纲（空壳标题）。
2. 自旧 README 逐段迁移内容至对应层级，不删技术事实。
3. 人工通读 + `openspec validate`（若适用）。
4. 无回滚依赖；若不满意可 git revert 单文件。

## Open Questions

（无。结构方案已在 proposal 与 spec 中闭合。）
