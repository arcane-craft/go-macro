## REMOVED Requirements

### Requirement: 框架 macroexpand（examples module）

**Reason**: 与 `macro-repo-layout` / `macro-core` 已澄清的「框架提供 expandtool、调用方承载入口、examples 为参考实现」冲突；「框架 macroexpand」命名误导。

**Migration**: 由 **「examples 参考 expand 入口（本仓库）」** requirement 承接本仓库内参考 `cmd/macroexpand` 的存在性与行为等价义务；调用方自建入口见 `go generate 集成` 与 `消费第三方宏库的附录路径`。

## ADDED Requirements

### Requirement: examples 参考 expand 入口（本仓库）

本仓库 MUST 在 **examples** 子 module 提供参考实现 `cmd/macroexpand`（路径 `github.com/arcane-craft/go-macro/examples/cmd/macroexpand`）。该实现 MUST 仅 blank import 所需 `register` 包（含 `contrib/register`）并调用 `expandtool.Main()`，MUST NOT 包含其它业务逻辑。

根 module MUST NOT 包含 `cmd/macroexpand`（布局细节以 `macro-repo-layout` 为准）。

该路径为宏使用方 **RECOMMENDED** 快速上手命令，MUST NOT 被解释为唯一允许的 expand 入口。

#### Scenario: 参考入口与 expandtool Main 等价

- **WHEN** 用户 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .` 且进程已 link 所需 `register`
- **THEN** 行为 MUST 与在同一进程内调用 `expandtool.Main()` 一致

#### Scenario: 调用方自建等价入口

- **WHEN** 宏使用方在项目内提供等价 `cmd/macroexpand`（blank import 所需 register 并调用 `expandtool.Main()`）
- **THEN** expand 行为 MUST 与使用本仓库 examples 参考入口等价（就 expand 语义而言）

## MODIFIED Requirements

### Requirement: go generate 集成

工具链 MUST 支持在宏主文件中通过**一行** generate 触发 expand，**无需**用户项目内 `tools/macroexpand`。触发的 expand 进程 MUST 通过 blank import 所需 `register` 包并调用 `macro/expandtool.Main()`（或与 `Run` 等价的接线模式）。

文档与快速上手 **RECOMMENDED** 使用：

```go
//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .
```

（整模块可用 `./...`；按包展开用 `.`。）

上述 RECOMMENDED 命令在本仓库中编译并运行 examples module 下的参考 `cmd/macroexpand`（其内部 blank import `contrib/register`）。宏使用方 MAY 将 generate 改为 `go run <本项目>/cmd/macroexpand` 等等价入口。

#### Scenario: generate 零项目 expand 文件

- **WHEN** 用户仅使用 contrib 官方宏库，宏主文件含 RECOMMENDED 上述 generate（或等价自建入口的 generate），且项目内无 `tools/macroexpand`
- **THEN** `go generate` MUST 成功写回 `*_macro_gen.go`

#### Scenario: 按包 generate

- **WHEN** 宏主文件含 `//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .`（或指向本项目等价入口的 generate）
- **THEN** MUST 仅展开该 generate 所在包（或指令指定的 patterns）

### Requirement: 幂等展开

对同一输入重复执行同一 expand 入口（RECOMMENDED 为 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand`，或调用方自建等价入口），生成文件 MUST 一致（时间戳除外）。

#### Scenario: 重复执行 expand

- **WHEN** 连续两次相同 expand 且主文件未变
- **THEN** `foo_macro_gen.go` MUST 一致

### Requirement: 仅展开当前主模块

由 `expandtool` 驱动的 expand 入口进程 MUST 仅处理**调用方所在**主 module 内包，MUST NOT 写 module cache。

#### Scenario: 本模块 expand

- **WHEN** 于某 module 根通过 expand 入口执行 expand（RECOMMENDED：`go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand ./...`）
- **THEN** MUST 仅更新该 module 内 `*_macro_gen.go`

### Requirement: init provider 生成 register 而非 expand 工具

`go tool macro init provider` MUST 为**宏作者**生成 provider 骨架，含 `register/register.go`：在 `init` 中 `expandtool.Register(<module>/provider/import/path>, ProviderExpand)`。

MUST NOT 生成 `tools/macroexpand` 或要求宏作者实现 expand main。README MUST 说明宏**使用方**须承载 expand 入口（register + `expandtool.Main()`），并 **RECOMMENDED** 提供 `examples/cmd/macroexpand` 的 generate 一行作为快速上手模板。

#### Scenario: 宏作者无 expand main 义务

- **WHEN** 用户执行 `go tool macro init provider mymac`
- **THEN** 输出 MUST 含 `register/register.go` 且 MUST NOT 含 `tools/macroexpand/main.go`
