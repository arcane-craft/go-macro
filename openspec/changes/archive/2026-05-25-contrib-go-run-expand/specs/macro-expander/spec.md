## MODIFIED Requirements

### Requirement: Provider 激活与 Expander 链接

`ExpandPackages(patterns, linked map[string]macro.Expander)` MUST 对每个待展开包：

1. 收集宏主文件所在包的 **import 路径集合**；
2. 取 **`linked` 的 key 与 import 集合的交集** 为候选宏库路径；
3. 对每个候选路径：解析 provider AST（含 `//macro:`）、注册 panic 桩，并绑定 `linked[path]` 的 `Expander`。

引擎 MUST NOT 维护官方宏库目录，MUST NOT 在识别或 splice 逻辑中对任何 `syntax-id` 硬编码分支。

根 module 的 `internal/expander` 包及其测试 MUST NOT import contrib。对外展开入口为 `macro/expandtool`（不导出 expander 包路径给宏使用方）。

#### Scenario: 未 import 的宏库不 link

- **WHEN** `linked` 含 `contrib/try`，但宏主文件未 import `contrib/try`
- **THEN** 该包展开时 MUST NOT 注册 `syntax-try`

#### Scenario: 已 import 但未 link 则展开失败

- **WHEN** 宏主文件 import `contrib/try` 并调用 `Try(...)`，但 expand 工具传入的 `linked` 为空或不包含该 path
- **THEN** 展开 MUST 失败（未知 stub 或未注册），且 MUST NOT 静默跳过

#### Scenario: 仅 link 已 import 的子集

- **WHEN** 宏主文件 import `contrib/inline` 与 `contrib/try`，但 `linked` 仅含 `contrib/inline`
- **THEN** 该包 MUST 仅注册 `syntax-inline`；对 `Try(...)` 调用 MUST 展开失败
