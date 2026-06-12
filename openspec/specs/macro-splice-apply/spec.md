# macro-splice-apply Specification

## Purpose
TBD - created by archiving change macro-syntax-rules.
## Requirements
### Requirement: Match 产出 SplicePlan

`site.Match(pattern)` 成功时，Match 结果（`MatchMeta`）MUST 除 `MatchedSpan`、`Bindings` 与 MAY 含 `MatchRoot` 外，包含 **`Plan []SpliceStep`**。`Plan` MUST 在 Match 时完全确定；引擎 MUST NOT 在 Expander 返回 `out` 后再根据 Call/Decl 种类或 `out` 形状 **推断** 贴回方式（无 normative `InferTarget`）。

`SpliceStep` 首版 MUST 支持两类原语：

- **`ReplaceInContainer`**：`Parent`、`ContainerField`（`BlockStmts` | `AssignRhs` | `ReturnResults` | `GenDeclSpecs` | `ExprSlot` 等）、`Index`（`ReplaceAll` 时省略）、`Mode`（`OneToOne` | `OneToMany` | **`ReplaceAll`**）
- **`InsertAfterInFileDecls`**：`After *ast.GenDecl`（MUST 为含 `MatchedSpan` TypeSpec 的 GenDecl）

`MatchedSpan` MUST remain 贴回语义边界：`Plan` 的每一步 MUST NOT 修改或删除未落入 `MatchedSpan` 语义范围的 AST，**除** `InsertAfterInFileDecls` 在 `file.Decls` 中 **追加** `out` 载荷中的 **新生成** decl 所必需之插入。

对 normative pattern 集合，给定 anchor site 成功 Match MUST 产出 **唯一** `Plan`。注册期 fatal 规则见 macro-pattern「Match 产出 SplicePlan」：仅**同一顶层形式 + 同一父上下文类**存在多种 Plan 时 fatal；CallPattern 在 assign vs return 下 Plan 不同 **MUST NOT** fatal。

#### Scenario: stmt 级 pattern 产出 block 替换 plan

- **GIVEN** 源码 `lhs, err := Try(inner)`，anchor 为该 `Try` 调用，pattern 为 `$lhs ... := Try($inner)`
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `MatchedSpan` MUST 为 enclosing `*ast.AssignStmt`
- **AND** `Plan` MUST 恰含一步 `ReplaceInContainer{ContainerField: BlockStmts, Mode: OneToMany}`，其 `Parent` 为 enclosing `*ast.BlockStmt`，`Index` 为该 `AssignStmt` 在 `Block.List` 中的下标

#### Scenario: call 级 pattern 产出 RHS 替换 plan

- **GIVEN** 源码 `lhs := Try(inner)`，anchor 为 `Try(inner)`，pattern 为 `Try($inner)`
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `MatchedSpan` MUST 为 anchor `*ast.CallExpr`
- **AND** `Plan` MUST 恰含一步 `ReplaceInContainer{ContainerField: AssignRhs, Mode: OneToOne}`

#### Scenario: type 级 pattern 产出两步 plan

- **GIVEN** 源码含 `type Item struct { …; Derive[I] }`（embed 与 fields 任意顺序），anchor 为 embed 处，pattern 为 `type $item struct { Derive[$iface] $field ... }`
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `MatchedSpan` MUST 为对应 `*ast.TypeSpec`
- **AND** `Plan` MUST 按顺序含：`ReplaceInContainer{ContainerField: GenDeclSpecs, Mode: OneToOne}` 与 `InsertAfterInFileDecls{After: 含该 TypeSpec 的 *ast.GenDecl}`

#### Scenario: return stmt 级 pattern 产出 block 替换 plan

- **GIVEN** 源码 `return a, Try(f())`，anchor 为 `Try(f())`，pattern 为 `return $vals ... , Try($inner)`
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `MatchedSpan` MUST 为 enclosing `*ast.ReturnStmt`
- **AND** `Plan` MUST 为 `BlockStmts` OneToMany

#### Scenario: assign `=` stmt 级 pattern 产出 block 替换 plan

- **GIVEN** 源码 `x, err = Try(f())`，anchor 为 `Try(f())`，pattern 为 `$lhs ... = Try($inner)`
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `MatchedSpan` MUST 为 enclosing `*ast.AssignStmt`（`Tok=ASSIGN`）
- **AND** `Plan` MUST 为 `BlockStmts` OneToMany

#### Scenario: var stmt 级 pattern 产出 block 替换 plan

