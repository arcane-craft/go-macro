# macro-contrib Specification

## Purpose
TBD - created by archiving change contrib-go-run-expand. Update Purpose after archive.
## Requirements
### Requirement: contrib 独立子 module

官方宏库 MUST 在**独立 Git 仓库**中作为根 Go module 发布，module 路径为 `github.com/arcane-craft/go-macro-contrib`。`go-macro` 仓库 MUST NOT 再包含 `contrib/` 目录。

`go-macro` 根 module 的 `internal/expander`、`macro/expandtool` 及根 module 内所有测试 MUST NOT import 官方宏实现包（`inline`、`try`、`register`）。

#### Scenario: expander 不依赖 contrib 实现

- **WHEN** 编译 `go-macro` 根 module 的 `internal/expander` 包
- **THEN** MUST NOT import `github.com/arcane-craft/go-macro-contrib/inline` 或 `.../try`

#### Scenario: 根测试不 import contrib

- **WHEN** 编译 `go-macro` 根 module 任意 `*_test.go`
- **THEN** MUST NOT import `github.com/arcane-craft/go-macro-contrib/...`

### Requirement: 官方宏库路径

官方宏库 MUST 仅通过下列 import 路径提供：

- `github.com/arcane-craft/go-macro-contrib/inline`
- `github.com/arcane-craft/go-macro-contrib/try`

`go-macro` 根 module MUST NOT 再包含 `inline/`、`try/` 或 `contrib/`。

#### Scenario: import 新路径

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro-contrib/try`
- **THEN** MUST 解析到 `go-macro-contrib` 仓库的 `try` 包

### Requirement: contrib/register 注册官方 Expander

`go-macro-contrib` MUST 提供 `register` 包，在其 `init` 中调用 `macro/expandtool.Register`，注册 `.../inline` 与 `.../try` 的 import path 及对应 `InlineExpand`、`TryExpand`。

`go-macro` 仓库 **examples** module 内的参考 `cmd/macroexpand` MUST 通过 blank import `_ "github.com/arcane-craft/go-macro-contrib/register"` 启用官方宏库 link。宏使用方在项目内自建等价 `cmd/macroexpand` 时 MAY blank import 同一 `register`（及所需其它宏库 `register` 包）。`go-macro-contrib` MUST NOT 在 `go-macro` 根 module 的 `go.mod` 中被 require。

`go-macro-contrib` MUST NOT 提供 `Main`/`Run` 作为用户 expand 入口（该职责在 `macro/expandtool`；可执行入口由宏调用方项目承载）。

#### Scenario: blank import 后 Registered 含官方库

- **WHEN** 进程 blank import `github.com/arcane-craft/go-macro-contrib/register`
- **THEN** `expandtool.Registered()` MUST 包含 inline 与 try 的 import path 条目

### Requirement: contrib 独立测试

`go-macro-contrib` 仓库 MUST 具备独立 `go test ./...`，含 inline/try 的 mactest 单测。

#### Scenario: contrib 测试不依赖 go tool macro expand

- **WHEN** 在 `go-macro-contrib` 仓库根执行 `go test ./...`
- **THEN** MUST 通过

### Requirement: contrib 依赖 go-macro 核心版本

`go-macro-contrib` 的**已提交** `go.mod` **MUST** `require` 已发布的 `github.com/arcane-craft/go-macro` 版本（semver tag，非 `v0.0.0` 占位）；**MUST NOT** 在已提交 `go.mod` 中包含指向 sibling 目录的 `replace` 指令。

README **MUST** 注明最低兼容核心版本，且所述版本 **MUST** 与 `go.mod` 中 pin 的 `require` 一致。

本地联调时，contrib 仓库 **SHOULD** 与 `go-macro` 位于同级目录（`go-macro-contrib` 与 `go-macro` 并列）。开发者 **MAY** 在本地向 `go-macro-contrib/go.mod` 添加 `replace github.com/arcane-craft/go-macro => ../go-macro` 以联调未发布的核心变更；该 `replace` **MUST NOT** 作为发布/tag 前提交态的硬性要求。

#### Scenario: 独立仓可解析核心依赖

- **WHEN** 在仅 clone `go-macro-contrib`、已提交 `go.mod` 无 `replace`，且模块代理可解析所 pin 的 `go-macro` tag 时执行 `go test ./...`
- **THEN** MUST 成功解析 `github.com/arcane-craft/go-macro` 模块依赖并通过测试

#### Scenario: 本地 replace 联调核心

- **WHEN** 开发者在 `go-macro-contrib` 本地添加 `replace github.com/arcane-craft/go-macro => ../go-macro` 后执行 `go test ./...`
- **THEN** MUST 使用 sibling 核心源码解析依赖（联调行为；不要求写入已提交 `go.mod`）

