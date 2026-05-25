## MODIFIED Requirements

### Requirement: go generate 集成

工具链 MUST 支持在宏主文件中通过**一行** generate 触发 expand，**无需**用户项目内 `tools/macroexpand`：

```go
//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .
```

（整模块可用 `./...`；按包展开用 `.`。）

该命令 MUST 编译并运行 **examples module** 下的 `cmd/macroexpand`，其内部 blank import `contrib/register` 并调用 `macro/expandtool.Main()`。

#### Scenario: generate 零项目 expand 文件

- **WHEN** 用户仅使用 contrib 官方宏库，宏主文件含上述 generate，且项目内无 `tools/macroexpand`
- **THEN** `go generate` MUST 成功写回 `*_macro_gen.go`

#### Scenario: 按包 generate

- **WHEN** 宏主文件含 `//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .`
- **THEN** MUST 仅展开该 generate 所在包（或指令指定的 patterns）

### Requirement: 幂等展开

对同一输入重复执行 `go run .../examples/cmd/macroexpand`，生成文件 MUST 一致（时间戳除外）。

#### Scenario: 重复执行 expand

- **WHEN** 连续两次相同 expand 且主文件未变
- **THEN** `foo_macro_gen.go` MUST 一致

### Requirement: 仅展开当前主模块

`examples/cmd/macroexpand` MUST 仅处理**调用方所在**主 module 内包，MUST NOT 写 module cache。

#### Scenario: 本模块 expand

- **WHEN** 于某 module 根执行 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand ./...`
- **THEN** MUST 仅更新该 module 内 `*_macro_gen.go`

## REMOVED Requirements

### Requirement: go tool macro CLI

**Reason**: `expand` 子命令删除；由 `examples/cmd/macroexpand` + `macro/expandtool` 替代。

**Migration**: `//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .`

## ADDED Requirements

### Requirement: 框架 macroexpand（examples module）

项目 MUST 在 **examples** 子 module 提供 `cmd/macroexpand`，作为宏使用方默认 expand 入口。另实现 MUST 仅 blank import `contrib/register` 并调用 `expandtool.Main()`，MUST NOT 包含其它业务逻辑。

根 module MUST NOT 包含 `cmd/macroexpand`。

#### Scenario: 与 expandtool Main 等价

- **WHEN** 用户 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .` 且进程已 link `contrib/register`
- **THEN** 行为 MUST 与在同一进程内调用 `expandtool.Main()` 一致

### Requirement: init provider 生成 register 而非 expand 工具

`go tool macro init provider` MUST 为**宏作者**生成 provider 骨架，含 `register/register.go`：在 `init` 中 `expandtool.Register(<module>/provider/import/path>, ProviderExpand)`。

MUST NOT 生成 `tools/macroexpand` 或要求宏作者实现 expand main。README MUST 指向宏**使用方**使用 `examples/cmd/macroexpand` 的 generate 一行。

#### Scenario: 宏作者无 expand main 义务

- **WHEN** 用户执行 `go tool macro init provider mymac`
- **THEN** 输出 MUST 含 `register/register.go` 且 MUST NOT 含 `tools/macroexpand/main.go`

### Requirement: 消费第三方宏库的附录路径

当宏使用方除 contrib 外还依赖其它带 `register` 子包的宏库时，文档 MAY 说明：复制 `examples/cmd/macroexpand` 为项目内 `cmd/macroexpand` 并**仅**追加 blank import 该宏库的 `register` 包，仍调用 `expandtool.Main()`。该路径 MUST NOT 作为默认快速上手内容。

#### Scenario: 附录不增加宏作者负担

- **WHEN** 第三方宏作者按脚手架发布 `register` 子包
- **THEN** 宏作者 MUST NOT 需要维护 expand 二进制；链接责任在使用方可选 cmd 或未来框架扩展