- **GIVEN** 源码 `var x, err = Try(f())`，anchor 为 `Try(f())`，pattern 为 `var $lhs ... = Try($inner)`
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `MatchedSpan` MUST 为 enclosing `*ast.DeclStmt`
- **AND** `Plan` MUST 为 `BlockStmts` OneToMany

#### Scenario: ExprStmt 中 CallPattern 产出 ExprSlot plan

- **GIVEN** 源码 `Try(f());`，anchor 为 `Try(f())`，pattern 为 `Try($inner)`（CallPattern，无分号）
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `Plan` MUST 为 `ExprSlot` OneToOne；`MatchedSpan` MUST 为 anchor `*ast.CallExpr`

#### Scenario: Match 无法构造合法 plan 则 match 失败

- **GIVEN** pattern 为 `Try($inner)`，但 anchor 所在 parent 不在引擎支持的 `ContainerField` 集合内
- **WHEN** `site.Match(pattern)` 执行
- **THEN** MUST 返回 match 失败（不得 silent 降级为更粗粒度 span）

#### Scenario: 同一顶层形式在同一父上下文类不得对应多种 Plan

- **GIVEN** 某 `Clause.Pattern` 为 CallPattern，且静态分析表明在 **assign RHS 父上下文类** 下可导出两种不同 `Plan`
- **WHEN** 注册 `SyntaxRules` / `SyntaxCase`
- **THEN** MUST fatal（与 pattern 解析失败同级）

#### Scenario: 不同父上下文类 Plan 不同非 fatal

- **GIVEN** CallPattern `Try($inner)` 在 assign RHS 与 return results 下分别导出 `AssignRhs` 与 `ReturnResults` Plan
- **WHEN** 注册该 pattern
- **THEN** MUST NOT fatal（运行时父链唯一决定 Plan）

CallPattern 在 anchor 位于 **表达式子树内**（非 assign/return/ExprStmt 的直接子节点，如 `x := foo + Try(y)`、`return wrap(Try(y))`、`outer(Try(y));`）时，Plan MUST 仍为对应容器字段 + **`OneToOne`**（含 `Index` 当适用）；**MUST NOT** 引入额外 `SpliceMode`。`MatchedSpan` MUST 仍为 anchor `*ast.CallExpr`。

ReturnResults 下 CallPattern Plan MUST 按 anchor 与 result 槽位关系二选一：

| 关系 | Plan | `out` 形状（Validate） |
|------|------|------------------------|
| anchor 为某 `ReturnStmt.Results[i]` 的**直接**子表达式（含 unwrap paren） | `ReturnResults` + **`ReplaceAll`** | `ToExprs()`，`len ≥ 1` |
| anchor 嵌于某 `Results[i]` 表达式树内（非直接子节点） | `ReturnResults` + **`OneToOne`**，`Index=i` | `ToExpr()` |

#### Scenario: call 级 pattern — return 直接 result 产出 ReplaceAll plan

- **GIVEN** 源码 `return Try(f())`，anchor 为 `Try(f())`，pattern 为 `Try($inner)`（CallPattern）
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `MatchedSpan` MUST 为 anchor `*ast.CallExpr`
- **AND** `Plan` MUST 为 `ReplaceInContainer{ContainerField: ReturnResults, Mode: ReplaceAll}`

#### Scenario: call 级 pattern — return 嵌套 result 产出 OneToOne plan

- **GIVEN** 源码 `return a, wrap(Try(f()))`，anchor 为内层 `Try(f())`，pattern 为 `Try($inner)`（CallPattern）
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `MatchedSpan` MUST 为 anchor `*ast.CallExpr`
- **AND** `Plan` MUST 为 `ReplaceInContainer{ContainerField: ReturnResults, Mode: OneToOne, Index: 1}`

#### Scenario: call 级 pattern — assign RHS 嵌套表达式产出 OneToOne plan

- **GIVEN** 源码 `x := foo + Try(f())`，anchor 为 `Try(f())`，pattern 为 `Try($inner)`（CallPattern）
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `MatchedSpan` MUST 为 anchor `*ast.CallExpr`
- **AND** `Plan` MUST 为 `ReplaceInContainer{ContainerField: AssignRhs, Mode: OneToOne, Index: 0}`

#### Scenario: call 级 pattern — ExprStmt 嵌套 call 产出 OneToOne plan

- **GIVEN** 源码 `outer(Try(f()));`，anchor 为内层 `Try(f())`，pattern 为 `Try($inner)`（CallPattern）
- **WHEN** `site.Match(pattern)` 成功
- **THEN** `Plan` MUST 为 `ReplaceInContainer{ContainerField: ExprSlot, Mode: OneToOne}`；`MatchedSpan` MUST 为 anchor call

### Requirement: ValidateSplice 校验 out 与 Plan

