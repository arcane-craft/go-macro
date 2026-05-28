# macro-expander Specification

## Purpose
TBD - created by archiving change go-macro-extension. Update Purpose after archive.
## Requirements
### Requirement: 宏调用识别与语义校验分离

展开引擎 MUST 将「是否为宏调用」（识别）、「语法桩是否被当作函数值使用」（引擎级值用法校验）与「宏调用是否合法」（provider 语义校验）分离。识别阶段与值用法校验阶段 MUST NOT 依赖特定宏的实参或返回类型规则。

#### Scenario: 识别不校验 Try 载荷

- **WHEN** 引擎识别到对已注册 `Try` 桩的调用
- **THEN** 识别阶段 MUST 仅依据 `go/types` 符号与注册表判定为宏调用，载荷合法性 MUST 由 `TryExpand` 在语义校验阶段处理

#### Scenario: 值用法校验不校验 Try 载荷

- **WHEN** 宏主文件含 `try.Try(badPayload)` 且实参在 `TryExpand` 语义下非法
- **THEN** 值用法校验 MUST NOT 因载荷非法而失败；失败 MUST 由 `TryExpand` 在展开阶段返回

### Requirement: 扫描范围

引擎 MUST 仅在宏主文件（含 `macro` build tag 的源文件）上识别宏调用**并**执行语法桩值用法校验，且 MUST NOT 将 `*_macro_gen.go` 作为上述扫描输入。

#### Scenario: 生成侧不参与识别

- **WHEN** 包内同时存在 `foo.go`（`macro` tag）与 `foo_macro_gen.go`（`!macro` tag）
- **THEN** 引擎 MUST 仅在 `foo.go` 上扫描宏调用与桩值用法

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

引擎 MUST 按 `ExpandResult.Target` 与对应载荷贴回 AST。引擎 MUST 在贴回前调用 `macro.ValidateExpandResult`（或等价校验）。引擎 MUST NOT 依据 `ctx.Site()` 选择贴回字段；`Site()` 仅用于 provider 语义与错误提示。

| `Target` | 行为 |
|----------|------|
| `SpliceReplaceAssignStmt` | 替换整条 `AssignStmt` 为 `Stmts` |
| `SpliceReplaceAssignRHS` | 在 enclosing `AssignStmt.Rhs` 中定位宏 `CallExpr` 所在槽位，仅用 `Expr` 替换该槽位；`Lhs` MUST 不变 |
| `SpliceReplaceReturnStmt` | 替换整条 `ReturnStmt` 为 `Stmts` |
| `SpliceReplaceReturnResults` | 仅将 `ReturnStmt.Results` 设为 `Exprs` |
| `SpliceReplaceExprStmt` | 替换整条 `ExprStmt` 为 `Stmts` |
| `SpliceReplaceCallExpr` | 在表达式槽中用 `Expr` 替换宏 `CallExpr`（含 `BinaryExpr`、`CallExpr` 参数、`CompositeLit` 等父节点） |

若 `Target` 不在当前调用处的结构合法集合内，或载荷与 `Target` 不匹配，MUST 报错。错误信息 MUST 含文件名与行号，并 SHOULD 列出 `LegalSpliceTargets()` 允许的目标名称。引擎 MUST NOT 对任何 `syntax-id` 硬编码 splice 分支。

#### Scenario: assign 仅换 RHS

- **WHEN** 某 `Expander` 对 `x := Macro()` 返回 `Target: SpliceReplaceAssignRHS` 与 `Expr: e`
- **THEN** 引擎 MUST 保留 `x` 在 `Lhs` 中，且 `Rhs` 中含宏的项 MUST 变为 `e`

#### Scenario: return 语境替换整条 return

- **WHEN** 某 `Expander`（如 contrib 仓 `TryExpand`）对 `return Macro(...)` 返回 `Target: SpliceReplaceReturnStmt` 与非空 `Stmts`
- **THEN** 引擎 MUST 用 `Stmts` 替换整条 `return` 语句

#### Scenario: return 语境仅换 Results

- **WHEN** 某 `Expander` 对 `return Macro(...)` 返回 `Target: SpliceReplaceReturnResults` 与非空 `Exprs`
- **THEN** 引擎 MUST 仅替换 `ReturnStmt.Results`，且 MUST NOT 替换整条 `ReturnStmt` 为语句块

#### Scenario: 表达式宏替换 CallExpr

- **WHEN** 某 `Expander` 对表达式槽宏返回 `Target: SpliceReplaceCallExpr` 与 `Expr: x`
- **THEN** 引擎 MUST 仅用 `x` 替换原宏 `CallExpr`

