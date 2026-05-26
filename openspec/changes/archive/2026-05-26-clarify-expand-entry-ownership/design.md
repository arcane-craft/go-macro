## Context

现有规格在多处将 `examples/cmd/macroexpand` 写成“官方 expand 二进制入口”，实践上可作为推荐默认，但语义上容易被理解为框架唯一且强绑定的执行入口。该表述与“框架提供库能力、调用方项目承载具体命令入口”的职责边界存在歧义，且会影响后续 `contrib` 独立仓库或多发行形态下的演进空间。

## Goals / Non-Goals

**Goals:**
- 将规范约束聚焦于能力与模式：`register + expandtool.Main()` 的等价接线能力。
- 明确 `examples` 角色为示例调用方项目，而非框架内部唯一工具归属。
- 在不改变运行时行为和 API 的前提下，消除“唯一入口路径”误解。

**Non-Goals:**
- 不修改 `macro/expandtool` API。
- 不变更当前示例工程目录结构或 generate 推荐命令。
- 不引入新的 capability，仅修订既有 capability 的 requirement 文本。

## Decisions

### D1: 规范与推荐分层
将规范层（MUST）定义为“调用方项目必须提供或使用等价 expand 入口接线模式”，而文档层（RECOMMENDED）保留以 `examples/cmd/macroexpand` 作为推荐示例。

备选方案：
- 继续将 `examples/cmd/macroexpand` 作为 MUST 的唯一入口路径。  
  放弃原因：限制未来发行形态，且与职责边界不一致。

### D2: 在 `macro-repo-layout` 中重定义 examples 职责措辞
将 examples 的职责由“官方 expand 二进制”改为“示例调用方工程，包含参考 expand 入口实现”。保留其可执行与可复制价值，但不再作为唯一规范路径。

备选方案：
- 仅在 README 澄清，不改 spec。  
  放弃原因：核心歧义发生在规范文本，文档修订不足以约束后续决策。

### D3: 在 `macro-core` 中强调框架边界
将 `expandtool` requirement 明确为：框架提供 `Run/Main/Register` 能力；provider 作者不负责 main；调用方（含 examples）负责选择具体承载命令入口的项目与路径。

备选方案：
- 将入口职责写入 `macro-repo-layout`，不改 `macro-core`。  
  放弃原因：`expandtool` 语义本身属于 `macro-core`，不改会保留跨 spec 不一致。

## Risks / Trade-offs

- [风险] “不唯一入口”会被误读为“任何路径都推荐”。  
  → Mitigation：在 requirement 中明确 MUST（等价能力）与 RECOMMENDED（examples 默认示例）并行。
- [风险] 现有读者已接受“官方入口=examples 路径”。  
  → Mitigation：在迁移说明中标注“语义澄清，无功能变更”。
- [风险] 后续 contrib 拆仓时再次引发术语偏差。  
  → Mitigation：新增关于“调用方承载入口”的场景化 requirement 文本，作为统一锚点。

## Migration Plan

1. 更新 `macro-repo-layout` 与 `macro-core` 的 delta spec。
2. 复核 README 与作者指南中“官方/唯一/通用”措辞，改为“推荐示例/等价模式”。
3. 在下次变更实现阶段按 tasks 执行文档同步，不触发代码行为变更。

## Open Questions

- 是否在后续单独变更中为“调用方自建 `cmd/macroexpand`”补充更完整模板与文档片段？
- 若 `contrib` 未来拆至独立仓库，是否需要在文档层额外给出多发行物下的推荐接线路径矩阵？

## 评审备注

**为何不再将 `examples/cmd/macroexpand` 路径当作唯一入口**

- Go 要求 `Expander` 在 expand 进程编译期 link；框架通过 `expandtool.Register` 提供能力，但**谁启动该进程**属于宏调用方项目职责。
- `examples` 在本仓库中的角色是「示例调用方」，其 `cmd/macroexpand` 是可复制参考，不是框架内核的一部分。
- 将路径写死为 MUST 会阻碍 contrib 独立仓库、多发行物或调用方自建入口等演进；规范层应约束**接线模式**（register + Main），文档层保留 RECOMMENDED 默认命令。

**边界**

- 本次为语义澄清：推荐 `go run .../examples/cmd/macroexpand` 不变，无 API/行为变更。
- `macro-codegen` 等其它 capability 若仍含「默认入口」旧措辞，可在后续变更中单独对齐。
