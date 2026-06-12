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

对每个识别到的宏站点（Call 调用点或 Decl 嵌入点），引擎 MUST **`ResolveSite` → `site Syntax`**，构造 **`Context`**，调用 **`Expander(ctx, site) (Syntax, error)`**。

引擎 MUST NOT 将 normative 分发建立在 `CallExpander(ctx, call *ast.CallExpr)` 与 `DeclExpander(ctx, DeclSite)` 分列 API 上（**Call** adapter 期 MAY 内部转换；**Decl** MUST NOT adapter，须 native `Expander`）。

#### Scenario: 按 syntax-id 分发统一 Expander

- **WHEN** 识别到 `Inline(...)` 且注册 `syntax-inline` Expander
- **THEN** 引擎 MUST 调用该 Expander 且传入 `site Syntax`

#### Scenario: Decl 站点统一 Expander

- **WHEN** 识别到嵌入 `Derive[T]` marker
- **THEN** 引擎 MUST 以 embed `*ast.Field` 为 anchor 调用 `ResolveSite` 构造 `site Syntax`，并调用同一 Expander 类型

### Requirement: ExpandResult 贴回（splice）

引擎 MUST 在 `Expander(ctx, site)` 返回 `out` 后，从 **`site` 内部 meta 槽**读取 **`MatchMeta`**（见 design D15、macro-syntax site 内部槽），再：

1. 调用 **`ValidateSplice(out, meta)`**
2. 调用 **`Apply(file, meta, out)`**
3. 调用 **`StampStmtPos(site.MacroPos(), ...)`**（适用时）

`MatchMeta` MUST NOT 由公开 `Expander` 返回值携带；`out Syntax` MUST NOT 携带 `Plan`。meta 槽为空时 MUST 在步骤 1 之前失败。

MUST NOT 要求 Expander 返回 `CallExpandResult` / `DeclExpandResult`（**Call** adapter 除外；**Decl** MUST NOT adapter）。MUST NOT 调用 normative **`InferTarget`**。Apply MUST 按 `meta.Plan` 执行；out 节点数 MAY 大于 MatchedSpan（Try 多 stmt、Derive 多 decl）。

#### Scenario: 推断 AssignStmt 贴回

- **WHEN** Try clause 为 stmt 级 pattern、MatchedSpan 为 AssignStmt、out 为多条 Stmts
- **THEN** Apply MUST 以 out 的 stmts 替换该 AssignStmt

#### Scenario: Derive 贴回不触碰未 match methods

- **WHEN** Derive clause match `type $item struct { ... }`、out 含 TypeSpec' 与生成 method
- **THEN** Apply MUST 仅处理 MatchedSpan 与 out 载荷；文件中未 match 的既有 methods MUST 保留

### Requirement: ApplyExpandResult 不依赖 CallSiteKind

贴回 MUST 依据 **`meta.Plan`**、**`MatchedSpan`** 与 `out Syntax` 形状；MUST NOT 将 `CallSiteKind` 作为分支条件（`CallSiteKind` MAY 从 author API 移除）。

#### Scenario: out 与 Plan 不兼容

- **WHEN** `meta.Plan` 要求 `ToExpr()` 但 `out` 无法 `ToExpr()`
- **THEN** `ValidateSplice` MUST 失败

### Requirement: 展开错误报告

当 **`CallExpander` 或 `DeclExpander`** 返回错误或识别失败时，引擎 MUST 报告文件名、行号、原因，且 MUST NOT 静默跳过或写回部分结果。

#### Scenario: DeclExpander 返回错误

- **WHEN** `DeriveStringerExpand` 返回 error
- **THEN** MUST 含文件路径与行号，且 MUST NOT 写 gen

### Requirement: 展开器函数签名约定

每个 **Call** Expander MUST 为 **`func(CallContext, *ast.CallExpr) (CallExpandResult, error)`**，doc 含 `//macro: <syntax-id>`。

每个 **Decl** Expander MUST 为 **`func(DeclContext, DeclSite) (DeclExpandResult, error)`**，doc 含 `//macro: <syntax-id>`（通常与 marker 类型同 syntax-id）。

#### Scenario: 注册表绑定 CallExpander

- **WHEN** provider 中 `InlineExpand` doc 含 `//macro: syntax-inline` 且签名符合 `CallExpander`
- **THEN** 注册表 MUST 将 `syntax-inline` 绑定到该函数

### Requirement: Provider 激活与 Expander 链接

`ExpandPackages` MUST：

1. 发现宏主文件 import 的 provider；
2. 解析各 syntax-id 的 Call 桩、Decl marker、CallExpander、DeclExpander；
3. 仅激活 **已 import 且已 link** 的 syntax-id；
4. 对每文件 **先 `ExpandDeclMacros` 再 `ExpandCallMacros`**。

link 生成 MUST 支持分别注册 Call 与 Decl Expander（按 syntax-id）。

#### Scenario: 先 Decl 后 Call

- **WHEN** 同一宏主文件含 struct marker 与 `Try(...)`
- **THEN** MUST 在 Call splice 之前完成 Decl 展开

### Requirement: Provider 语义以外置规范为准