#### Scenario: Target 与锚点不一致

- **WHEN** 宏位于 `return Macro()` 但 `Expander` 返回 `Target: SpliceReplaceCallExpr`
- **THEN** expand MUST 失败，且 MUST NOT 写回部分展开结果

#### Scenario: 载荷与 Target 不匹配

- **WHEN** `Expander` 返回 `Target: SpliceReplaceAssignRHS` 且 `Expr` 为 nil
- **THEN** expand MUST 失败

### Requirement: ApplyExpandResult 不依赖 CallSiteKind

`ApplyExpandResult`（或等价 splice 入口）MUST 仅接受 `file`、`call`、`ExpandResult`；MUST NOT 将 `CallSiteKind` 作为贴回分支条件。

#### Scenario: Site 与 Target 解耦

- **WHEN** `ctx.Site()` 为 `SiteAssign` 但 `ExpandResult.Target` 为 `SpliceReplaceCallExpr` 且宏在 assign RHS
- **THEN** splice MUST 失败（即使 `Site` 为 assign 语境）

### Requirement: 展开错误报告

当 `Expander` 返回错误或识别失败时，引擎 MUST 报告文件名、行号、原因，且 MUST NOT 静默跳过。

#### Scenario: Expander 返回错误

- **WHEN** `TryExpand` 对非法 Site 返回 `error`
- **THEN** 引擎 MUST 输出含文件路径与行号的错误信息，且 MUST NOT 写回部分展开结果

### Requirement: 展开器函数签名约定

每个 `Expander` MUST 为 `func(Context, *ast.CallExpr) (ExpandResult, error)`，且 MUST 在其 doc 中含 `//macro: <syntax-id>`。注册表 MUST 将该函数绑定为对应 syntax-id 的展开器（并与同 syntax-id 的桩关联）。

#### Scenario: 注册表绑定 Expander 签名

- **WHEN** provider 包中 `InlineExpand` 的 doc 含 `//macro: syntax-inline` 且签名符合 `Expander`
- **THEN** 注册表 MUST 将 `syntax-inline` 的展开实现绑定到 `linked` 提供的 `InlineExpand`

### Requirement: Provider 激活与 Expander 链接

`ExpandPackages(patterns, linked map[string]macro.Expander)` MUST 对每个待展开包：

1. 收集宏主文件所在包的 **import 路径集合**；
2. 取 **`linked` 的 key 与 import 集合的交集** 为候选宏库路径；
3. 对每个候选路径：解析 provider AST，从**各函数 doc** 读取 `//macro:`，登记桩与 syntax-id，并绑定 `linked[path]` 的 `Expander`。

引擎 MUST NOT 维护官方宏库目录，MUST NOT 在识别或 splice 逻辑中对任何 `syntax-id` 硬编码分支。

`go-macro` 根 module 的 `internal/expander` 包及其测试 MUST NOT import `go-macro-contrib`。对外展开入口为 `macro/expandtool` 与 `cmd/macro expand`。

#### Scenario: 未 import 的宏库不 link

- **WHEN** `linked` 含 `github.com/arcane-craft/go-macro-contrib/try`，但宏主文件未 import 该 path
- **THEN** 该包展开时 MUST NOT 注册 `syntax-try`

#### Scenario: 已 import 但未 link 则展开失败

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro-contrib/try` 并调用 `Try(...)`，但 expand 工具传入的 `linked` 为空或不包含该 path
- **THEN** 展开 MUST 失败（未知 stub 或未注册），且 MUST NOT 静默跳过

#### Scenario: 仅 link 已 import 的子集

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro-contrib/inline` 与 `.../try`，但 `linked` 仅含 `.../inline`
- **THEN** 该包 MUST 仅注册 `syntax-inline`；对 `Try(...)` 调用 MUST 展开失败

#### Scenario: 识别使用 importPath 与桩名

- **WHEN** 两个 provider 均含名为 `Macro` 的桩且各自 doc 含 `//macro:`，宏主文件分别通过不同 import path 调用
- **THEN** 引擎 MUST 按各自 import path 分发到对应 Expander，且 MUST NOT 因桩名相同而错绑

### Requirement: Provider 语义以外置规范为准

`TryExpand`、`InlineExpand` 的载荷校验、Site 禁止规则、展开语句形态等 provider 级语义 MUST 以 `go-macro-contrib` 仓库内 `syntax-try`、`syntax-inline` OpenSpec 为准。本规范仅定义展开引擎的识别、分发、`ExpandResult` 贴回与 link/import 边界。

#### Scenario: 修改 Try 展开语义

