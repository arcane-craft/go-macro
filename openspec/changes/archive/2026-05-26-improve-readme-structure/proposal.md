## Why

根目录 `README.md` 在多次 spec 对齐后信息完整，但章节顺序与信息密度不利于首次阅读：快速上手、命令表、contrib、模块路径与术语澄清混在同一层级，读者难以在 30 秒内建立「这是什么 → 我要做什么 → 进阶去哪看」的心智模型。需要在**不改动**技术事实与 OpenSpec 既有 MUST/RECOMMENDED 语义的前提下，重组叙述结构，提升人类可读性。

## What Changes

- 重组 `README.md` 章节：开篇一句话定位 → 受众与阅读路径 → 分步快速上手 → 常用命令 → 进阶链接（作者指南、示例）→ 可选官方宏库与本地联调 → 模块路径等参考信息 → gopls 等工具提示。
- 从 README **移除** 历史 `contrib/` 路径的 BREAKING 迁移表（迁移说明保留在 CHANGELOG）；将本地 `replace` 说明、expand 入口归属等**参考性**内容下沉为次级小节，避免打断主流程。
- 统一术语与语气：README 面向用户使用自然中文；快速上手以项目内 `cmd/macroexpand` 为先，`examples` 仅作对照。
- 文档中的脚手架入口改为 `go run github.com/arcane-craft/go-macro/cmd/macro@latest`（不再以 `go tool macro` 为文档主路径）。
- **不**修改 expand 运行时语义、模块路径或 `examples` 参考实现的存在性；通过 spec delta 对齐 README 与 `macro-codegen` / `macro-core` 的文档层要求。

## Capabilities

### New Capabilities

- `project-readme`：定义根 `README.md` 的信息架构、必备章节顺序、主/次内容分层，以及与 `docs/author-guide.md` 的职责边界。

### Modified Capabilities

- `macro-codegen`：README 快速上手以自建 expand 为先；`examples/cmd/macroexpand` 限定为 examples 模块内 RECOMMENDED；`init provider` 文档入口改为 `go run .../cmd/macro@latest`。
- `macro-core`：`init provider` 脚手架的文档 RECOMMENDED 调用方式为 `go run .../cmd/macro@latest`。

## Impact

- 受影响文件：`README.md`（主交付物）、`docs/author-guide.md`（脚手架命令）、`cmd/macro` usage 文案。
- 归档时合并 delta：`project-readme`（新增）、`macro-codegen`、`macro-core`（MODIFIED）。
- **无** expand 引擎、codegen 或模块布局的运行时行为变更。
