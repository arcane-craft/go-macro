## ADDED Requirements

### Requirement: Syntax 统一 AST 片段类型

系统 MUST 在 `macro` 包提供 `Syntax` 接口，作为宏 Match、Quote 与贴回载荷的统一类型。`Syntax` MUST 提供 `Match(pattern string) (Bindings, error)`、`ToExpr` / `ToExprs` / `ToStmt` / `ToStmts` / `ToDecl` / `ToDecls`（严格形状，失败返回 error）、`MacroPos() token.Pos`、`Underlying() ast.Node`。

#### Scenario: ToStmts 校验非空语句列表

- **WHEN** 某 `Syntax` 内含两条 `ast.Stmt`
- **THEN** `ToStmts()` MUST 成功且长度 ≥ 1

#### Scenario: ToExpr 拒绝语句列表

- **WHEN** 某 `Syntax` 仅含 `[]ast.Stmt`
- **THEN** `ToExpr()` MUST 返回 error

### Requirement: Bindings 捕获表

系统 MUST 提供 `Bindings` 接口，至少含：

- `Get(name string) (Syntax, bool)` — 单项捕获（如 `$inner`、`$iface`、`$item`）
- `Elems(name string) ([]Syntax, bool)` — ellipsis 捕获（如 `$field ...`、`$lhs ...`、`$vals ...`）

Match 捕获名 MUST 映射为 `Syntax` 值；ellipsis 捕获 MUST 经 `Elems` 读取，**MUST NOT** 仅用 `Get` 表示列表。

#### Scenario: 获取已捕获名

- **WHEN** Match 成功且 pattern 含 `$inner`
- **THEN** `Bindings.Get("inner")` MUST 返回 `(Syntax, true)`

#### Scenario: 获取 ellipsis 捕获

- **WHEN** Match 成功且 pattern 含 `$field ...`，struct 含 2 个具名字段
- **THEN** `Bindings.Elems("field")` MUST 返回长度 2 的切片与 `true`

#### Scenario: 缺失捕获名

- **WHEN** 调用 `Get("missing")` 或 `Elems("missing")`
- **THEN** MUST 返回 `(nil, false)` 或零值与 `false`

### Requirement: Underlying 读取 embed 元数据

`Syntax.Underlying()` **MUST** 返回绑定或 `site` 所包装的真实 `ast.Node`，供 Decl 宏读取 today `DeclSite` 等价信息（见 decl-macro、design D16）。**MUST NOT** 为 MacroTag、类型实参等单独增加 `site` accessor；**MUST** 保留公开 **`macro.ParseMacroTag(*ast.BasicLit) MacroTag`** 供 Transform 解析 `` `macro:"k=v"` ``。

Transform **MAY** 对 `Underlying()` 结果做 type assert 与 `ast.Inspect`（读未 match AST、遍历 file 等）；此为 escape，非默认 Derive 路径。

#### Scenario: ParseMacroTag 公开

- **WHEN** Decl Transform 自 `*ast.Field.Tag` 读取配置
- **THEN** MUST 可调用 `macro.ParseMacroTag` 而无需 `DeclContext`

#### Scenario: Types 与 type 实参绑定组合

- **WHEN** `binds.Get("iface").Underlying()` 为 `ast.Expr`
- **THEN** `ctx.Types().TypeOf(expr)` MUST 可用于取得 marker 类型实参的 `types.Type`

### Requirement: site Syntax 与 anchor

引擎 MUST 为每轮 expand 构造 `site Syntax`，内部含本轮待展开 anchor 与 enclosing 根；anchor MUST NOT 作为公开字段暴露。Call/Decl anchor 语义见 design **D18**：

| 站点 | anchor | `site.MacroPos()` |
|------|--------|-------------------|
| **Call** | `*ast.CallExpr` | `call.Pos()` |
| **Decl** | embed `*ast.Field` | `embedField.Pos()`（与 today `DeclContext.MacroPos()` 一致） |

`site.Match(pattern)` 产生的 **MatchedSpan**（见 macro-pattern）为贴回边界；Call/Decl MUST 共用同一 Match/Apply 模型。

#### Scenario: 同句多宏各轮独立 site

- **WHEN** 同一 `AssignStmt` 含两个宏 Call 且按 Pos 降序展开
- **THEN** 每轮 Expander MUST 收到不同 anchor 的 `site`，且 `Match` 仅 unify 当前 anchor

#### Scenario: Decl site 以 embed 为 anchor

- **GIVEN** 源码 `type Item struct { provider.Derive[Stringer]; Name string }`
- **WHEN** 引擎对 Derive embed 调用 `ResolveSite(file, embedField)`
- **THEN** `site.MacroPos()` MUST 为 embed 字段位置；`DeclPattern` match MUST 以 enclosing `TypeSpec` 为根

### Requirement: site 内部 MatchMeta 槽（引擎专用）

每轮 expand 的 `site Syntax` 实现 MUST 含**内部 meta 槽**（不暴露为公开 API）。`site.Match(pattern)` 成功时 MUST 将完整 **`MatchMeta`**（`Bindings`、`MatchedSpan`、`Plan`；MAY 含 `MatchRoot`）写入该槽。公开 `Syntax` 接口 MUST NOT 提供读取 `MatchMeta` 的方法；`Match` MUST NOT 向作者返回 `MatchMeta`。

引擎包（`internal/expander`）MUST 提供 `MatchMetaFromSite(site Syntax) (MatchMeta, bool)` 供 expand 流水线读取；该函数 MUST NOT 位于 provider 可 import 的公开路径。Call adapter 在 `TargetToPlan` 成功后 MUST 通过同一机制写入 meta 槽（无 `site.Match` 时）。

`SyntaxRules` / `SyntaxCase` 在 runtime match 或 fender 失败时 MUST 清空 meta 槽；仅最终胜出 clause 的 meta 可保留至 `Expander` 返回。`Clause.Plan` override 在 match 成功后 MUST 覆盖槽内 `Plan`。

#### Scenario: Match 成功写入 meta 槽

- **WHEN** `site.Match("$lhs ... := Try($inner)")` 成功
- **THEN** meta 槽 MUST 含 `MatchedSpan`、非空 `Plan` 与对应 `Bindings`；作者调用方仍只收到 `Bindings`

#### Scenario: meta 槽为空则引擎失败

- **WHEN** `Expander` 返回 `out` 但 meta 槽未写入（如裸 `Expander` 未调用 `site.Match`，见 design D19）
- **THEN** 引擎 MUST 在 `ValidateSplice` 前返回非 nil error；MUST NOT 调用 `Apply`

#### Scenario: 裸 Expander 经 Match 写入 meta

- **WHEN** 手写 `Expander` 在返回 `out` 前调用 `site.Match(pattern)` 且成功
- **THEN** meta 槽 MUST 已写入；引擎 MUST 与 `SyntaxCase` 路径同样执行 `ValidateSplice` / `Apply`

#### Scenario: fender 失败清空 meta 槽

- **WHEN** 某 clause match 成功但 fender 返回 error，下一 clause 随后 match 成功
- **THEN** 传入 `ValidateSplice` 的 meta MUST 来自胜出 clause，MUST NOT 残留失败 clause 的 `Plan`
