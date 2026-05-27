## Why

官方宏库代码已迁入独立仓库 `go-macro-contrib`，但 OpenSpec 中与之对应的规范（`macro-contrib`、`syntax-inline`、`syntax-try`）仍留在 `go-macro/openspec/specs/`。规范归属与代码仓库不一致，会在修改 inline/try 或 contrib 发布策略时误导维护者、造成双仓重复维护，也不利于 contrib 贡献者在本仓内独立演进规范。

## What Changes

- 在 `go-macro-contrib` 仓库初始化 OpenSpec（`openspec/specs/`），迁入并作为权威来源维护 `macro-contrib`、`syntax-inline`、`syntax-try` 三份 spec。
- 从 `go-macro` 主规范库**删除**上述三份 spec（经本 change 归档后不再保留正文）。
- 更新 `go-macro` 侧仍引用 contrib 行为的 spec（`macro-repo-layout`、`macro-expander`、`macro-core`、`macro-codegen` 等）：保留框架边界与跨仓依赖要求，将 provider 级语义改为引用 contrib 仓库规范，避免重复全文。
- 更新 `author-guide`、`project-readme` 规范：官方宏库细节链至 contrib 仓库 README / OpenSpec，而非在 core 仓 spec 中展开。
- 将原 `syntax-try` 中「Try 端到端示例（readfile）」类要求迁至 `go-macro` 的 `macro-repo-layout`（示例归属 core/examples，行为仍依赖 contrib 包）。

## Capabilities

### New Capabilities

（无。contrib 侧能力名称不变，仅在 `go-macro-contrib` 仓库新建同名 spec 目录。）

### Modified Capabilities

- `macro-contrib`：**REMOVED** 自 `go-macro`（整份 spec 迁至 contrib 仓）。
- `syntax-inline`：**REMOVED** 自 `go-macro`（整份 spec 迁至 contrib 仓）。
- `syntax-try`：**REMOVED** 自 `go-macro`；其中 examples 端到端要求拆分至 `macro-repo-layout`。
- `macro-repo-layout`：新增「contrib 规范外置」与 readfile 端到端示例要求；弱化对 contrib 实现细节的重复描述。
- `macro-expander`：contrib 相关场景改为引用外置规范，保留框架侧 link/import 行为不变。
- `macro-core`：对 `syntax-try` / `syntax-inline` 的交叉引用改为指向 contrib 仓 OpenSpec。
- `macro-codegen`：官方宏库路径说明对齐「规范在 contrib 仓」。
- `author-guide`：参考节指向 contrib 仓库文档与 OpenSpec，而非 core 仓内 contrib spec。
- `project-readme`：官方宏库参考链至 contrib 仓库。

## Impact

- **go-macro**：删除 `openspec/specs/{macro-contrib,syntax-inline,syntax-try}/`；更新其余 spec 与文档中的交叉引用；归档本 change 后主规范集仅描述框架与 examples 边界。
- **go-macro-contrib**：新增 `openspec/` 树（含三份 spec 与可选 `config.yaml` / README 说明）；后续 inline/try 行为变更以 contrib 仓 OpenSpec 为准。
- **工作流**：双仓 PR 可能并存（core 删 spec + contrib 增 spec）；需在 README 中注明规范来源，避免贡献者改错仓库。
