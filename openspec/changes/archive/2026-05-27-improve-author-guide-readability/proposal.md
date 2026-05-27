## Why

`docs/author-guide.md` 在多次 spec 对齐后技术事实完整，但呈现方式仍偏「规范摘录」：章节扁平、要点堆砌、存在对终端读者无操作价值的 meta 尾注（如「术语澄清（无行为变更）」），且与已优化为自然中文叙述的 `README.md` 风格不一致。宏作者在 30 秒内难以建立「我是谁 → 先做什么 → 契约细节去哪查」的阅读路径。需要在**不改动**框架运行时语义与 `macro-core` / `macro-codegen` 等实现契约的前提下，重组叙述结构并提升人类可读性。

## What Changes

- 重组 `docs/author-guide.md` 章节：阅读指引 → 角色分工（仅列各角色职责）→ 编写宏库（**脚手架优先**，再契约、宏位置与返回字段表、mactest）→ 宏使用方（主文件 + 生成文件的 build tag、expand 入口、发布前建议）→ 参考（官方 contrib、本地联调、第三方 register）。
- 删除或合并仅描述 spec 对齐过程的 meta 节（「术语澄清（无行为变更）」）；expand 归属等有用信息并入「角色分工」与「宏使用方」。
- 将 dense bullet 改写为自然中文；保留 normative 技术事实（provider 签名、`ExpandResult`、`EnclosingFunc`、//line 等），**不在正文**使用内部设计代号（「方案 C」「首版」）或读者无需知的附录（Try 桩族表）；`macro-codegen` 中的方案 C 语义仍以自然语言描述。
- 拆分官方宏库的本地联调为 `### 本地联调`，避免打断主路径。
- **不**修改 expand 运行时语义、模块路径或 `examples` 参考实现；通过新增 `author-guide` spec 与 `project-readme` delta 对齐**文档层**要求（含人类可读叙述风格）。

## Capabilities

### New Capabilities

- `author-guide`：定义 `docs/author-guide.md` 的信息架构、必备章节顺序、主/次内容分层、叙述风格（含避免 spec/plan 腔与内部代号），以及与 `README.md` 的职责边界。

### Modified Capabilities

- `project-readme`：更新「与作者指南的职责边界」要求，明确 author-guide 须具备独立可读结构（阅读指引、角色分工、编写/使用方分节），README 不复制其长契约正文。

## Impact

- 受影响文件：`docs/author-guide.md`（主交付物）。
- 归档时合并 delta：`author-guide`（新增）、`project-readme`（MODIFIED）。
- **无** expand 引擎、codegen、模块布局或 API 的运行时行为变更。
