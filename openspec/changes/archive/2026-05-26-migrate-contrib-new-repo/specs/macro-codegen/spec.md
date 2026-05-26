## MODIFIED Requirements

### Requirement: go generate 集成

工具链 MUST 支持在宏主文件中通过**一行** generate 触发 expand，**无需**用户项目内 `tools/macroexpand`。触发的 expand 进程 MUST 通过 blank import 所需 `register` 包并调用 `macro/expandtool.Main()`（或与 `Run` 等价的接线模式）。

文档与快速上手 **RECOMMENDED** 使用：

```go
//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .
```

（整模块可用 `./...`；按包展开用 `.`。）

上述 RECOMMENDED 命令在本仓库中编译并运行 examples module 下的参考 `cmd/macroexpand`（其内部 blank import `github.com/arcane-craft/go-macro-contrib/register`）。宏使用方 MAY 将 generate 改为 `go run <本项目>/cmd/macroexpand` 等等价入口。

#### Scenario: generate 零项目 expand 文件

- **WHEN** 用户仅使用官方宏库（`go-macro-contrib`），宏主文件含 RECOMMENDED 上述 generate（或等价自建入口的 generate），且项目内无 `tools/macroexpand`
- **THEN** `go generate` MUST 成功写回 `*_macro_gen.go`

#### Scenario: 按包 generate

- **WHEN** 宏主文件含 `//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .`（或指向本项目等价入口的 generate）
- **THEN** MUST 仅展开该 generate 所在包（或指令指定的 patterns）

### Requirement: examples 参考 expand 入口（本仓库）

本仓库 MUST 在 **examples** 子 module 提供参考实现 `cmd/macroexpand`（路径 `github.com/arcane-craft/go-macro/examples/cmd/macroexpand`）。该实现 MUST 仅 blank import 所需 `register` 包（含 `github.com/arcane-craft/go-macro-contrib/register`）并调用 `expandtool.Main()`，MUST NOT 包含其它业务逻辑。

根 module MUST NOT 包含 `cmd/macroexpand`（布局细节以 `macro-repo-layout` 为准）。

该路径为宏使用方 **RECOMMENDED** 快速上手命令，MUST NOT 被解释为唯一允许的 expand 入口。

#### Scenario: 参考入口与 expandtool Main 等价

- **WHEN** 用户 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .` 且进程已 link 所需 `register`
- **THEN** 行为 MUST 与在同一进程内调用 `expandtool.Main()` 一致

#### Scenario: 调用方自建等价入口

- **WHEN** 宏使用方在项目内提供等价 `cmd/macroexpand`（blank import 所需 register 并调用 `expandtool.Main()`）
- **THEN** expand 行为 MUST 与使用本仓库 examples 参考入口等价（就 expand 语义而言）

### Requirement: 消费第三方宏库的附录路径

当宏使用方除 `go-macro-contrib` 外还依赖其它带 `register` 子包的宏库时，文档 MAY 说明：复制 `examples/cmd/macroexpand` 为项目内 `cmd/macroexpand` 并**仅**追加 blank import 该宏库的 `register` 包，仍调用 `expandtool.Main()`。该路径 MUST NOT 作为默认快速上手内容。

#### Scenario: 附录不增加宏作者负担

- **WHEN** 第三方宏作者按脚手架发布 `register` 子包
- **THEN** 宏作者 MUST NOT 需要维护 expand 二进制；链接责任在使用方可选 cmd 或未来框架扩展
