## MODIFIED Requirements

### Requirement: DeclExpander 签名与 DeclExpandResult

`macro` 包 MUST NOT 将 **`DeclExpander func(ctx DeclContext, site DeclSite) (DeclExpandResult, error)`** 作为 normative 作者 API。

Decl 宏 MUST 使用统一 **`Expander(ctx Context, site Syntax) (Syntax, error)`**。`site` MUST 表达 decl 宏锚点（pattern 如 `type $item struct { ... Derive[...] }`；**MatchedSpan 粒度由 pattern 决定**，引擎 MUST NOT 强加默认）。

成功时 Expander 返回的 `out Syntax` MUST 与 **MatchedSpan** 在父节点上下文中可拼接；**MUST NOT** 要求 out 含 Target 上 **未** match 的既有 methods 或 fields。引擎 MUST 通过 **`ValidateSplice(out, meta)` + `Apply(file, meta, out)`** 执行 Match 产出的 **`Plan`**；**MUST NOT** 修改 MatchedSpan 之外的 AST（含文件中已有、pattern 未覆盖的 `FuncDecl`）。

`out` 节点数 **MAY** 大于 MatchedSpan（与 Try 在 stmt 级替换中新增 `if err` 同理）：Derive 的 `Plan` MAY 含 `InsertAfterInFileDecls`，在 `out.ToDecls()` 中附带 **新生成** FuncDecls，均属同一 `Plan`，**MUST NOT** 引入作者级单独 Insert API。

#### Scenario: Derive 保留未 match 的既有 methods

- **WHEN** Target 类型已有 `func (T) Foo()`，pattern 为 `type $item struct { Derive[$iface] $field ... }`（与 embed/field 书写顺序无关），out 含新 TypeSpec 与生成 `String()` method
- **THEN** Apply 后 `Foo()` MUST 仍存在；仅 MatchedSpan（TypeSpec）及 out 附带的 **新** decls 被写入

#### Scenario: Derive 替换载荷含生成 methods

- **WHEN** Derive Expander 的 out `ToDecls()` 为 `[TypeSpec', FuncDecl(String)]`
- **THEN** Apply MUST 以 TypeSpec' 替换 MatchedSpan 中的 TypeSpec，并将 `String` method 作为替换载荷的一部分写入；**MUST NOT** 要求 out 列出 Target 全部既有 methods

#### Scenario: 旧 DeclExpander 无 adapter

- **WHEN** provider 仍注册 `DeclExpander` / 返回 `DeclExpandResult`
- **THEN** MUST NOT 提供语义等价 adapter；MUST 改写为 `SyntaxCase` Expander 方可展开

## MODIFIED Requirements

### Requirement: DeclContext

`DeclContext` MUST NOT 作为 normative 作者 API。Decl 展开 MUST 使用 **`Context`（三字段）** + **`site Syntax`**。

#### Scenario: 无 DeclContext 公开接口

- **WHEN** 阅读 macro 包 normative Context 接口
- **THEN** MUST NOT 包含 `DeclContext` 类型名作为 Expander 参数

### Requirement: Decl embed 元数据经 Bindings 与 Underlying

Decl 宏 **MUST NOT** 恢复 `DeclSite`、`DeclContext.Site()`、`MarkerTypeName()`、`TargetMethods()` 等为 normative 作者 API。embed 元数据 **MUST** 经 **`Bindings` + `Syntax.Underlying()`** 读取（见 design D16、macro-pattern 绑定形状、macro-syntax `Underlying`）。

| 信息 | normative 路径 |
|------|----------------|
| Target 类型名 / `TypeSpec` | `binds.Get("item")`（pattern `$item`） |
| 普通 struct 字段 | `binds.Elems("field")`；每项 `Underlying()` **MUST** 为 `*ast.Field` |
| marker 类型实参（`types.Type`） | `binds.Get("iface")` 等；`Underlying()` **MUST** 为 type `ast.Expr`；`types.Type` **MUST** 由 `ctx.Types().TypeOf(expr)` 取得 |
| `MacroTag`（`` `macro:"k=v"` ``） | 自 embed 对应 `*ast.Field.Tag` 读取，**MUST** 使用 **`macro.ParseMacroTag`**；**MUST NOT** 要求 `site.MacroTag()` 或 pattern tag 字面量语法 |
| 未 match 的既有 methods | **MUST NOT** 恢复 `TargetMethods()`；MAY 经 `site` / `MatchedSpan` 上级 `Underlying()` + `ast.Inspect`（escape，非默认 Derive 路径） |

首版 pattern 语言 **MUST NOT** 要求 tag 字面量匹配；tag 为 `*ast.Field` 的字段，与 `Type` 平级，随 field 绑定或自 MatchedSpan 内 struct 定位 embed field 取得。

#### Scenario: Derive 类型实参

- **GIVEN** 源码 `type Item struct { provider.Derive[Stringer]; Name string }`，pattern `type $item struct { Derive[$iface] $field ... }`
- **WHEN** `SyntaxCase` Transform 读取接口类型
- **THEN** `binds.Get("iface")` MUST 成功且 `Underlying()` MUST 为表示 `Stringer` 的 `ast.Expr`；`ctx.Types().TypeOf(expr)` MUST 给出对应 `types.Type`

#### Scenario: Decl pattern 与字段顺序无关

- **GIVEN** 源码 `type Item struct { Name string; provider.Derive[Stringer] }`，pattern `type $item struct { Derive[$iface] $field ... }`
- **WHEN** match 执行
- **THEN** MUST 成功；`Elems("field")` MUST 含 `Name string` 字段；结果 MUST 与 embed 在前写法等价

#### Scenario: MacroTag 自 embed Field

- **GIVEN** 源码 `type Item struct { provider.Wire[Config] \`macro:"name=foo"\` }`，pattern 含 literal `Wire[$iface]` 或等价 embed 匹配
- **WHEN** Transform 读取 `macro` tag
- **THEN** MUST 定位 embed 的 `*ast.Field`（`Underlying()`），读取 `field.Tag`，并以 `macro.ParseMacroTag(field.Tag)` 解析；MUST NOT 依赖 `DeclContext`

#### Scenario: field ellipsis 含 Tag

- **GIVEN** 源码 `type Item struct { Name string \`json:"name"\` }`，pattern `type $item struct { $field ... }`
- **WHEN** Transform 遍历 `binds.Elems("field")`
- **THEN** 每项 `Underlying()` MUST 为 `*ast.Field`，且 `field.Tag` MUST 可读取（含 `json:"name"`）

#### Scenario: 不恢复 TargetMethods

- **WHEN** Derive 仅生成新 `String()` method
- **THEN** Transform MUST NOT 要求返回文件中已有 methods；Apply 后未 match 的 `Foo()` MUST 仍保留（见 MatchedSpan scenario）

## REMOVED Requirements

### Requirement: DeclExpandResult 全量 Fields Methods 引擎 merge

**Reason**: Apply 仅替换 MatchedSpan；out 不必含未 match 的既有 methods；无引擎 merge。

**Migration**: Derive 用 pattern 划定 type 边界，Quote 产出 type + 新生成 methods；**无 Decl adapter**；见 author-guide。

### Requirement: DeclExpander 过渡 adapter

**Reason**: 旧 Decl 全量 merge / `removeTargetMethods` 与新 MatchedSpan 模型不可忠实编译（方案 C）。

**Migration**: 必须改写为 `SyntaxCase`；Call 侧 MAY 使用 `TargetToPlan` adapter 过渡。
