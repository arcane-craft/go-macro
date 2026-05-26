## MODIFIED Requirements

### Requirement: examples 参考 expand 入口（本仓库）

本仓库 MUST 在 **examples** 子 module 提供参考实现 `cmd/macroexpand`（路径 `github.com/arcane-craft/go-macro/examples/cmd/macroexpand`）。该实现 MUST 仅 blank import 所需 `register` 包（含 `github.com/arcane-craft/go-macro-contrib/register`）并调用 `expandtool.Main()`，MUST NOT 包含其它业务逻辑。

根 module MUST NOT 包含 `cmd/macroexpand`（布局细节以 `macro-repo-layout` 为准）。

`go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand` MUST 作为 **examples 子 module 内** 文档与示例源码的 RECOMMENDED expand 命令。该路径 MUST NOT 被解释为宏使用方项目的全局默认入口，亦 MUST NOT 被解释为唯一允许的 expand 入口。

#### Scenario: 参考入口与 expandtool Main 等价

- **WHEN** 用户 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .` 且进程已 link 所需 `register`
- **THEN** 行为 MUST 与在同一进程内调用 `expandtool.Main()` 一致

#### Scenario: 调用方自建等价入口

- **WHEN** 宏使用方在项目内提供等价 `cmd/macroexpand`（blank import 所需 register 并调用 `expandtool.Main()`）
- **THEN** expand 行为 MUST 与使用本仓库 examples 参考入口等价（就 expand 语义而言）

### Requirement: init provider 生成 register 而非 expand 工具

`go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>`（或与 `cmd/macro` 等价的已发布 module 路径）MUST 为**宏作者**生成 provider 骨架，含 `register/register.go`：在 `init` 中 `expandtool.Register(<module>/provider/import/path>, ProviderExpand)`。

MUST NOT 生成 `tools/macroexpand` 或要求宏作者实现 expand main。README MUST 说明宏**使用方**须在使用宏的项目内承载 expand 入口（`register` + `expandtool.Main()`）。README `快速上手` MUST 以创建项目内 `cmd/macroexpand` 为首要步骤，并 RECOMMENDED 在使用宏的文件中使用 `//go:generate go run ./cmd/macroexpand .`（或模块下等价路径）。README MUST 通过链接或文字引用 `examples/cmd/macroexpand` 与 `examples/readfile` 作为对照示例，MUST NOT 将 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand` 表述为宏使用方项目的默认长期依赖命令。

#### Scenario: 宏作者无 expand main 义务

- **WHEN** 用户执行 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac`
- **THEN** 输出 MUST 含 `register/register.go` 且 MUST NOT 含 `tools/macroexpand/main.go`

#### Scenario: README 快速上手以自建 expand 为先

- **WHEN** 读者按 README `快速上手` 顺序操作
- **THEN** MUST 在遇到 `//go:generate` 示例之前看到创建项目内 `cmd/macroexpand` 的说明
