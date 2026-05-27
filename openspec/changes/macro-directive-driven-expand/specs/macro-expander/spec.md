## MODIFIED Requirements

### Requirement: 展开器函数签名约定

每个 `Expander` MUST 为 `func(Context, *ast.CallExpr) (ExpandResult, error)`。Expander 函数 MUST 在其 doc 注释中含 `//macro: <syntax-id>`。注册表 MUST 将该函数绑定为对应 syntax-id 的展开器（并与同 syntax-id 的桩函数关联）。

#### Scenario: 注册表绑定 Expander 签名

- **WHEN** provider 包中 `InlineExpand` 的 doc 含 `//macro: syntax-inline` 且签名符合 `Expander`
- **THEN** 注册表 MUST 将 `syntax-inline` 的展开实现绑定到 `linked` 提供的 `InlineExpand` 函数值

### Requirement: Provider 激活与 Expander 链接

`ExpandPackages(patterns, linked map[string]macro.Expander)` MUST 对每个待展开包：

1. 收集宏主文件所在包的 **import 路径集合**；
2. 取 **`linked` 的 key 与 import 集合的交集** 为候选宏库路径；
3. 对每个候选路径：解析 provider AST，从**各函数 doc** 读取 `//macro:`，登记桩与 syntax-id，并绑定 `linked[path]` 的 `Expander`。

引擎 MUST NOT 维护官方宏库目录，MUST NOT 在识别或 splice 逻辑中对任何 `syntax-id` 硬编码分支。

`go-macro` 根 module 的 `internal/expander` 包及其测试 MUST NOT import `go-macro-contrib`。对外展开入口为 `macro/expandtool` 与 `cmd/macro expand`。

#### Scenario: 未 import 的宏库不 link

- **WHEN** `linked` 含某 provider path，但宏主文件未 import 该 path
- **THEN** 该包展开时 MUST NOT 注册该 provider 的桩

#### Scenario: 已 import 但未 link 则展开失败

- **WHEN** 宏主文件 import 某 provider 并调用其桩，但 `linked` 不包含该 path
- **THEN** 展开 MUST 失败（未知 stub 或未注册），且 MUST NOT 静默跳过

#### Scenario: 仅 link 已 import 的子集

- **WHEN** 宏主文件 import `inline` 与 `try` 两个 provider，但 `linked` 仅含 `inline` path
- **THEN** 该包 MUST 仅注册 `inline` 的桩；对 `try` 桩的调用 MUST 展开失败

#### Scenario: 识别使用 importPath 与桩名

- **WHEN** 两个 provider 均含名为 `Macro` 的桩且各自 doc 含 `//macro:`，宏主文件分别通过 `a.Macro` 与 `b.Macro` 调用
- **THEN** 引擎 MUST 按各自 import path 分发到对应 Expander，且 MUST NOT 因桩名相同而错绑