Expander 返回 `out` 后，引擎 MUST 调用 **`ValidateSplice(out Syntax, meta MatchMeta) error`**。Validate MUST 检查 `out` 形状与 `meta.Plan` 每一步兼容；MUST NOT 引入 Call/Decl 分支。

首版兼容规则：

| Plan 步骤 | out 要求 |
|-----------|----------|
| `ReplaceInContainer` + `BlockStmts` + `OneToMany` | `out.ToStmts()` 成功且 `len ≥ 1`（`MatchedSpan` 为 `AssignStmt`、`DeclStmt`、`ReturnStmt`、`ExprStmt` 等） |
| `ReplaceInContainer` + `AssignRhs` / `ExprSlot` + `OneToOne` | `out.ToExpr()` 成功 |
| `ReplaceInContainer` + `ReturnResults` + `ReplaceAll` | `out.ToExprs()` 成功且 `len ≥ 1` |
| `ReplaceInContainer` + `ReturnResults` + `OneToOne` | `out.ToExpr()` 成功（anchor 嵌于 result 表达式内） |
| `ReplaceInContainer` + `GenDeclSpecs` + `OneToOne` | `out.ToDecls()` 成功，`decls[0]` 为 `*ast.TypeSpec` |
| `InsertAfterInFileDecls` | `out.ToDecls()[1:]` 每项为 `*ast.FuncDecl`，receiver 指向 MatchedSpan 类型名 |

#### Scenario: Try stmt 展开 — Validate 接受 ToStmts

- **GIVEN** `meta.Plan` 为单步 `BlockStmts` + `OneToMany`，`MatchedSpan` 为 `*ast.AssignStmt`
- **WHEN** `out.ToStmts()` 返回 3 条 stmt（init、`if err`、assign）
- **THEN** `ValidateSplice(out, meta)` MUST 成功

#### Scenario: Try call 展开 — Validate 接受 ToExpr

- **GIVEN** `meta.Plan` 为单步 `AssignRhs` + `OneToOne`，`MatchedSpan` 为 `*ast.CallExpr`
- **WHEN** `out.ToExpr()` 成功
- **THEN** `ValidateSplice(out, meta)` MUST 成功

#### Scenario: return 嵌套 call — Validate 接受 ToExpr

- **GIVEN** `meta.Plan` 为 `ReturnResults` + `OneToOne`（anchor 嵌于 `wrap(Try(f()))`），`MatchedSpan` 为内层 `*ast.CallExpr`
- **WHEN** `out.ToExpr()` 成功
- **THEN** `ValidateSplice(out, meta)` MUST 成功

#### Scenario: return 直接 call — Validate 要求 ToExprs

- **GIVEN** `meta.Plan` 为 `ReturnResults` + `ReplaceAll`，`MatchedSpan` 为 `*ast.CallExpr`
- **WHEN** `out` 仅可 `ToExpr()` 不可 `ToExprs()`
- **THEN** `ValidateSplice(out, meta)` MUST 失败

#### Scenario: Derive 展开 — Validate 接受 TypeSpec 与 methods

- **GIVEN** `meta.Plan` 含 `GenDeclSpecs` 替换与 `InsertAfterInFileDecls`
- **WHEN** `out.ToDecls()` 为 `[TypeSpec', FuncDecl(String)]`，且 `TypeSpec'.Name` 与 MatchedSpan TypeSpec 同名
- **THEN** `ValidateSplice(out, meta)` MUST 成功

#### Scenario: out 形状与 Plan 不符 — Validate 失败

- **GIVEN** `meta.Plan` 为 `AssignRhs` + `OneToOne`
- **WHEN** `out` 仅可 `ToStmts()` 不可 `ToExpr()`
- **THEN** `ValidateSplice(out, meta)` MUST 返回非 nil error

#### Scenario: Derive TypeSpec 改名 — Validate 失败

- **GIVEN** MatchedSpan 为 `type Item struct …`
- **WHEN** `out.ToDecls()[0]` 为 `type Other struct …`
- **THEN** `ValidateSplice(out, meta)` MUST 失败

#### Scenario: 生成 method 与已有同名 method 冲突 — Validate 失败

- **GIVEN** 文件中已有 `func (Item) String()`，且该 `FuncDecl` 不在 `MatchedSpan` 内
- **WHEN** `out.ToDecls()[1:]` 含 `func (Item) String()`
- **THEN** `ValidateSplice(out, meta)` MUST 失败

#### Scenario: Derive 不要求 out 含未 match 的既有 methods

