## ADDED Requirements

### Requirement: 宏调用识别与语义校验分离

展开引擎 MUST 将「是否为宏调用」（识别）与「宏调用是否合法」（语义校验）分离。识别阶段 MUST NOT 依赖特定宏的实参或返回类型规则。

### Requirement: 扫描范围

引擎 MUST 仅在宏主文件（含 `macro` build tag 的源文件）上识别宏调用，且 MUST NOT 将 `*_macro_gen.go` 作为宏调用扫描输入。

### Requirement: 基于 go/types 的符号识别

引擎 MUST 使用 `go/types` 确认调用指向**已注册 provider 包**的**包级**语法桩函数。禁止仅依据函数名匹配。

#### Scenario: 方法调用不识别

- **WHEN** 源码为 `s.MacroStub(expr)` 且 `MacroStub` 为类型方法
- **THEN** MUST NOT 识别为宏调用

### Requirement: 通用展开器分发

对每个识别到的调用点，引擎 MUST 构造 `macro.Context` 并调用 `Expander(ctx, call *ast.CallExpr) (macro.ExpandResult, error)`。

### Requirement: ExpandResult 贴回（splice）

引擎 MUST 按 `ctx.Site()` 与 `ExpandResult` 字段贴回 AST：

| Site | 字段 | 行为 |
|------|------|------|
| `SiteAssign` | `Stmts` | 替换整条 `AssignStmt` |
| `SiteReturn` | `Stmts` | 替换整条 `ReturnStmt` |
| `SiteStmt` | `Stmts` | 替换 `ExprStmt` |
| `SiteExpr` | `Expr` | 仅替换 `CallExpr` |
| `SiteReturn` | `Exprs` | 仅替换 `ReturnStmt` 的 Results 列表（首版保留；`syntax-try` 不得使用，见 `syntax-try` spec） |

若字段与 Site 不匹配，MUST 报错。引擎 MUST NOT 对 `syntax-try` 或任何 `syntax-id` 硬编码分支；贴回规则仅依赖上表。

#### Scenario: return 语境使用 Stmts（引擎行为）

- **WHEN** 某 `Expander`（如 `TryExpand`）对 `SiteReturn` 返回非空 `Stmts`
- **THEN** 引擎 MUST 用 `Stmts` 替换整条 `return` 语句

#### Scenario: 表达式宏替换 CallExpr

- **WHEN** 某 `Expander` 在 `SiteExpr` 返回 `ExpandResult{Expr: x}`
- **THEN** 引擎 MUST 仅用 `x` 替换原宏 `CallExpr`

### Requirement: 展开错误报告

当 `Expander` 返回错误或识别失败时，引擎 MUST 报告文件名、行号、原因，且 MUST NOT 静默跳过。

### Requirement: 展开器函数签名约定

每个 `Expander` MUST 为 `func(Context, *ast.CallExpr) (ExpandResult, error)`，并通过 provider 包内 `//macro: <syntax-id>` 绑定。

### Requirement: Provider 激活与 Expander 链接

`ExpandPackages(patterns, extra []Provider)` MUST 对每个待展开包：

1. 收集宏主文件所在包的 **import 路径集合**；
2. 将 `extra` 中 **import 路径命中** 的项与 **官方宏库目录** 中命中项合并为候选（同一路径时 `extra` 覆盖官方项）；
3. 仅对候选 provider 解析 AST、注册桩名并绑定 `Expander`。

引擎 MUST NOT 在识别或 splice 逻辑中对 `syntax-try`、`syntax-inline` 或任何 `syntax-id` 硬编码分支。官方宏库（`inline`、`try`）的 `Expander` 链接 MAY 集中在 `expander` 包；**MUST NOT** 在 `cmd/macro` 中默认传入完整官方列表（须由 import 驱动激活）。

#### Scenario: 未 import 的官方库不链接

- **WHEN** 宏主文件仅 import `inline`、未 import `try`
- **THEN** 该包展开时 MUST 仅注册 `syntax-inline`，MUST NOT 注册 `syntax-try`

#### Scenario: extra 覆盖官方库

- **WHEN** 调用方传入 `extra` 中 `ImportPath` 与官方 `inline` 相同且 `SyntaxID` 不同
- **THEN** 该路径 MUST 使用 `extra` 中的 `Expand`，而非官方目录中的默认项
