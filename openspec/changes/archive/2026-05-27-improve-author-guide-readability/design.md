## Context

改版前的 `docs/author-guide.md` 涵盖：框架契约、角色分工、纯 Expand 单测、主文件/生成文件的 build tag（设计文档中称方案 C）、`init provider`、发布 checklist、官方 contrib 与本地联调、第三方 register、Try 桩族表等。内容与 `macro-codegen`、`macro-core`、`macro-repo-layout` 等实现 spec 已对齐，但呈现上存在：

- **缺少入口导航**：无「阅读指引」，宏作者与使用方难以快速跳转。
- **层级扁平**：契约、测试、build tag、contrib 混在同一层级，缺少「编写宏库」与「宏使用方」两大主路径。
- **meta 尾注**：「术语澄清（无行为变更）」对终端读者无操作价值。
- **风格不一致**：规范腔、内部代号（方案 C、首版）、划界式「不负责」表述，与 README 的自然中文不对齐。
- **参考信息打断主路径**：官方宏库一节含长篇 `replace` / `go.work` 说明，宜下沉为参考子节。

约束：文档须继续承载与实现 spec 一致的**技术事实**（provider 契约、`init provider` 命令、examples 为示例调用方、build tag 双文件、对外提交 gen、第三方 `register` 等）；人类可读正文以中文自然叙述为主，RFC 2119 与可检验 MUST 留在 OpenSpec spec，而非堆砌在 author-guide 首屏。

## Goals / Non-Goals

**Goals:**

- 定义并落实固定的 author-guide 信息架构（见 `author-guide` spec）。
- 让宏作者在首屏获得：文档定位、角色分工、动手路径（init → 契约 → mactest）、使用方要点与 README 链接。
- 将 contrib 本地联调放入参考层；**不**在 author-guide 保留 Try 桩族附录（细节由 contrib / 示例承担）。
- 删除 meta 节；语气与 README 一致（使用指引，而非 OpenSpec 变更说明）。

**Non-Goals:**

- 不重写 `README.md`（仅确保互链与职责边界一致）。
- 不新增功能、不改 `go.mod`、不修改 `macro-core` / `macro-codegen` 等运行时 spec 正文。
- 不引入英文版 author-guide。

## Decisions

### 1. 采用「双主路径 + 参考层」结构

| 层级 | 章节（建议 `##` 标题） | 内容 |
|------|------------------------|------|
| L0 导航 | `阅读指引` | 宏库作者 / 宏使用方 / 查参考 三类跳转 |
| L1 共识 | `角色分工` | 各角色**职责**（两列表格即可）；`examples` 为示例调用方 |
| L2 编写路径 | `编写宏库`（含 `###` 子节） | **脚手架 → 契约 → 宏位置与返回字段 → mactest**（子节顺序以利于上手为准） |
| L2 使用方路径 | `宏使用方`（含 `###` 子节） | 主文件 + 生成文件（build tag / //line）、expand 入口、发布前建议 |
| L3 参考 | `参考` | 官方 contrib、本地联调、消费第三方宏库 |

**理由**：与 README 的「扫读 → 动手 → 深挖」一致；编写者与使用方分轨，参考信息后置。

**备选**：维持现状仅润色措辞 —— 拒绝，无法解决结构与 meta 尾注问题。

### 2. 删除「术语澄清（无行为变更）」与 Try 桩族附录

meta 节对读者无操作价值，已删除。Try 桩名/k/error 位置等属于 `go-macro-contrib/try` 的细节，不放在宏作者指南正文；官方宏库表保留 `syntax-try` 模块路径即可。

### 3. 人类文档 vs 规范文档

| 层面 | author-guide（人类） | OpenSpec `author-guide` spec（可检验） |
|------|----------------------|----------------------------------------|
| 用语 | 自然中文、「主文件 + 生成文件」 | 可引用与 `macro-codegen` 一致的方案 C **语义** |
| 代号 | 不出现方案 C、首版、Site（可用「宏出现的位置」） | 允许约束语义等价 |
| 角色表 | 只写「职责」 | 不要求「不负责」列 |
| RFC 2119 | 尽量避免 | spec 中保留 MUST/SHOULD |

### 4. 与 README 的分工（更新 project-readme 边界）

| 文档 | 职责 |
|------|------|
| README | 使用方快速上手、命令表、gopls、链至 author-guide |
| author-guide | provider 契约、build tag 布局要点、init provider、mactest、发布建议与 contrib 参考 |

README 不复制 author-guide 长段；author-guide 首段与阅读指引链回 README 快速上手。

### 5. 官方宏库参考节拆分

`replace`、`go.work`、`GOWORK=off` 等放入 `### 本地联调`；主节保留模块路径与 syntax-id。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 重组后外部链接锚点失效 | 保留语义相近的 `##` 标题；重大改名可记入 CHANGELOG |
| 润色时遗漏 normative 事实 | 对照 `macro-codegen`、`macro-core`、`macro-repo-layout` 与 `author-guide` spec 核对 |
| 与 README 内容重复 | 宏使用方操作步骤以 README 为主，author-guide 只保留布局与入口要点 |

## Migration Plan

1. 按 `tasks.md` 与 `author-guide` spec 重组 `docs/author-guide.md`。
2. 自旧稿迁移技术事实，删除 meta 节与 Try 附录；统一语气。
3. 同步更新本变更的 `author-guide` / `project-readme` delta spec，使归档要求与终稿一致。
4. 无运行时回滚依赖；文档不满意可 revert 单文件。

## Open Questions

（无。终稿结构已与 `author-guide` spec 对齐。）