- **GIVEN** 文件含 `func (Item) Foo()`，pattern 仅 match `type Item struct { … }`
- **WHEN** `out.ToDecls()` 为 `[TypeSpec', FuncDecl(Bar)]`，不含 `Foo()`
- **THEN** `ValidateSplice(out, meta)` MUST 成功

#### Scenario: Validate 失败不得 Apply

- **GIVEN** `ValidateSplice` 返回 error
- **WHEN** expand 流水线继续
- **THEN** MUST NOT 调用 `Apply`；MUST NOT 写回部分 AST

### Requirement: Apply 执行 SplicePlan

贴回 MUST 通过 **`Apply(file *ast.File, meta MatchMeta, out Syntax) error`** 完成。Apply MUST 按 `meta.Plan` 顺序执行；MUST 使用 Match 已记录的 parent/index 定位槽位；MUST NOT 要求 provider 再次 `ast.Inspect` 查找替换点。

`out` 在父节点上下文中的节点数 **MAY** 大于 `MatchedSpan`（无单独 Insert API）：Stmt 上下文 1 槽 → N stmt；Decl 上下文 TypeSpec 替换 + `InsertAfterInFileDecls` 追加新 FuncDecls。

Apply 完成后，对 stmt 级替换 MUST 调用 **`StampStmtPos(site.MacroPos(), …)`**（见 macro-codegen）。

Apply 对 `AssignRhs` / `ExprSlot` + `OneToOne` MUST 使用 **`MatchedSpan`（anchor `*ast.CallExpr`）** 在 Plan 定位的槽位表达式内替换 call 子树：槽位表达式与 anchor 相同时 MUST 等价于整槽替换；否则 MUST 仅替换含 anchor 的子树，保留外围表达式结构（如 `BinaryExpr`、`CallExpr` 实参等）。

`ReturnResults` + `OneToOne` MUST 对 `Results[Index]` 应用同上子树替换规则。`ReturnResults` + `ReplaceAll` MUST 以 `out.ToExprs()` 替换整条 `ReturnStmt.Results`。

#### Scenario: Try — 1 条 AssignStmt 替换为 3 条 stmt

- **GIVEN** `Plan = [ReplaceInContainer{BlockStmts, OneToMany, Index=i}]`，`MatchedSpan` 为 block 中第 i 条 `AssignStmt`
- **WHEN** `Apply` 执行且 `out.ToStmts()` 长度为 3
- **THEN** `block.List` 中原 `[i]` 单槽 MUST 被 3 条 stmt 占据
- **AND** 原 `MatchedSpan` 节点 MUST 不再出现在 AST 中

#### Scenario: Try call — 仅替换 RHS 中的 CallExpr

- **GIVEN** `Plan = [ReplaceInContainer{AssignRhs, OneToOne}]`，`MatchedSpan` 为 `Try(inner)`
- **WHEN** `Apply` 执行且 `out.ToExpr()` 为 `inner'`
- **THEN** `AssignStmt.Lhs` MUST 不变
- **AND** 对应 `Rhs` 槽 MUST 变为 `inner'`

#### Scenario: Try call — 嵌套 RHS 仅替换子树 CallExpr

- **GIVEN** 源码 `x := foo + Try(inner)`，`Plan = [ReplaceInContainer{AssignRhs, OneToOne, Index=0}]`，`MatchedSpan` 为 `Try(inner)`
- **WHEN** `Apply` 执行且 `out.ToExpr()` 为 `inner'`
- **THEN** `AssignStmt.Lhs` MUST 不变
- **AND** 对应 `Rhs[0]` MUST 变为 `foo + inner'`（或等价 AST）；`foo` 子树 MUST 保留

#### Scenario: Try call — return 嵌套 result 仅替换子树

- **GIVEN** 源码 `return a, wrap(Try(inner))`，`Plan = [ReplaceInContainer{ReturnResults, OneToOne, Index=1}]`，`MatchedSpan` 为 `Try(inner)`
- **WHEN** `Apply` 执行且 `out.ToExpr()` 为 `inner'`
- **THEN** `ReturnStmt.Results[0]`（`a`）MUST 不变
- **AND** `Results[1]` MUST 变为 `wrap(inner')`（或等价 AST）

#### Scenario: Try call — return 直接 result 替换全部 Results

- **GIVEN** 源码 `return Try(inner)`，`Plan = [ReplaceInContainer{ReturnResults, ReplaceAll}]`，`MatchedSpan` 为 `Try(inner)`
- **WHEN** `Apply` 执行且 `out.ToExprs()` 为 `[v, errExpr]`
- **THEN** `ReturnStmt.Results` MUST 变为 `[v, errExpr]`；MUST NOT 保留原 `Try(inner)` 单 result 形态

