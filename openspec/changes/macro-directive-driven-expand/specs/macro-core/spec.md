## MODIFIED Requirements

### Requirement: ExpandResult 与 Expander 签名

`macro` 包 MUST 定义：

- `ExpandResult` 含可选字段 `Stmts []ast.Stmt`、`Exprs []ast.Expr`、`Expr ast.Expr`（首版 **保留 `Exprs`**，供罕见 return 表达式列表替换；`syntax-try` 在 `SiteReturn` 禁止使用 `Exprs`，见 `syntax-try` spec）
- `Expander func(ctx Context, call *ast.CallExpr) (ExpandResult, error)`

Expander 函数 MUST 在**该函数**的 doc 注释中含 `//macro: <syntax-id>`，且 MUST 符合上述签名。语法桩 MUST 在**各桩函数**的 doc 中含 `//macro: <syntax-id>`。系统 MUST 支持多个 provider 并存（不同 `syntax-id`），规则仅存在于各 `Expand` 实现。

#### Scenario: 表达式宏

- **WHEN** 某宏在 `SiteExpr` 语境仅返回 `ExpandResult{Expr: e}`
- **THEN** 类型与文档 MUST 允许该形式，且引擎仅替换 `CallExpr`

#### Scenario: 语句宏

- **WHEN** `TryExpand` 在 `SiteAssign` 展开
- **THEN** MUST 返回非空 `Stmts` 以替换整条赋值语句

### Requirement: 通用宏注册与查找

系统 MUST 支持：对**宏主文件所在包已 import 的**宏库包，解析各桩函数与 Expander 函数上的 `//macro: <syntax-id>`，并结合 **`linked map[string]macro.Expander`**（由 `cmd/macro expand` 生成的 link 代码调用 `expandtool.Register`，或测试显式传入）构建注册表。`internal/expander` 与 `macro/expandtool` MUST NOT 硬编码任何具体宏库 Expander。

宏调用识别 MUST 使用 `(importPath, stubName)` 在注册表中查找；同一 stub 名在不同 import path 下 MUST NOT 相互覆盖。

#### Scenario: 仅注册已 import 且已 link 的宏库

- **WHEN** 生成 link 已 `expandtool.Register` 某宏库，但宏主文件未 import 该包
- **THEN** 注册表 MUST NOT 包含该包的桩

#### Scenario: 注册多桩名到同一 syntax-id

- **WHEN** 宏库包中桩 `Try`、`Try2` 的 doc 均为 `//macro: syntax-try`，`TryExpand` 的 doc 为 `//macro: syntax-try`，且 `linked` 含该 import path
- **THEN** 注册表 MUST 将 `Try`、`Try2` 均映射到同一 `TryExpand`

#### Scenario: 已 import 但未 link

- **WHEN** 宏主文件 import 某 provider 并调用其桩，但 `linked` 不含该 path
- **THEN** 展开 MUST 失败，且 MUST NOT 静默跳过

### Requirement: init provider 脚手架

`github.com/arcane-craft/go-macro/cmd/macro` MUST 提供 `init provider` 子命令，生成**最小** provider 目录：含桩函数与 Expander 函数各自的 `//macro:`、`Expand` 占位、**单个**语法桩及 `expand_test.go`（mactest 模板）；MUST NOT 生成 `register/` 目录；MUST NOT 默认生成 Try 式多桩族模板。

用户文档 RECOMMENDED 通过 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>` 调用该子命令。

#### Scenario: 初始化新 provider

- **WHEN** 用户执行 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac`
- **THEN** MUST 创建可编译的 provider 骨架（无 register 包）且文档指向作者指南

### Requirement: 语法桩运行时防护

provider 包内的语法桩在运行时 SHOULD 通过 `panic` 标明不可直接调用；该行为 RECOMMENDED 但 MUST NOT 作为注册表识别桩的唯一条件。

#### Scenario: 直接调用语法桩

- **WHEN** 运行时代码直接调用已登记的 provider 语法桩（非经 expand 写回）
- **THEN** 若桩实现为 panic，MUST panic 并提示不可直接调用

### Requirement: expandtool 展开入口

`macro` 包 MUST 提供子包 `macro/expandtool`，至少包含：

- `Register(importPath string, expand Expander)` — 注册可 link 的 Expander（供生成 link 代码与测试使用）
- `Registered() map[string]Expander` — 返回当前已注册副本
- `Run(args []string, linked map[string]Expander) error` — `args` 为空时 MUST 默认 `[]string{"./..."}`；`linked` 为 `nil` 时使用 `Registered()`；内部调用 `expander.ExpandPackages`
- `Main()` — `Run(os.Args[1:], nil)`，错误时 MUST 写 stderr 并以非零状态退出

宏作者 MUST NOT 实现或维护 `register/` 包。宏使用方 MUST NOT 被要求维护 `cmd/macroexpand`；默认 expand 由 `cmd/macro expand` 承担（见 `macro-codegen`）。

#### Scenario: expand 子命令使用生成 link

- **WHEN** 用户执行 `go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .`，且宏主文件 import 的 provider 已在生成 link 中 Register
- **THEN** MUST 展开已 import 且已 link 的宏调用

#### Scenario: Run 默认包路径

- **WHEN** 调用 `Run(nil, linked)` 或 `Run([]string{}, linked)`
- **THEN** MUST 等价于对 `./...` 执行 expand
