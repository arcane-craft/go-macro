## Why

当前规格对 `examples/cmd/macroexpand` 的描述容易被理解为“框架内置且唯一的通用展开工具”，与目标职责边界不一致。需要明确：框架提供 `macro/expandtool` 能力，宏调用方项目负责承载具体的 expand 入口；`examples` 只是示例调用方项目。

## What Changes

- 澄清职责边界：框架负责 expand 协议与库能力（`expandtool`），调用方负责具体可执行入口（如项目内 `cmd/macroexpand`）。
- 将 `examples/cmd/macroexpand` 的语义从“官方通用入口”调整为“推荐示例实现”，避免被解读为唯一强约束路径。
- 保留文档层面的推荐用法（例如基于 examples 的一行 generate），但与规范层 `MUST` 约束分层表达。
- 明确在 `contrib` 独立仓库等未来演进下，仍可通过“register + expandtool.Main()` 等价模式”满足规范。

## Capabilities

### New Capabilities
- （无）

### Modified Capabilities
- `macro-repo-layout`: 调整对 `examples/cmd/macroexpand` 的定位与措辞，明确其示例属性与规范/推荐分层。
- `macro-core`: 明确框架职责边界为提供 expand 能力与接线模式，不强制调用方绑定唯一入口路径。

## Impact

- 受影响规格：`openspec/specs/macro-repo-layout/spec.md`、`openspec/specs/macro-core/spec.md`。
- 受影响文档：README/作者指南中关于 generate 与 macroexpand 的叙述将需与新措辞一致（语义对齐，无功能变更）。
- 代码与 API：无行为变更，主要为规范文本与术语澄清。
