# macro-core Specification

## Purpose
TBD - created by archiving change go-macro-extension. Update Purpose after archive.
## Requirements
### Requirement: 宏上下文 API

`macro` 包 MUST 提供 `Context` 接口，至少包含：`FileSet`、`types.Info`、当前 `*types.Package`、宏调用 `*ast.CallExpr`、`StubName`、`SyntaxID`、`CallSiteKind`、**`EnclosingFunc`（`*ast.FuncDecl` 或 `*ast.FuncLit`，首版必选）**，以及生成临时标识符的能力（如 `TempIdent`）。

`EnclosingFunc` 为通用调用语境，供各 provider（如 `syntax-try`、`syntax-inline`）自行解释；`macro` 包 MUST NOT 内置 error 位置、载荷个数 k 等 Try 专用规则。

#### Scenario: 展开器获取调用语境

- **WHEN** `TryExpand` 处理 `return Try(expr)`
- **THEN** `ctx.Site()` MUST 为 `SiteReturn`，且 `ctx.EnclosingFunc()` MUST 提供外层返回签名

### Requirement: ExpandResult 与 Expander 签名

`macro` 包 MUST 定义：

- `ExpandResult` 含可选字段 `Stmts []ast.Stmt`、`Exprs []ast.Expr`、`Expr ast.Expr`（首版 **保留 `Exprs`**，供罕见 return 表达式列表替换；`syntax-try` 在 `SiteReturn` 禁止使用 `Exprs`，见 `syntax-try` spec）
- `Expander func(ctx Context, call *ast.CallExpr) (ExpandResult, error)`

Provider 的 `//macro:` 标注函数 MUST 符合 `Expander` 签名。系统 MUST 支持多个 error/DSL 类 provider 并存（不同 `syntax-id`），规则仅存在于各 `Expand` 实现。

#### Scenario: 表达式宏

- **WHEN** 某宏在 `SiteExpr` 语境仅返回 `ExpandResult{Expr: e}`
- **THEN** 类型与文档 MUST 允许该形式，且引擎仅替换 `CallExpr`

#### Scenario: 语句宏

- **WHEN** `TryExpand` 在 `SiteAssign` 展开
- **THEN** MUST 返回非空 `Stmts` 以替换整条赋值语句

### Requirement: AST 节点抽象

`macro` 包 MAY 提供 `ast.Node` 包装以便跨包隐藏 `go/ast`；展开器 MUST 能构造 `ast.Stmt` / `ast.Expr` 供 `ExpandResult` 使用。

#### Scenario: ExpandResult 使用 go/ast 节点

- **WHEN** 某 `Expander` 返回 `ExpandResult{Stmts: []ast.Stmt{...}}`
- **THEN** 展开器 MUST 接受标准 `go/ast` 节点并完成 splice

### Requirement: 通用宏注册与查找

系统 MUST 支持：对**宏主文件所在包已 import 的**宏库包扫描 `//macro: <syntax-id>` 与 panic 桩，并结合 **`linked map[string]macro.Expander`**（来自 `expandtool.Registered()` 或显式传入）构建注册表。`internal/expander` 与 `macro/expandtool` MUST NOT 硬编码任何具体宏库 Expander。

#### Scenario: 仅注册已 import 且已 link 的宏库

- **WHEN** 宏库已 `expandtool.Register`，但宏主文件未 import 该包
- **THEN** 注册表 MUST NOT 包含该包的桩

#### Scenario: 注册多桩名到同一 syntax-id

- **WHEN** 宏库包中 `//macro: syntax-try` 绑定 `TryExpand`，存在桩 `Try` 与 `Try2`，且 `linked` 含该 import path
- **THEN** 注册表 MUST 将 `Try`、`Try2` 均映射到同一 `TryExpand`

### Requirement: 轻薄 AST 辅助（首版）

`macro` 包首版 MUST 仅提供最小辅助（`Context`、`ExpandResult`、`TempIdent`、定位/错误辅助）。MUST NOT 在首版要求厚重 astbuilder；后续宏增多后再抽取。

#### Scenario: 首版无 astbuilder 依赖

- **WHEN** provider 作者实现 `Expander` 并构造展开 AST
- **THEN** MUST 仅依赖 `macro` 包首版 API，且 MUST NOT 要求引入独立 astbuilder 包

### Requirement: Provider 纯 Expand 测试辅助

`macro` 包 MUST 提供 `mactest`（或等价子包），使 provider 作者在不使用 `//go:build macro`、不跑全链路 expand 的情况下，能构造 `Context` 并调用 `Expander` 做单元测试。

#### Scenario: TryExpand 单测

- **WHEN** `try` 包测试调用 `mactest.Expand(TryExpand, snippet)`
- **THEN** 测试 MUST 无需 macro tag 即可 `go test ./try/...`

### Requirement: init provider 脚手架

`github.com/arcane-craft/go-macro/cmd/macro` MUST 提供 `init provider` 子命令，生成**最小** provider 目录：含 `//macro:`、`Expand` 占位、**单个** panic 语法桩及 `expand_test.go`（mactest 模板）；MUST NOT 默认生成 Try 式多桩族模板。

用户文档 RECOMMENDED 通过 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>` 调用该子命令（`go tool macro` MAY 在已安装 tool 的环境下使用，但 MUST NOT 作为唯一文档入口）。

#### Scenario: 初始化新 provider

- **WHEN** 用户执行 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac`
- **THEN** MUST 创建可编译的 provider 骨架且文档指向作者指南

### Requirement: 语法桩运行时防护

provider 包内的语法桩（包级 panic 函数）在运行时 MUST panic，并标明宏名与不可直接调用。

#### Scenario: 直接调用语法桩

- **WHEN** 运行时代码直接调用 provider 包级语法桩（非经 expand 写回）
- **THEN** MUST panic 并提示不可直接调用

### Requirement: expandtool 展开入口

`macro` 包 MUST 提供子包 `macro/expandtool`，供 expand 二进制与高级集成使用，至少包含：

- `Register(importPath string, expand Expander)` — 注册可 link 的 Expander
- `Registered() map[string]Expander` — 返回当前已注册副本
- `Run(args []string, linked map[string]Expander) error` — `args` 为空时 MUST 默认 `[]string{"./..."}`；`linked` 为 `nil` 时使用 `Registered()`；内部调用 `expander.ExpandPackages`
- `Main()` — `Run(os.Args[1:], nil)`，错误时 MUST 写 stderr 并以非零状态退出

框架 MUST 仅负责提供上述展开能力与接线模式；provider 作者 MUST NOT 被要求实现或维护 expand main。默认可执行入口 MUST 为 `cmd/macro expand`（或其等价调用），MUST NOT 要求宏使用方维护 `cmd/macroexpand`。

#### Scenario: Main 使用 Registered 注册表

- **WHEN** 某调用方入口进程已通过 `register` 包 `init` 调用 `expandtool.Register` 注册宏库，且执行该入口命令
- **THEN** MUST 展开宏主文件中已 import 且已 Register 的宏调用

#### Scenario: Run 默认包路径

- **WHEN** 调用 `Run(nil, linked)` 或 `Run([]string{}, linked)`
- **THEN** MUST 等价于对 `./...` 执行 expand

