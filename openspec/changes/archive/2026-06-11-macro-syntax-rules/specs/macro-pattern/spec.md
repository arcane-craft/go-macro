## ADDED Requirements

### Requirement: Pattern 顶层形式（首版封闭子集）

每个 normative pattern MUST 属于下列顶层形式之一（见 design D17）；解析失败 MUST 在注册 `SyntaxRules` / `SyntaxCase` 时 fatal。

| 顶层 | 语法概要 | `MatchRoot` | `MatchedSpan` |
|------|----------|-------------|---------------|
| **CallPattern** | `Callee` `(` ArgList `)` | `Call` | anchor `*ast.CallExpr` |
| **StmtPattern** | `$lhs ... :=` / `$lhs ... =` / `var $lhs ... =` / Return / `CallPattern` `;` 形 stmt | `Stmt` | 整条 `ast.Stmt`（`var` 形为 `*ast.DeclStmt`） |
| **DeclPattern** | `type` `$item` IDENT `struct` `{` DeclFieldList `}` | `Decl` | `*ast.TypeSpec` |

首版 **MUST NOT** 支持上述以外的顶层（如独立 `ExprPattern`、函数体 block pattern）。

#### Scenario: Call 级顶层

- **WHEN** pattern 为 `Try($inner)` 且 anchor 为 stub call
- **THEN** MUST 解析为 CallPattern；`MatchRoot` MUST 为 `Call`

#### Scenario: Stmt 级 assign 顶层（`:=`）

- **WHEN** pattern 为 `$lhs ... := Try($inner)`
- **THEN** MUST 解析为 StmtPattern；`MatchRoot` MUST 为 `Stmt`

#### Scenario: Stmt 级 assign 顶层（`=`）

- **WHEN** pattern 为 `$lhs ... = Try($inner)`
- **THEN** MUST 解析为 StmtPattern；`MatchRoot` MUST 为 `Stmt`

#### Scenario: Stmt 级 var 顶层

- **WHEN** pattern 为 `var $lhs ... = Try($inner)`
- **THEN** MUST 解析为 StmtPattern；`MatchRoot` MUST 为 `Stmt`；`MatchedSpan` MUST 为含该 `var` 的 `*ast.DeclStmt`

#### Scenario: Decl 级顶层

- **WHEN** pattern 为 `type $item struct { Derive[$iface] $field ... }`
- **THEN** MUST 解析为 DeclPattern；`MatchRoot` MUST 为 `Decl`

### Requirement: Callee literal 按 invoked name 匹配

CallPattern 与 DeclPattern 中的 **stub / marker literal**（如 `Try`、`Derive`）MUST 按 **invoked name** match anchor：

- callee 为 `*ast.Ident` 时 MUST 比较 `Name`
- callee 为 `*ast.SelectorExpr` 时 MUST 比较 `Sel.Name`

pattern `Try($inner)` MUST match 源码 `Try(...)` 与 `tr.Try(...)`。pattern `tr.Try($inner)` MAY 额外要求 selector 左端为 `tr`（import alias 敏感）。

#### Scenario: alias 调用 match 短 literal

- **GIVEN** 源码 `tr.Try(f())`，anchor 为该 call，pattern `Try($inner)`
- **WHEN** `site.Match(pattern)` 执行
- **THEN** MUST match 成功且 `$inner` 绑定 `f()`

#### Scenario: invoked name 匹配 Derive embed

- **GIVEN** 源码 `type T struct { provider.Derive[Stringer] }`，pattern 含 literal `Derive[$iface]`
- **WHEN** DeclPattern match 执行
- **THEN** MUST match 成功；`$iface` MUST 绑定 `Stringer` 类型实参

### Requirement: Pattern 捕获与 literal

Pattern 语言 MUST 使用 `$` 前缀表示 capture（`$name`）。MUST 支持 `$_` 表示匹配但不绑定。MUST NOT 提供 `Clause.Literals` 字段；pattern 中 **非 `$` 前缀** 的标识符 token MUST 作为 **literal** 与源码文本一致 match（如 `Try`、`return`）。

#### Scenario: capture 与 literal 区分

- **WHEN** pattern 为 `Try($inner)` 且 site anchor 为 `Try(f())`
- **THEN** MUST match 成功且 `inner` 绑定 `f()` 子树；`Try` MUST NOT 进入 Bindings