`TryExpand`、`InlineExpand` 的载荷校验、Site 禁止规则、展开语句形态等 provider 级语义 MUST 以 `go-macro-contrib` 仓库内 `syntax-try`、`syntax-inline` OpenSpec 为准。本规范仅定义展开引擎的识别、分发、Call/Decl 贴回与 link/import 边界。

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

### Requirement: Decl 宏识别

引擎 MUST 在宏主文件（`macro` build tag）上扫描 struct 的 **匿名嵌入**字段。对解析为已注册 marker 的嵌入，MUST 构造 `DeclSite`（含 `Target`、`EmbedIndex`、`MarkerTypeArgs`、`MacroTag`）。

MUST NOT 将 `*_macro_gen.go` 作为 Decl 扫描输入。

#### Scenario: 泛型 marker 实例

- **WHEN** 嵌入 `WireExtra[MyType]` 且 `WireExtra` 为已注册泛型 marker
- **THEN** `DeclSite.MarkerTypeArgs` MUST 含 `MyType`

### Requirement: DeclExpandResult 贴回

引擎 MUST 校验 **`DeclExpandResult`**：`Fields` 与 `Methods` 均非 nil；`Methods` receiver 均为 Target。

贴回 MUST：

1. 将 `Target` 的 `StructType.Fields.List` 替换为 `result.Fields`；
2. 从 `file.Decls` 移除 receiver 为 Target 的既有 `*ast.FuncDecl`；
3. 在 `Target` 的 `TypeSpec` 之后插入 `result.Methods`。

MUST NOT 写入包级 `const`/`var` 或其它类型。

#### Scenario: 全量 Methods 替换

- **WHEN** `DeclExpander` 返回的 `Methods` 漏掉 Target 原有方法 `Validate`
- **THEN** 贴回后文件 MUST NOT 再含 `Validate`（作者责任；引擎不自动合并）

#### Scenario: 校验零值 result

- **WHEN** `DeclExpander` 返回零值 `DeclExpandResult` 与 nil error
- **THEN** expand MUST 失败

### Requirement: 多 Marker 顺序

对同一 Target 多个嵌入站点，MUST 按字段声明顺序依次调用 `DeclExpander` 并贴回。每次贴回后 MAY 重新扫描剩余站点以稳定索引。

#### Scenario: 双 marker 顺序

- **WHEN** `type T struct { WireJSON; DeriveStringer; X int }`
- **THEN** MUST 先处理 `WireJSON` 再处理 `DeriveStringer`

### Requirement: 扫描范围扩展

引擎扫描范围 MUST 包含：Call 宏调用、语法桩值用法校验、**Decl 嵌入 marker**。三者均在宏主文件上进行。

#### Scenario: gen 不参与 Decl 扫描

- **WHEN** 存在 `foo.go`（macro）与 `foo_macro_gen.go`（!macro）
- **THEN** Decl 扫描 MUST 仅针对 `foo.go`

### Requirement: ResolveSite

引擎 MUST 在 `internal/expander` 提供 `ResolveSite(file, anchor) Syntax`，在每轮 expand 前对 **当前 AST** 调用。anchor 类型（design D18）：

- **Call**：`*ast.CallExpr`
- **Decl**：embed `*ast.Field`

每轮 MUST 构造**空** meta 槽的 `site`。同句多宏 MUST 每轮重新 Resolve。

#### Scenario: splice 后重新 Resolve

- **WHEN** 同函数内两颗宏先后展开
- **THEN** 第二轮 ResolveSite MUST 反映第一轮 splice 后的 stmt 形状

#### Scenario: Decl ResolveSite anchor

- **GIVEN** `type T struct { provider.Derive[I]; X int }`
- **WHEN** 引擎展开 Derive
- **THEN** MUST 以 Derive 的 embed `*ast.Field` 为 anchor 调用 `ResolveSite`；`site.MacroPos()` MUST 为 embed 位置

### Requirement: MatchMeta 从 site 读取

每轮 expand 流水线 MUST 遵循：`ResolveSite` 构造**空** meta 槽的 `site` → `Expander(ctx, site) (Syntax, error)` → `MatchMetaFromSite(site)` → `ValidateSplice` → `Apply`。normative 路径下 meta 由 `SyntaxRules` / `SyntaxCase` 或手写 `Expander` 内 `site.Match` 写入（design D19）；Call adapter 路径由 `TargetToPlan` 写入。

#### Scenario: SyntaxCase 端到端 meta 来源

- **GIVEN** `SyntaxCase` 某 clause match 成功且 `Transform` 返回 `out`
- **WHEN** `Expander` 返回
- **THEN** 引擎 MUST 从同一 `site` 读取含该 clause `Plan` 的 `MatchMeta`；MUST NOT 要求 `Transform` 返回或设置 `MatchMeta`

#### Scenario: Call adapter 写入 meta 槽

- **GIVEN** Call adapter 包装旧 `CallExpander` 且 `TargetToPlan` 成功
- **WHEN** adapter 版 `Expander` 返回 `out`
- **THEN** meta 槽 MUST 已含 `TargetToPlan` 产出的 `MatchMeta`；引擎 MUST 与 normative 路径使用相同读取与 `ValidateSplice` / `Apply` 流程

