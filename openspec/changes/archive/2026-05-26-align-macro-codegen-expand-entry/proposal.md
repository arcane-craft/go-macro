## Why

`clarify-expand-entry-ownership` 已在 `macro-repo-layout` 与 `macro-core` 中明确：框架提供 `expandtool` 能力，宏调用方项目承载可执行 expand 入口，`examples/cmd/macroexpand` 为推荐参考实现而非唯一路径。`macro-codegen`（及 `macro-contrib` 一处）仍使用「框架 macroexpand」「默认入口」「MUST 编译运行 examples 二进制」等旧措辞，与已归档决策不一致，易在实现与文档评审时产生回退误解。

## What Changes

- 修订 `macro-codegen` 中 `go generate 集成`、`仅展开当前主模块`、`init provider` 等 requirement：规范层约束接线模式，文档层保留 RECOMMENDED 的 examples 一行。
- 将 `框架 macroexpand（examples module）` 重表述为「本仓库 examples 参考 expand 入口」，去除「框架拥有 macroexpand」语义。
- 微调 `幂等展开` 等仍硬编码 examples 路径的 requirement，补充「或等价 expand 入口」。
- 修订 `macro-contrib` 中 `contrib/register` requirement：明确 examples 参考实现 MUST  blank import `contrib/register`，而非暗示所有使用方必须用 examples 路径。
- 对齐 README / 作者指南中与 codegen 相关的残留表述（若与 spec 重复）。

## Capabilities

### New Capabilities

- （无）

### Modified Capabilities

- `macro-codegen`: expand 入口职责与 MUST/RECOMMENDED 分层；重命名/替换「框架 macroexpand」requirement。
- `macro-contrib`: `contrib/register` 与 examples 参考接线的措辞。

## Impact

- 受影响规格：`openspec/specs/macro-codegen/spec.md`、`openspec/specs/macro-contrib/spec.md`。
- 无 API 或运行时行为变更；推荐 `go run .../examples/cmd/macroexpand` 命令不变。
- 依赖方：仅阅读 spec 或文档的维护者；无 breaking import 路径变更。