#### Scenario: 忽略绑定

- **WHEN** pattern 为 `$name $_ $_` 且 field 为 `Name string`
- **THEN** MUST match 成功且仅 `name` 出现在 Bindings

### Requirement: Ellipsis

Pattern MUST 支持 `$name ...`（标识符、空格、三点）表示重复段。首版 normative ellipsis 位置：

| 位置 | 示例 | 绑定内容 |
|------|------|----------|
| Decl 具名字段 | `$field ...` | struct 内全部具名字段（见 Decl 无序约束） |
| Assign lhs（`:=` / `=`） | `$lhs ...` | `AssignStmt.Lhs` 全部项 |
| `var` 名字列表 | `$lhs ...`（`var $lhs ... =` 形） | `GenDecl` 内 `ValueSpec.Names` 全部项 |
| Return 前缀 results | `$vals ...` | anchor 之前的 `ReturnStmt.Results` 前缀 |

Match 结果 MUST 经 **`Bindings.Elems(name) ([]Syntax, bool)`** 读取；0 项时 MUST 返回空切片与 `true`。迭代顺序 MUST 为源码 AST 子节点顺序。

#### Scenario: struct 多 field ellipsis

- **WHEN** pattern 为 `type $item struct { Derive[$iface] $field ... }` 且 struct 含 2 个具名字段（embed 任意位置）
- **THEN** `Bindings.Elems("field")` MUST 长度为 2

#### Scenario: assign lhs ellipsis（`:=`）

- **WHEN** pattern 为 `$lhs ... := Try($inner)` 且源码为 `x, err := Try(f())`
- **THEN** `Elems("lhs")` MUST 长度为 2，每项 `Underlying()` 为对应 `ast.Expr`

#### Scenario: assign lhs ellipsis（`=`）

- **WHEN** pattern 为 `$lhs ... = Try($inner)` 且源码为 `x, err = Try(f())`
- **THEN** `Elems("lhs")` MUST 长度为 2，每项 `Underlying()` 为对应 `ast.Expr`

#### Scenario: var 名字 ellipsis

- **WHEN** pattern 为 `var $lhs ... = Try($inner)` 且源码为 `var x, err = Try(f())`
- **THEN** `Elems("lhs")` MUST 长度为 2；`MatchedSpan` MUST 为 enclosing `*ast.DeclStmt`

#### Scenario: return 前缀 ellipsis

- **WHEN** pattern 为 `return $vals ... , Try($inner)` 且源码为 `return a, Try(f())`
- **THEN** `Elems("vals")` MUST 长度为 1 且绑定 `a`；`$inner` MUST 绑定 `f()`

### Requirement: 绑定 Underlying 形状

Match 捕获的 `Syntax` 绑定 **`Underlying()` 形状 MUST 确定**，供 Decl / Call Transform 与 `ctx.Types()` 组合使用（见 decl-macro、design D16）。

| 捕获 | `Underlying()` MUST 为 |
|------|------------------------|
| `$item`（type pattern 中类型名） | `*ast.TypeSpec` 或等价 decl 根 |
| `$field`（struct field ellipsis 每项） | `*ast.Field`（含 `Names`、`Type`、**`Tag`**；`Tag` 与 `Type` 平级，均属同一 field 节点） |
| `$iface` 等（`Derive[$iface]`、index 类型实参） | type `ast.Expr`（如 `*ast.Ident`、`*ast.SelectorExpr`） |
| `$lhs`（lhs ellipsis 每项） | `ast.Expr`（多为 `*ast.Ident`） |
| `$vals`（return ellipsis 每项） | `ast.Expr` |
| `$inner` 等（call / expr 捕获） | 对应 `ast.Expr` 子树 |

首版 **MUST NOT** 提供 pattern tag 字面量语法（如 `` `macro:"k=v"` ``）；struct tag 内容 **MUST** 经 `*ast.Field.Tag` + `macro.ParseMacroTag` 在 Transform 内读取。首版 Decl struct **MUST NOT** 支持逐字段 `$name $type` pattern；具名字段 **MUST** 经 `$field ...` ellipsis 捕获。

#### Scenario: field 绑定含 Tag

