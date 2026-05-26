## MODIFIED Requirements

### Requirement: contrib/register 注册官方 Expander

contrib MUST 提供 `contrib/register` 包，在其 `init` 中调用 `macro/expandtool.Register`，注册 `contrib/inline` 与 `contrib/try` 的 import path 及对应 `InlineExpand`、`TryExpand`。

本仓库 **examples** module 内的参考 `cmd/macroexpand` MUST 通过 blank import `_ "github.com/arcane-craft/go-macro/contrib/register"` 启用官方宏库 link。宏使用方在项目内自建等价 `cmd/macroexpand` 时 MAY blank import 同一 `contrib/register`（及所需其它 `register` 包）。contrib MUST NOT 在根 module 的 `go.mod` 中被 require。

contrib MUST NOT 提供 `Main`/`Run` 作为用户 expand 入口（该职责在 `macro/expandtool`；可执行入口由宏调用方项目承载）。

#### Scenario: blank import 后 Registered 含官方库

- **WHEN** 进程 blank import `contrib/register`
- **THEN** `expandtool.Registered()` MUST 包含 inline 与 try 的 import path 条目
