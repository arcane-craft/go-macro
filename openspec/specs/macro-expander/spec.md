# macro-expander Specification

## Purpose
TBD - created by archiving change go-macro-extension. Update Purpose after archive.
## Requirements
### Requirement: 宏调用识别与语义校验分离

展开引擎 MUST 将「是否为宏调用」（识别）与「宏调用是否合法」（语义校验）分离。识别阶段 MUST NOT 依赖特定宏的实参或返回类型规则。

#### Scenario: 识别不校验 Try 载荷

- **WHEN** 引擎识别到对已注册 `Try` 桩的调用
- **THEN** 识别阶段 MUST 仅依据 `go/types` 符号与注册表判定为宏调用，载荷合法性 MUST 由 `TryExpand` 在语义校验阶段处理

### Requirement: 扫描范围

引擎 MUST 仅在宏主文件（含 `macro` build tag 的源文件）上识别宏调用，且 MUST NOT 将 `*_macro_gen.go` 作为宏调用扫描输入。

#### Scenario: 生成侧不参与识别

- **WHEN** 包内同时存在 `foo.go`（`macro` tag）与 `foo_macro_gen.go`（`!macro` tag）
- **THEN** 引擎 MUST 仅在 `foo.go` 上扫描宏调用

### Requirement: 基于 go/types 的符号识别

引擎 MUST 使用 `go/types` 确认调用指向**已注册 provider 包**的**包级**语法桩函数。禁止仅依据函数名匹配。

#### Scenario: 方法调用不识别

- **WHEN** 源码为 `s.MacroStub(expr)` 且 `MacroStub` 为类型方法
- **THEN** MUST NOT 识别为宏调用

### Requirement: 通用展开器分发

对每个识别到的调用点，引擎 MUST 构造 `macro.Context` 并调用 `Expander(ctx, call *ast.CallExpr) (macro.ExpandResult, error)`。

#### Scenario: 按 syntax-id 分发

- **WHEN** 识别到 `Inline(...)` 且注册表映射到 `syntax-inline`
- **THEN** 引擎 MUST 调用对应 `InlineExpand`，且 MUST NOT 调用 `TryExpand`

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

#### Scenario: Expander 返回错误

- **WHEN** `TryExpand` 对非法 Site 返回 `error`
- **THEN** 引擎 MUST 输出含文件路径与行号的错误信息，且 MUST NOT 写回部分展开结果

### Requirement: 展开器函数签名约定

每个 `Expander` MUST 为 `func(Context, *ast.CallExpr) (ExpandResult, error)`，并通过 provider 包内 `//macro: <syntax-id>` 绑定。

#### Scenario: 注册表绑定 Expander 签名

- **WHEN** provider 包中 `//macro: syntax-inline` 标注的函数签名符合 `Expander`
- **THEN** 注册表 MUST 将该函数注册为 `syntax-inline` 的展开器

### Requirement: Provider 激活与 Expander 链接

`ExpandPackages(patterns, linked map[string]macro.Expander)` MUST 对每个待展开包：

1. 收集宏主文件所在包的 **import 路径集合**；
2. 取 **`linked` 的 key 与 import 集合的交集** 为候选宏库路径；
3. 对每个候选路径：解析 provider AST（含 `//macro:`）、注册 panic 桩，并绑定 `linked[path]` 的 `Expander`。

引擎 MUST NOT 维护官方宏库目录，MUST NOT 在识别或 splice 逻辑中对任何 `syntax-id` 硬编码分支。

`go-macro` 根 module 的 `internal/expander` 包及其测试 MUST NOT import `go-macro-contrib`。对外展开入口为 `macro/expandtool`（不导出 expander 包路径给宏使用方）。

#### Scenario: 未 import 的宏库不 link

- **WHEN** `linked` 含 `github.com/arcane-craft/go-macro-contrib/try`，但宏主文件未 import 该 path
- **THEN** 该包展开时 MUST NOT 注册 `syntax-try`

#### Scenario: 已 import 但未 link 则展开失败

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro-contrib/try` 并调用 `Try(...)`，但 expand 工具传入的 `linked` 为空或不包含该 path
- **THEN** 展开 MUST 失败（未知 stub 或未注册），且 MUST NOT 静默跳过

#### Scenario: 仅 link 已 import 的子集

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro-contrib/inline` 与 `.../try`，但 `linked` 仅含 `.../inline`
- **THEN** 该包 MUST 仅注册 `syntax-inline`；对 `Try(...)` 调用 MUST 展开失败