- **GIVEN** 源码 `type Item struct { Name string \`json:"name"\` }`，pattern `type $item struct { $field ... }`
- **WHEN** match 成功且 `binds.Elems("field")` 含该项
- **THEN** 该项 `Underlying()` MUST 为 `*ast.Field`，且 `.Tag` MUST 非 nil

#### Scenario: 类型实参绑定

- **WHEN** 源码含 `Derive[Stringer]` 且 pattern 含 `Derive[$iface]`
- **THEN** `binds.Get("iface").Underlying()` MUST 为表示 `Stringer` 的 `ast.Expr`

### Requirement: Anchored match

**Call 站点**：`site.Match(pattern)` MUST 将 pattern 中与当前 stub 同形的 Call 节点 **仅** unify 到本轮 anchor `*ast.CallExpr`。同句其它未展开宏 Call MUST NOT 绑定到本轮 capture。

**Decl 站点**（design D18）：anchor 为 embed `*ast.Field`；`DeclPattern` MUST 在 enclosing `*ast.TypeSpec` 上 match；marker embed MUST 经无序约束集与 anchor 所在 embed 对齐；**不**适用 Call unify 规则。

#### Scenario: 同句两 Try 第一轮

- **WHEN** 源码为 `a, b := Try(f()), Try(g())` 且第一轮 anchor 为 `Try(g())`，pattern 为 `$lhs ... := Try($inner)`
- **THEN** `$inner` MUST 绑定 `g()` 而非 `f()`

### Requirement: Decl struct 顺序无关约束集

DeclPattern 的 struct `FieldList` **MUST** 作**无序约束集** match（见 design D17）；**MUST NOT** 要求 pattern 内元素书写顺序与源码 `Fields.List` 顺序一致。

- `EmbedMarker[$capture]`（如 `Derive[$iface]`）：MUST 存在**恰好一个**匿名 embed（`Names == nil`），invoked name 与 literal 一致。
- `$field ...`：MUST 绑定 struct 内**全部**具名字段，**不含**已匹配 embed；MAY 为 0 项。
- pattern 内 `{ Derive[$iface], $field ... }` 与 `{ $field ..., Derive[$iface] }` **MUST** 语义等价。
- 首版 **MUST NOT** 支持同一 struct 内多个 marker embed。

#### Scenario: embed 在具名字段之前

- **GIVEN** 源码 `type Item struct { provider.Derive[Stringer]; Name string }`，pattern `type $item struct { Derive[$iface] $field ... }`
- **WHEN** match 执行
- **THEN** MUST 成功；`iface` 与 `field` 绑定 MUST 正确

#### Scenario: embed 在具名字段之后

- **GIVEN** 源码 `type Item struct { Name string; provider.Derive[Stringer] }`，pattern `type $item struct { $field ... Derive[$iface] }`
- **WHEN** match 执行
- **THEN** MUST 成功；结果 MUST 与 embed 在前情形绑定等价（`Elems("field")` 长度均为 1）

#### Scenario: 仅 embed 无具名字段

- **GIVEN** 源码 `type Item struct { Derive[Stringer] }`，pattern `type $item struct { Derive[$iface] }`
- **WHEN** match 执行
- **THEN** MUST 成功；`Elems("field")` MAY 为空

### Requirement: MatchedSpan 贴回边界

`site.Match(pattern)` MUST 记录 **`MatchedSpan`**：pattern 成功 match 的语法子树（或引擎可确定的等价 span）。**MatchedSpan MUST 由 pattern 划定**；引擎 MUST NOT 为 Call/Decl 强加默认贴回粒度。

Match 结果（`MatchMeta`）MUST 写入 **`site` 内部 meta 槽**（见 macro-syntax、design D15），向引擎暴露 **MatchedSpan** 与 **`Plan []SpliceStep`**，供 `ValidateSplice` 与 `Apply` 使用。未落入 MatchedSpan 的 AST 节点 MUST NOT 被 Apply 修改或删除。

#### Scenario: stmt 级 pattern 的 MatchedSpan

- **WHEN** pattern 为 `$lhs ... := Try($inner)` 且 match 成功
- **THEN** MatchedSpan MUST 为整条 `AssignStmt`

#### Scenario: ExprStmt 语法糖 stmt 级 MatchedSpan

