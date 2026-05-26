# macro-contrib Specification

## Purpose
TBD - created by archiving change contrib-go-run-expand. Update Purpose after archive.
## Requirements
### Requirement: contrib 独立子 module

`contrib/` MUST 作为独立 Go module 发布（`github.com/arcane-craft/go-macro/contrib`）。根 module 的 `internal/expander`、`macro/expandtool` 及根 module 内所有测试 MUST NOT import contrib 内的宏实现包（`inline`、`try`、`register`）。

#### Scenario: expander 不依赖 contrib 实现

- **WHEN** 编译根 module 的 `internal/expander` 包（非测试消费方另 module）
- **THEN** MUST NOT import `contrib/inline` 或 `contrib/try`

#### Scenario: 根测试不 import contrib

- **WHEN** 编译根 module 任意 `*_test.go`
- **THEN** MUST NOT import `github.com/arcane-craft/go-macro/contrib/...`

### Requirement: 官方宏库路径

官方宏库 MUST 仅通过下列 import 路径提供：

- `github.com/arcane-craft/go-macro/contrib/inline`
- `github.com/arcane-craft/go-macro/contrib/try`

根 module MUST NOT 再包含 `inline/`、`try/`。

#### Scenario: import 新路径

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro/contrib/try`
- **THEN** MUST 解析到 contrib 子 module 的 `try` 包

### Requirement: contrib/register 注册官方 Expander

contrib MUST 提供 `contrib/register` 包，在其 `init` 中调用 `macro/expandtool.Register`，注册 `contrib/inline` 与 `contrib/try` 的 import path 及对应 `InlineExpand`、`TryExpand`。

本仓库 **examples** module 内的参考 `cmd/macroexpand` MUST 通过 blank import `_ "github.com/arcane-craft/go-macro/contrib/register"` 启用官方宏库 link。宏使用方在项目内自建等价 `cmd/macroexpand` 时 MAY blank import 同一 `contrib/register`（及所需其它 `register` 包）。contrib MUST NOT 在根 module 的 `go.mod` 中被 require。

contrib MUST NOT 提供 `Main`/`Run` 作为用户 expand 入口（该职责在 `macro/expandtool`；可执行入口由宏调用方项目承载）。

#### Scenario: blank import 后 Registered 含官方库

- **WHEN** 进程 blank import `contrib/register`
- **THEN** `expandtool.Registered()` MUST 包含 inline 与 try 的 import path 条目

### Requirement: contrib 独立测试

contrib MUST 具备独立 `go test ./...`，含 inline/try 的 mactest 单测。

#### Scenario: contrib 测试不依赖 go tool macro expand

- **WHEN** 在 contrib 目录执行 `go test ./...`
- **THEN** MUST 通过

