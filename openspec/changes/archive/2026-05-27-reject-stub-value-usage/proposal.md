# Proposal: reject-stub-value-usage

## Why

宏语法桩在设计上不是一等函数值，而是调用点上的语法标记。当前 expand 仅识别 `pkg.Stub(...)` 直调，对「把桩当参数传递、赋值、反射」等用法静默忽略，容易在宏主文件中留下会在运行时 hit 桩 `panic` 的代码。应在 expand 期统一报错，与 provider 语义非法（如 `TryExpand` 非法 Site）的失败方式一致。

## What

- 在 `internal/expander` 增加宏主文件扫描：对已注册桩的非直调引用 expand 失败。
- 更新 `macro-expander` 与 `author-guide` 规范。
- 补充 `recognize` / expand 集成测试。

## Capabilities

- `macro-expander` — 引擎级值用法校验（normative）
- `author-guide` — 宏作者/使用方文档要点（与 expand 行为对齐）

## Boundaries（已定）

| 议题 | 决策 |
|------|------|
| 特殊场景 | 一视同仁，不为 provider / 死代码 / 反射等开豁免 |
| 未 link | 注册表无桩则不做法向值用法校验 |
| 死代码 | 仍报错，不做可达性分析 |
| shadow | 与现有识别一致：仅 provider 包级已注册桩才校验 |

## Non-goals

- 不通过数据流分析拦截「先赋给变量再 `fn(...)`」且赋值行未出现桩符号的间接调用（首版依赖「桩符号出现在非 callee 位置即失败」）。
- 不修改 contrib provider 的载荷/Site 语义。
- 不对未 link 的 import 做值用法诊断。

## Success criteria

- `go test ./internal/expander/...` 覆盖列出的 spec 场景。
- `openspec validate reject-stub-value-usage` 通过。
