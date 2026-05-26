## Context

`macro-repo-layout` / `macro-core` 已完成 expand 入口归属澄清。`macro-codegen` 仍承担「generate 如何触发 expand」的规范，但其中多处将 **examples 路径** 与 **框架义务** 混写，需与上游 capability 对齐，且避免重复定义已在 `macro-repo-layout` 中的布局约束。

## Goals / Non-Goals

**Goals:**

- `macro-codegen` 只规范 codegen 相关行为：generate 触发、写回、幂等、模块边界、文档对使用方的说明。
- expand 入口：MUST = `register` + `expandtool.Main()`（或等价 `Run`）；RECOMMENDED = examples 参考命令与本仓 `examples/cmd/macroexpand` 存在性。
- `macro-contrib` 仅修正 examples 参考接线一句，不扩大 contrib 职责。

**Non-Goals:**

- 不修改 `expandtool` API 或 examples 目录结构。
- 不更改推荐 generate 一行字符串。
- 不重写与 build tag / line / 方案 C 无关的 requirement。

## Decisions

### D1: 删除「框架 macroexpand」requirement 标题

用 **「examples 参考 expand 入口（本仓库）」** 替代，义务限定为：本仓库 examples module 内保留参考 `cmd/macroexpand`，行为与 `expandtool.Main()` 一致；根 module 仍不得含 `cmd/macroexpand`（可引用 `macro-repo-layout`，此处不重复展开）。

### D2: go generate 分层

- MUST：一行 generate 可触发 expand，且无需 `tools/macroexpand`；触发的进程 MUST 满足 register 接线。
- RECOMMENDED：文档与快速上手使用 `go run .../examples/cmd/macroexpand`。
- 删除「该命令 MUST 编译并运行 examples module 下的 cmd」作为唯一路径的表述。

### D3: 行为约束主语抽象化

「仅展开当前主模块」「幂等展开」等 requirement 的主语改为 **expand 入口进程**（由 `expandtool` 驱动），examples 路径仅出现在 RECOMMENDED 或 Scenario 示例中。

### D4: macro-contrib 单句对齐

将「examples module 的 cmd/macroexpand MUST blank import contrib/register」改为「本仓库 examples 参考实现 MUST …」，与「调用方自建 cmd 亦 MAY blank import contrib/register」不冲突。

## Risks / Trade-offs

- [风险] 多处 REMOVED/ADDED 导致 archive 时 diff 较大。  
  → Mitigation：变更范围仅限 2 个 capability，requirement 块完整复制 MODIFIED 内容。
- [风险] 与 `macro-repo-layout` 重复定义 examples 职责。  
  → Mitigation：codegen 侧只写 generate/写回相关义务，布局以 repo-layout 为准。

## Migration Plan

1. 合并 delta 至 `openspec/specs/macro-codegen/spec.md`、`macro-contrib/spec.md`。
2. 抽查 README、`docs/author-guide.md`、`cmd/macro` 脚手架 README 模板是否与 spec 一致。
3. 无代码变更；归档 change。

## Open Questions

- 是否在 `macro-codegen` Purpose 段一并更新（当前仍为 TBD）？本次可顺带写一句 Purpose，或留待专门文档变更。