- **WHEN** 维护者需变更 `return Try` 的错误路径语句形态
- **THEN** MUST 修改 contrib 仓 `syntax-try` spec 及 `try` 实现，而非在本 spec 中新增 Try 专用引擎分支

### Requirement: 语法桩禁止值用法（expand 期校验）

在宏调用识别与展开之前，展开引擎 MUST 对**宏主文件**执行语法桩**值用法**校验：凡 `go/types` 解析为**当前注册表**中已登记 `(importPath, stubName)` 的 provider **包级**语法桩，且该 AST 节点**不是**宏直调 `CallExpr` 的 callee（剥除外层 `ParenExpr` 后等于 `CallExpr.Fun`），expand MUST 失败。

校验 MUST 使用与宏调用识别相同的符号解析规则（包级 `*types.Func`、`SelectorExpr` 的 `X` 为 `*types.PkgName`、dot-import 裸 `Ident` 等）。引擎 MUST NOT 仅依据函数名匹配。

校验 MUST 在**同一宏主文件**内一视同仁：MUST NOT 为 provider 作者、死代码、反射等场景单独豁免。

错误信息 MUST 含文件名、行号、列号、桩名（及 import path 或本地 import 名若可解析），并 MUST 提示宏桩须以 `pkg.Stub(...)`（或 dot-import 下 `Stub(...)`）**直接调用**，MUST NOT 作为函数值使用。引擎 MUST NOT 静默跳过。

#### Scenario: 桩作函数实参传递

- **WHEN** 宏主文件已 link 并注册 `example.com/try` 的桩 `Try`，源码为 `apply(try.Try)` 且 `apply` 为普通函数
- **THEN** expand MUST 在 `try.Try` 处失败，且 MUST NOT 展开该文件

#### Scenario: 桩赋值给变量

- **WHEN** 已注册桩 `Try`，源码为 `fn := try.Try`
- **THEN** expand MUST 在 `try.Try` 处失败

#### Scenario: 桩作为 return 值

- **WHEN** 已注册桩 `Try`，源码为 `return try.Try`
- **THEN** expand MUST 在 `try.Try` 处失败

#### Scenario: reflect 获取桩函数值

- **WHEN** 已注册桩 `Try`，源码为 `reflect.ValueOf(try.Try)` 或 `reflect.TypeOf(try.Try)`
- **THEN** expand MUST 在 `try.Try` 处失败

#### Scenario: 死代码中的桩值引用

- **WHEN** 已注册桩 `Try`，源码为 `if false { _ = try.Try }`
- **THEN** expand MUST 在 `try.Try` 处失败（引擎 MUST NOT 做可达性分析而跳过）

#### Scenario: 直调仍为合法宏调用

- **WHEN** 已注册桩 `Try`，源码为 `return try.Try(expr)` 或 `(try.Try)(expr)`
- **THEN** 值用法校验 MUST NOT 对该 `try.Try` 报错；引擎 MUST 按现有规则识别为宏调用并展开

#### Scenario: 未 link 的宏库不校验值用法

- **WHEN** 宏主文件 `import` 了 `example.com/try` 但本次 expand 的 `linked` 未包含该 path（注册表无 `Try` 桩）
- **THEN** 引擎 MUST NOT 因 `var _ = try.Try` 等值用法失败；对 `try.Try(...)` 直调 MUST 仍按现有「未注册桩 / 展开失败」规则处理

#### Scenario: shadow 同名不误报

- **WHEN** 宏主文件包内定义 `func Try(int) int` 且 `return Try(1)` 或 `_ = Try`
- **THEN** `go/types` 将 `Try` 解析为本包函数，非 provider 桩
- **THEN** 值用法校验 MUST NOT 报错，且宏识别 MUST NOT 将此类调用识别为宏（与现有 shadow 行为一致）

#### Scenario: 方法名与桩同名不误报

- **WHEN** 源码为 `s.Try(1)` 且 `Try` 为类型 `S` 的方法
- **THEN** 值用法校验 MUST NOT 报错，且 MUST NOT 识别为宏调用

#### Scenario: 嵌套直调中的内层桩

- **WHEN** 源码为 `outer(try.Try(1))` 且 `Try` 已注册
- **THEN** 内层 `try.Try` 作为 `CallExpr` callee MUST NOT 触发值用法错误；外层 `outer(...)` MUST NOT 被识别为宏

#### Scenario: 校验失败阻断写回

- **WHEN** 值用法校验在宏主文件任一处失败
- **THEN** expand MUST 失败，且 MUST NOT 写回 `*_macro_gen.go` 或部分展开结果

