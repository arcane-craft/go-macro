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

`macro` 包 MUST 定义：

- `DeclExpander func(ctx DeclContext, site DeclSite) (DeclExpandResult, error)`
- `DeclExpandResult` 含 `Fields []ast.Field` 与 `Methods []*ast.FuncDecl`

`DeclExpander` 成功时 MUST 返回 **全量** `Fields`（Target 展开后的完整 struct 字段列表）与 **全量** `Methods`（receiver 为 Target 的完整方法集）。MUST NOT 使用零值/空 result 表示成功；MUST NOT 用 `nil` 切片表示「该部分未修改」。

`Methods` 中每个 `*ast.FuncDecl` 的 receiver MUST 指向 `site.Target` 所声明的类型名。

失败时 MUST 仅通过 `error` 表达；引擎 MUST NOT 写回部分结果。

#### Scenario: 成功必须双全量

- **WHEN** `DeclExpander` 返回 `nil` error
- **THEN** `Fields` 与 `Methods` MUST 均为非 nil，且分别表达 Target 的完整字段集与方法集

#### Scenario: 零值 result 失败

- **WHEN** `DeclExpander` 返回零值 `DeclExpandResult` 与 `nil` error
- **THEN** `ValidateDeclExpandResult`（或引擎等价校验）MUST 失败

#### Scenario: 纯 Contract 失败

- **WHEN** `DeclExpander` 对非法 Target 返回非 nil `error`
- **THEN** expand MUST 失败且 MUST NOT 写 gen

### Requirement: DeclContext

`DeclContext` MUST 至少提供：`FileSet`、`File`（`*ast.File`）、`Types`、`Package`、`Site()`（`DeclSite`）、`SyntaxID`、`MarkerTypeName`、`TargetMethods()`（当前 Target 在文件内的全部 `*ast.FuncDecl` 方法，供作者复制后修改）、`TempIdent`、`MacroPos()`（嵌入 marker 位置，供 `//line`）。

`DeclContext` MUST NOT 提供 `Call()`、`EnclosingFunc()` 或 `LegalSpliceTargets()`。

#### Scenario: 作者获取现有方法

- **WHEN** `DeriveStringerExpand` 需要保留 Target 已有方法
- **THEN** `ctx.TargetMethods()` MUST 返回文件中 receiver 为 Target 的全部方法

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

