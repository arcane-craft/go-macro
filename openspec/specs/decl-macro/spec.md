# decl-macro Specification

## Purpose
定义嵌入声明宏（Decl macro）的识别规则、API（`DeclContext`/`DeclExpandResult`/`DeclExpander`）、展开顺序、作用域约束及与 Call 宏的注册/link 分工。
## Requirements
### Requirement: 声明宏定义与 Marker 语法

声明宏（Decl macro）MUST 通过宏主文件 struct 的**匿名嵌入**已注册 marker 类型触发。provider MUST 在 marker **类型**定义的 doc 中含 `//macro: <syntax-id>`。

Marker 形态 MUST 符合下列合法 Go 模板（可选参数仅通过嵌入字段 struct tag 传递）：

| 形态 | 示例 |
|------|------|
| 无参 | `DeriveStringer` |
| 必选类型实参 | `JSONMarker[Item]`（`Item` MUST 为类型名，MUST NOT 为无类型字符串常量） |
| 可选 KV | `` `macro:"k=v"` `` 写在嵌入字段上 |
| 必选 + 可选 | `Wire[T] `macro:"k=v"`` |

marker 类型内的 struct 字段 MUST 仅作 godoc 提示；展开引擎与 `DeclExpander` MUST NOT 读取其值作为配置或默认值。

#### Scenario: 匿名嵌入触发

- **WHEN** `type Item struct { provider.DeriveStringer; A int }` 且 `DeriveStringer` 已注册为 marker
- **THEN** 引擎 MUST 识别为一次 decl macro 站点

#### Scenario: 具名嵌入不触发

- **WHEN** `type Item struct { m provider.DeriveStringer }`
- **THEN** MUST NOT 作为 decl macro 站点

#### Scenario: 未注册类型不触发

- **WHEN** 用户在同一包定义 `type MyMarker struct{}` 并匿名嵌入，且未 link 对应 DeclExpander
- **THEN** MUST NOT 作为 decl macro 站点

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

### Requirement: DeclContext

`DeclContext` MUST NOT 作为 normative 作者 API。Decl 展开 MUST 使用 **`Context`（三字段）** + **`site Syntax`**。

#### Scenario: 无 DeclContext 公开接口

- **WHEN** 阅读 macro 包 normative Context 接口
- **THEN** MUST NOT 包含 `DeclContext` 类型名作为 Expander 参数

### Requirement: 声明宏作用域 MUST NOT

Decl 宏展开结果 MUST 仅影响 **被标记的 Target 类型**：其 struct 字段（含 tag）及以该类型为 receiver 的方法。

Decl 宏 MUST NOT 生成或修改：包级 `const`/`var`、其它类型声明、非 Target receiver 的函数、独立 `*_test.go` 或 mock 文件、包外 sidecar 文件。

#### Scenario: 禁止包级常量

- **WHEN** 某 `DeclExpander` 返回的 `Methods` 合法但引擎检测到试图追加 `GenDecl` 常量
- **THEN** 校验 MUST 失败（结果仅通过 Fields/Methods 通道表达）

### Requirement: 多 Marker 展开顺序与粒度

对同一 `Target` 上多个匿名嵌入 marker，引擎 MUST **按 struct 字段声明顺序**（自上而下）依次处理。每个嵌入站点 MUST 单独调用一次对应 `DeclExpander`。

#### Scenario: 字段顺序

- **WHEN** `type T struct { WireJSON; DeriveStringer; X int }` 且两者均为 marker
- **THEN** MUST 先展开 `WireJSON` 对应站点，再展开 `DeriveStringer` 对应站点

#### Scenario: 单站点单次调用

- **WHEN** `Target` 仅嵌入一个 marker
- **THEN** 引擎 MUST 仅调用一次 `DeclExpander`

### Requirement: 声明宏注册与 link

注册表 MUST 支持同一 provider 包内 **多个** `syntax-id`。每个 syntax-id MUST 可独立绑定 **至多一个** `CallExpander` 与 **至多一个** `DeclExpander`（可为 nil 表示该 syntax 仅支持一种宏）。

marker 类型通过 `(syntax-id, markerTypeName)` 或等价键查找 `DeclExpander`。仅当宏主文件 **import** provider 且 expand 已 **link** 对应 `DeclExpander` 时，嵌入才触发展开。

`macro/expandtool` MUST 提供 Decl Expander 注册入口（如 `RegisterDecl(syntaxID, DeclExpander)`），与 Call 注册分离。

#### Scenario: 同包多 syntax

- **WHEN** provider 含 `//macro: derive-stringer` 与 `//macro: wire-json` 两种 marker 类型且均已 link
- **THEN** 注册表 MUST 将各 marker 映射到各自 syntax-id 的 DeclExpander

#### Scenario: 未 link 不展开

- **WHEN** 宏主文件嵌入已注册 marker 但未 link 对应 DeclExpander
- **THEN** MUST NOT 展开该站点（MAY 报错或忽略——实现 MUST 在 author-guide 说明；推荐 expand 失败并提示 link）

### Requirement: Expand 顺序

在同一宏主文件上，引擎 MUST **先** 执行 Decl 宏展开，**再** 执行 Call 宏展开。

#### Scenario: Decl 先于 Call

- **WHEN** 同一文件含 struct marker 与 `Try(...)` 调用
- **THEN** struct 字段/方法 MUST 在 Call splice 之前完成更新

### Requirement: mactest Decl 辅助

`macro/mactest` MUST 提供构造 `DeclContext` 并调用 `DeclExpander` 的辅助（如 `ExpandDecl`），以及 `ValidateDecl(ctx, result) error` 校验全量 Fields/Methods 与 receiver 合法性。

#### Scenario: Decl 单测无需 macro tag

- **WHEN** provider 测试调用 `mactest.ExpandDecl(DeriveStringerExpand, snippet)`
- **THEN** MUST 无需 `//go:build macro` 即可 `go test`

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

