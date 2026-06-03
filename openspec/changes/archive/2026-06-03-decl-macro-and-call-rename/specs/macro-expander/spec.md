## MODIFIED Requirements

### Requirement: 通用展开器分发

对每个识别到的 **Call** 调用点，引擎 MUST 构造 `macro.CallContext` 并调用 **`CallExpander(ctx, call *ast.CallExpr) (macro.CallExpandResult, error)`**。

对每个识别到的 **Decl** 嵌入站点，引擎 MUST 构造 `macro.DeclContext` 并调用 **`DeclExpander(ctx, site DeclSite) (macro.DeclExpandResult, error)`**。

#### Scenario: 按 syntax-id 分发 Call

- **WHEN** 识别到 `Inline(...)` 且注册表映射到 `syntax-inline`
- **THEN** 引擎 MUST 调用对应 `CallExpander`，且 MUST NOT 调用 `TryExpand`

#### Scenario: 按 syntax-id 分发 Decl

- **WHEN** 识别到嵌入 `DeriveStringer` 且注册表映射到 `derive-stringer`
- **THEN** 引擎 MUST 调用对应 `DeclExpander`

### Requirement: ExpandResult 贴回（splice）

引擎 MUST 按 **`CallExpandResult.Target`** 与载荷贴回 Call 宏 AST。引擎 MUST 在贴回前调用 **`macro.ValidateCallExpandResult`**。规则表（六种 `SpliceReplace*`）不变。

#### Scenario: assign 仅换 RHS

- **WHEN** 某 `CallExpander` 对 `x := Macro()` 返回 `Target: SpliceReplaceAssignRHS` 与 `Expr: e`
- **THEN** 引擎 MUST 保留 `x` 在 `Lhs` 中

### Requirement: ApplyExpandResult 不依赖 CallSiteKind

`ApplyExpandResult` MUST 仅接受 `file`、`call`、**`CallExpandResult`**；MUST NOT 将 `CallSiteKind` 作为贴回分支条件。

#### Scenario: Site 与 Target 解耦

- **WHEN** `ctx.Site()` 为 `SiteAssign` 但 `CallExpandResult.Target` 为 `SpliceReplaceCallExpr` 且宏在 assign RHS
- **THEN** splice MUST 失败

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

## ADDED Requirements

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
