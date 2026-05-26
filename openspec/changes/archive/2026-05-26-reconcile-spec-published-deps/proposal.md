## Why

`go-macro` 与 `go-macro-contrib` 已在 `v0.1.0` / `v0.1.1` 打 tag 并改为通过版本化 `require` 消费依赖；`go-macro` 仓库随后移除了根级 `go.work`，`examples/go.mod` 也不再提交对 contrib 的 `replace`。当前 `macro-repo-layout` 与 `macro-contrib` 仍按迁移期「必须 `go.work` + 本地 `replace` 联调」叙事，与仓库实际状态和发布消费路径不一致，导致 spec 评审与实现对照时出现假冲突。

## What Changes

- 修订 `macro-repo-layout` 中「本地开发 workspace」：不再 **MUST** 提供根 `go.work`；明确根 module 与 `examples` module 的分模块测试方式；`examples/go.mod` 以已发布版本 `require` 为提交态常态，contrib 本地 `replace` 降为文档化可选联调手段。
- 修订 `macro-contrib` 中「contrib 依赖 go-macro 核心版本」：已提交 `go.mod` 以发布 tag 为准、不含 `replace`；README 须与 pin 版本一致；本地 `replace` 仅作为开发者本地 overlay，不要求写入仓库。
- 对齐 README、`CHANGELOG`、`docs/author-guide.md` 中与上述 spec 冲突的表述（文档任务，无运行时/API 变更）。

## Capabilities

### New Capabilities

（无。）

### Modified Capabilities

- `macro-repo-layout`：本地开发与双 module 测试、`examples` 依赖消费方式（发布 `require` vs 可选 `replace`）。
- `macro-contrib`：contrib 仓对已发布 `go-macro` 的依赖表述与本地联调边界。

## Impact

- 受影响规格：`openspec/specs/macro-repo-layout/spec.md`、`openspec/specs/macro-contrib/spec.md`。
- **无** Go 代码、模块路径、expand 行为或 tag 策略变更。
- 文档（README / CHANGELOG / author-guide）须与归档后的 spec 一致；实现（`go.mod`、`go.work` 缺失）保持不变。