#### Scenario: Derive — 替换 TypeSpec 并插入新 methods

- **GIVEN** `Plan = [Replace GenDeclSpecs, InsertAfterInFileDecls]`
- **WHEN** `out.ToDecls()` 为 `[TypeSpec', FuncDecl(String)]`
- **THEN** GenDecl 内 MatchedSpan TypeSpec MUST 被 `TypeSpec'` 替换
- **AND** `String` method MUST 出现在含该 GenDecl 的 `file.Decls` 条目之后
- **AND** 文件中未 match 的 `func (Item) Foo()` MUST 仍存在

#### Scenario: Apply 不修改 Plan 范围外的 AST

- **GIVEN** 同文件另有无关 `func Other()` 与未 match 的 `func (Item) Foo()`
- **WHEN** Derive Apply 成功
- **THEN** `Other()` 与 `Foo()` MUST 与 Apply 前语义等价保留

#### Scenario: Apply 不重复 Inspect

- **GIVEN** `meta.Plan` 已含 `BlockStmts` 的 `Index`
- **WHEN** `Apply` 执行
- **THEN** MUST NOT 再次扫描 file 以查找 `AssignStmt`；MUST 使用 Plan 中已解析的 parent/index

### Requirement: MatchMeta 经 site 槽传递至 ValidateSplice / Apply

当 Expander 为 **`SyntaxRules` / `SyntaxCase`** 时，匹配成功 clause 的 **`MatchMeta`（含 Plan）** MUST 在 expand 过程中写入 **`site` 内部 meta 槽**（见 macro-syntax、design D15）。引擎在 `Expander` 返回后 MUST 从 `site` 读取 `MatchMeta` 并传入 `ValidateSplice` 与 `Apply`。Expander 返回的 `out` MUST NOT 携带贴回 `Plan`（Plan 来自 Match，不来自 `out`）。

#### Scenario: SyntaxCase Try clause 端到端

- **GIVEN** clause `Pattern: $lhs ... := Try($inner)`，`Transform` 返回 `out` 为多条 stmt
- **WHEN** expand 成功
- **THEN** 引擎 MUST 从 `site` 读取该 clause match 产生的 `meta.Plan` 并执行 Apply
- **AND** MUST NOT 调用 `InferTarget` 或读取 `CallExpandResult.Target`

#### Scenario: 无匹配 clause

- **GIVEN** 所有 clause `site.Match` 均失败
- **WHEN** `SyntaxCase` Expander 执行
- **THEN** MUST 返回 `no matching syntax rule`（或等价 error），且 MUST NOT 调用 Apply

### Requirement: Clause Plan override

`Clause` MAY 含 **`Plan []SpliceStep` override**。当非零时，match 成功 MUST 使用 override 替代默认由 pattern 推导的 `Plan`；`ValidateSplice` / `Apply` MUST 仍使用同一套规则。首版 MAY 仅用于 adapter 或极端 Transform。

#### Scenario: Plan override 用于 Call adapter

- **GIVEN** Call adapter 将旧 `CallExpandResult.Target` 编译为 `Plan`，并写入 `site` meta 槽（经 `TargetToPlan` 或 `Clause.Plan` override）
- **WHEN** `ValidateSplice` 通过
- **THEN** Apply MUST 按槽内 `meta.Plan` 执行

### Requirement: TargetToPlan（Call-only adapter）

系统 MAY 提供 **`TargetToPlan(file *ast.File, call *ast.CallExpr, result CallExpandResult) (MatchMeta, error)`** 与 **`CallExpandResultToSyntax(result) (Syntax, error)`**，**仅**用于 Call 宏 adapter。**MUST NOT** 提供 `DeclExpandResult` → `Plan` 的等价函数。

`TargetToPlan` MUST 与 today 六种 `SpliceReplace*` 行为一致（见 design D13 映射表）。**MUST NOT** 用于 normative `SyntaxRules` / `SyntaxCase` 路径（该路径 Plan 由 Match 产出）。

#### Scenario: ReturnResults 编译为 ReplaceAll

- **GIVEN** `result.Target == SpliceReplaceReturnResults` 且 `result.Exprs` 非空
- **WHEN** `TargetToPlan` 执行
- **THEN** `Plan` MUST 含 `ReplaceInContainer{ContainerField: ReturnResults, Mode: ReplaceAll}`

#### Scenario: Decl 无 TargetToPlan

- **GIVEN** 旧 `DeclExpander` 返回 `DeclExpandResult`
- **WHEN** 引擎展开 Decl 站点
- **THEN** MUST NOT 调用 `TargetToPlan`；MUST 要求 native `SyntaxCase` Expander