- **WHEN** pattern 为 `Try($inner);` 且 anchor 位于 `Try(f());` 形 `ExprStmt`
- **THEN** MatchedSpan MUST 为整条 `ExprStmt`（非仅 call 子树）

#### Scenario: type 级 pattern 的 MatchedSpan

- **WHEN** pattern 为 `type $item struct { $field ... Derive[$iface] }` 且 match 成功
- **THEN** MatchedSpan MUST 为对应 `TypeSpec`（或 pattern 实际覆盖的 decl 子树）

#### Scenario: 未 match 的 methods 不在 span 内

- **WHEN** 文件含 `type T struct { ... }` 与 `func (T) Foo()`，pattern 仅 match `type $item struct { ... }`
- **THEN** MatchedSpan MUST NOT 包含 `func (T) Foo()` 的 `FuncDecl`

### Requirement: MatchRoot 辅助元数据

Match 实现 MAY 记录 `MatchRoot`（`Stmt` | `Call` | `Decl`）供错误信息与调试。`MatchRoot` MUST NOT 作为贴回边界或 `Plan` 的权威来源；权威来源 MUST 为 **MatchedSpan** 与 Match 时确定的 **Plan**。

Stmt 级 pattern（如 `$lhs ... := Try($inner)`）MUST 标记 `MatchRoot=Stmt`；仅 match anchor Call 的 pattern MUST 标记 `MatchRoot=Call`；decl 根 pattern MUST 标记 `MatchRoot=Decl`。

#### Scenario: assign 整句 pattern

- **WHEN** pattern 为 `$lhs ... := Try($inner)`
- **THEN** MatchRoot MUST 为 `Stmt` 且 MatchedSpan MUST 为 enclosing AssignStmt

#### Scenario: 仅 Call pattern

- **WHEN** pattern 为 `Try($inner)` 且仅 match anchor
- **THEN** MatchRoot MUST 为 `Call` 且 MatchedSpan MUST 为 anchor CallExpr

### Requirement: Match 产出 SplicePlan

`site.Match(pattern)` 成功时 MUST 在 `MatchMeta` 中填入 **`Plan []SpliceStep`**（见 macro-splice-apply）。`Plan` MUST 在 Match 时完全确定；MUST NOT 延迟至 Expander 返回 `out` 后再推断。

**推导规则**（design D17）：

| 顶层 | Plan |
|------|------|
| **StmtPattern**（含 `$lhs ... :=` / `$lhs ... =` / `var $lhs ... =` / return / `Try(...);`） | 单步 `ReplaceInContainer{BlockStmts, OneToMany}` |
| **CallPattern** | 单步，由 anchor **直接父槽位**决定：`AssignRhs` / `ReturnResults` / `ExprSlot` 等 `OneToOne` |
| **DeclPattern** | `GenDeclSpecs` OneToOne + `InsertAfterInFileDecls` |

CallPattern 在 `ExprStmt` 父上下文中 MUST 使用 `ExprSlot` OneToOne（**不得**自动提升为 stmt 级）；stmt 级 1→N MUST 使用 StmtPattern `Try($inner);`。

注册期：仅当**同一顶层形式**在**同一父上下文类**下可导出两种 Plan 时 MUST fatal。不同父上下文（assign vs return）下 CallPattern Plan 不同 **MUST NOT** 视为注册歧义。

#### Scenario: type pattern 含两步 Plan

- **WHEN** pattern 为 `type $item struct { Derive[$iface] $field ... }` 且 match 成功
- **THEN** `MatchMeta.Plan` MUST 含 `GenDeclSpecs` 替换与 `InsertAfterInFileDecls` 两步

#### Scenario: ExprStmt 中 CallPattern 窄 Plan

- **GIVEN** 源码 `Try(f());`，pattern `Try($inner)`（CallPattern）
- **WHEN** match 成功
- **THEN** `Plan` MUST 为 `ExprSlot` OneToOne；`MatchedSpan` MUST 为 anchor call（非整条 `ExprStmt`）

#### Scenario: ExprStmt 分号语法糖宽 Plan

- **GIVEN** 同上源码，pattern `Try($inner);`（StmtPattern）
- **WHEN** match 成功
- **THEN** `Plan` MUST 为 `BlockStmts` OneToMany；`MatchedSpan` MUST 为 `ExprStmt`
