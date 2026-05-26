## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: contrib 依赖 go-macro 核心版本

`go-macro-contrib` 的 `go.mod` MUST `require` 已发布的 `github.com/arcane-craft/go-macro` 版本；README MUST 注明最低兼容核心版本。本地联调时仓库 SHOULD 位于 `go-macro` 同级目录 `../go-macro-contrib`，并 MAY 使用 `replace github.com/arcane-craft/go-macro => ../go-macro`。

#### Scenario: 独立仓可解析核心依赖

- **WHEN** 在仅 clone `go-macro-contrib` 且已 `go get` 兼容的 `go-macro` 版本时执行 `go test ./...`
- **THEN** MUST 成功解析 `github.com/arcane-craft/go-macro` 模块依赖
