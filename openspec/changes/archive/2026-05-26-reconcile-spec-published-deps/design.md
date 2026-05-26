## Context

迁移 change `2026-05-26-migrate-contrib-new-repo` 归档时，规范假定：

- `go-macro` 根目录 **MUST** 保留 `go.work`（`use` 根 + `./examples`）。
- `examples/go.mod` **SHOULD** 提交 `replace github.com/arcane-craft/go-macro-contrib => ../go-macro-contrib`。
- `go-macro-contrib` 开发期可在 `go.mod` 中 **MAY** `replace` 本地 `../go-macro`。

发 tag 后的实际仓库状态（`6cd2d0b` 之后提交）为：

| 项 | 当前实现 |
|----|----------|
| `go-macro` 根 `go.work` | 已删除（`b59b4eb`） |
| `examples/go.mod` | `require go-macro v0.1.0`、`go-macro-contrib v0.1.1`；仅 `replace go-macro => ../` |
| `go-macro-contrib/go.mod` | `require go-macro v0.1.0`；无 `replace`（`ee27929`） |

二者在「发布消费」上合理，但与现行 spec 的 MUST/SHOULD 冲突。本 change **只改规范叙事**，不回滚实现。

## Goals / Non-Goals

**Goals:**

- 使 `macro-repo-layout`、`macro-contrib` 准确描述「双仓已发布、examples 版本化依赖、分 module 测试」的现状。
- 保留本地双仓联调能力，但表述为 **文档化 / 本地 overlay**（`replace`、可选 `go.work`），而非提交态 MUST。
- 为 CI 与贡献者提供可执行的测试场景（根 `GOWORK=off`、examples 目录内 `go test`）。

**Non-Goals:**

- 不恢复或删除 `go.work`、不修改任何 `go.mod` / `go.sum`。
- 不改变 import 路径、register 模型、RECOMMENDED generate 命令。
- 不规定具体 semver 矩阵（仅要求「兼容已发布 tag」；当前 pin 作为示例写在文档任务中）。

## Decisions

### D1: 根 `go.work` 从 MUST 降为 MAY

**选择**：`go-macro` 根 **MAY** 提供 `go.work`（`use` 为 `.` 与 `./examples`）；**MUST NOT** 要求仓库必须包含 `go.work` 才能通过规范。

**理由**：发 tag 后移除 `go.work` 可避免将 sibling `go-macro-contrib` 误纳入 workspace；根与 examples 边界更清晰。

**备选**：恢复 `go.work` 并改实现 —— 用户明确不改实现，故不采用。

### D2: `examples/go.mod` 提交态以版本化 `require` 为准

**选择**：已提交的 `examples/go.mod` **MUST** 通过 `require` 引用已发布的 `go-macro` 与 `go-macro-contrib` tag；**MAY** 保留 `replace github.com/arcane-craft/go-macro => ../` 以便在本仓开发核心时联调 examples。

对 contrib：**MAY** 在本地临时添加 `replace github.com/arcane-craft/go-macro-contrib => ../go-macro-contrib`（README 说明路径为相对 `examples/` 的 `../go-macro-contrib`），但 **MUST NOT** 要求该 `replace` 必须出现在已提交 `go.mod` 中。

**理由**：与 `go get` 消费路径一致；本地并行改 contrib 仍可通过未提交的 replace 或 `go work` 完成。

### D3: 双 module 测试场景拆分

**选择**：

- 根：`GOWORK=off go test ./...` **MUST** 通过（仅根 module）。
- examples：在 `examples/` 下 `go test ./...` **MUST** 通过（依赖已 `require` 的 contrib 版本或开发者本地 replace）。

删除「根目录 `go test ./...` 默认经 `go.work` 覆盖两 module」的 MUST 场景。

### D4: `go-macro-contrib` 已提交 `go.mod` 不含 `replace`

**选择**：contrib 仓库已提交的 `go.mod` **MUST** `require` 已发布 `go-macro` 版本且 **MUST NOT** 包含指向 sibling 的 `replace`。本地联调时开发者 **MAY** 本地添加 `replace github.com/arcane-craft/go-macro => ../go-macro`（不提交）。

**理由**：与 `ee27929` 及模块代理消费一致；README 仍描述 sibling 布局。

### D5: 文档与 spec 同步为 apply 任务

README / CHANGELOG / author-guide 中「已含 contrib replace」「根 go.work 仅根+examples」等句子的修正列入 `tasks.md`，不写入 requirement 正文。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 仅 clone `go-macro` 且未 `go get` contrib 时 examples 测试失败 | spec 场景明确 examples 依赖已发布 tag；README 说明 `go get` 或本地 replace |
| 贡献者习惯根目录一条 `go test ./...` 覆盖全仓 | 文档与 CI 写明两步；可选本地自建 `go.work`（MAY） |
| spec 与 `go-macro-contrib` README 仍写 v0.0.0 | tasks 中更新 contrib README（若在本 change apply 时一并改文档） |

## Migration Plan

1. 归档本 change 的 spec delta 至 `openspec/specs/`。
2. 按 `tasks.md` 更新 `go-macro` 文档（及可选 `go-macro-contrib` README 版本句）。
3. 无代码回滚、无 tag 重发。

**回滚**：恢复 spec 旧条文即可；实现无需变动。

## Open Questions

- 是否在 `go-macro` CI 中显式分两步跑根与 examples 测试（实现任务，本 change 不强制，可在 follow-up 记录）。
