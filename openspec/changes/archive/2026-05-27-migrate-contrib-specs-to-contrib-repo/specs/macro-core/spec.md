## MODIFIED Requirements

### Requirement: 宏上下文 API

`macro` 包 MUST 提供 `Context` 接口，至少包含：`FileSet`、`types.Info`、当前 `*types.Package`、宏调用 `*ast.CallExpr`、`StubName`、`SyntaxID`、`CallSiteKind`、**`EnclosingFunc`（`*ast.FuncDecl` 或 `*ast.FuncLit`，首版必选）**，以及生成临时标识符的能力（如 `TempIdent`）。

`EnclosingFunc` 为通用调用语境，供各 provider（如 contrib 仓 `syntax-try`、`syntax-inline`）自行解释；`macro` 包 MUST NOT 内置 error 位置、载荷个数 k 等 Try 专用规则（Try 规则见 contrib 仓 OpenSpec）。

#### Scenario: 展开器获取调用语境

- **WHEN** `TryExpand` 处理 `return Try(expr)`
- **THEN** `ctx.Site()` MUST 为 `SiteReturn`，且 `ctx.EnclosingFunc()` MUST 提供外层返回签名

#### Scenario: 框架不内置 Try k 规则

- **WHEN** 阅读 `macro` 包公开 API
- **THEN** MUST NOT 出现仅服务于 `Try`/`Try2` 的 k 校验函数或常量

### Requirement: ExpandResult 与 Expander 签名

`macro` 包 MUST 定义：

- `ExpandResult` 含可选字段 `Stmts []ast.Stmt`、`Exprs []ast.Expr`、`Expr ast.Expr`（首版 **保留 `Exprs`**，供罕见 return 表达式列表替换；contrib 仓 `syntax-try` 规范规定 `TryExpand` 在 `SiteReturn` 禁止使用 `Exprs`）
- `Expander func(ctx Context, call *ast.CallExpr) (ExpandResult, error)`

Provider 的 `//macro:` 标注函数 MUST 符合 `Expander` 签名。系统 MUST 支持多个 error/DSL 类 provider 并存（不同 `syntax-id`），规则仅存在于各 `Expand` 实现。

#### Scenario: 表达式宏

- **WHEN** 某宏在 `SiteExpr` 语境仅返回 `ExpandResult{Expr: e}`
- **THEN** 类型与文档 MUST 允许该形式，且引擎仅替换 `CallExpr`

#### Scenario: 语句宏

- **WHEN** `TryExpand` 在 `SiteAssign` 展开
- **THEN** MUST 返回非空 `Stmts` 以替换整条赋值语句

### Requirement: 通用宏注册与查找

系统 MUST 支持：对**宏主文件所在包已 import 的**宏库包扫描 `//macro: <syntax-id>` 与 panic 桩，并结合 **`linked map[string]macro.Expander`**（来自 `expandtool.Registered()` 或显式传入）构建注册表。`internal/expander` 与 `macro/expandtool` MUST NOT 硬编码任何具体宏库 Expander。

#### Scenario: 仅注册已 import 且已 link 的宏库

- **WHEN** 宏库已 `expandtool.Register`，但宏主文件未 import 该包
- **THEN** 注册表 MUST NOT 包含该包的桩

#### Scenario: 注册多桩名到同一 syntax-id

- **WHEN** contrib 仓 `try` 包中 `//macro: syntax-try` 绑定 `TryExpand`，存在桩 `Try` 与 `Try2`，且 expand 已 link 该 import path
- **THEN** 注册表 MUST 将 `Try`、`Try2` 等均映射到同一 `TryExpand`
