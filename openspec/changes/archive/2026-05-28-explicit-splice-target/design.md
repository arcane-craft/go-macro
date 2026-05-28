# Design: explicit-splice-target

## Context

今日数据流：

```text
classifySiteInFile(call) → CallSiteKind
Expander → ExpandResult { Stmts?, Exprs?, Expr? }  // 隐式：由 Site 决定用哪个
ApplyExpandResult(file, call, site, result)       // 矩阵查表
```

问题：作者读 `ExpandResult` 时不知道「换的是哪颗 AST 子树」；`SiteReturn`+`Exprs` vs `Stmts` 易混；`mactest.Expand` 只返回 result，不验证贴回合法性。

已确认约束：显式 Target（E1）、D1 仅换 assign RHS、可读性与早校验并重（C）。

现有实现：`internal/expander/splice.go` 的 `replaceCallExpr` 已覆盖多种表达式父节点；`ReplaceAssignRHS` 可复用「在 `AssignStmt.Rhs` 中按指针替换」逻辑，不必替换整条 `AssignStmt`。

## Goals / Non-Goals

**Goals:**

- `macro.SpliceTarget` 枚举 + `ExpandResult.Target` 必填
- 六种贴回目标与载荷规则表（见 spec）
- 从 `call` 解析**结构锚点**，校验 `Target` 与锚点一致
- `Context.LegalSpliceTargets()` 供 Expander / `mactest` 使用
- `mactest` 提供 `ValidateExpandResult(ctx, result) error`（或合并进 Expand 返回值路径）
- 更新 author-guide、本仓所有 Expander 与测试

**Non-Goals:**

- 通用 AST path patch
- 在 `if cond` 等表达式位 splice 多条语句（不新增 `SpliceReplaceIfCond` 等）
- expand 后全文件二次 typecheck（可选 follow-up）
- contrib 仓实现细节（仅 tasks 跟踪）

## Decisions

### D1：`SpliceTarget` 命名与枚举

```go
type SpliceTarget int

const (
    SpliceReplaceAssignStmt SpliceTarget = iota
    SpliceReplaceAssignRHS
    SpliceReplaceReturnStmt
    SpliceReplaceReturnResults
    SpliceReplaceExprStmt
    SpliceReplaceCallExpr
)
```

**理由**：动词+锚点，作者读 `Target` 即知替换范围。保留 `CallSiteKind`（`SiteAssign` 等）作**语境提示**，但 splice **只认 `Target`**。

### D2：载荷规则（与 Target 绑定）

| `Target` | 必填载荷 | 禁止 |
|----------|----------|------|
| `SpliceReplaceAssignStmt` | `len(Stmts)>0` | 非空 `Expr`/`Exprs` |
| `SpliceReplaceAssignRHS` | `Expr != nil` | 非空 `Stmts`/`Exprs` |
| `SpliceReplaceReturnStmt` | `len(Stmts)>0` | 非空 `Expr`/`Exprs` |
| `SpliceReplaceReturnResults` | `len(Exprs)>0` | 非空 `Stmts`/`Expr` |
| `SpliceReplaceExprStmt` | `len(Stmts)>0` | 非空 `Expr`/`Exprs` |
| `SpliceReplaceCallExpr` | `Expr != nil` | 非空 `Stmts`/`Exprs` |

**理由**：消除「多字段同时非空谁赢」的歧义；校验可在 `ApplyExpandResult` 与 `mactest.ValidateExpandResult` 共用 `macro.ValidateExpandResult`。

### D3：`ReplaceAssignRHS` 实现

1. `findEnclosingBlockStmt` 得到 `*ast.AssignStmt`
2. 在 `assign.Rhs` 中找 `rhs == call`（或 `unwrapParen(rhs)==call`）的索引 `i`
3. `assign.Rhs[i] = result.Expr`
4. 若宏不在任何 `Rhs` 槽位 → 报错

**多值 assign** `a, b := Macro()`：`Lhs` 长度为 2，`Rhs` 仍为单个 `CallExpr`；替换为另一 `CallExpr`（或括号表达式）返回两值。**不**用 `Exprs` 展开为多个 `Rhs` 元素（Go 语法不允许 `a, b := e1, e2` 与单宏混用时的特殊规则除外，首版仅支持替换**一个**含宏的 RHS 槽位）。

**备选**：`Exprs` 填满整个 `Rhs` → 改变 assign 形态，与 D1「保留 Lhs」部分冲突；不采用。

### D4：结构锚点 → `LegalSpliceTargets()`

对宏调用 `call` 扫描 enclosing 语句（与今日 `classifySiteInFile` 类似，但更细）：

| 结构条件 | 合法 Target |
|----------|-------------|
| `call` 为某 `AssignStmt.Rhs` 元素 | `SpliceReplaceAssignRHS`, `SpliceReplaceAssignStmt` |
| `call` 为 `ReturnStmt.Results` 元素 | `SpliceReplaceReturnResults`, `SpliceReplaceReturnStmt` |
| `call` 为 `ExprStmt.X` | `SpliceReplaceExprStmt` |
| 否则（表达式槽，含嵌套于 `BinaryExpr` 等） | `SpliceReplaceCallExpr` |

**注意**：assign/return 同时提供「整条语句」与「局部」两种 Target，由**作者选择**；引擎在 apply 时校验 `Target` 属于上表且与载荷匹配。

`ctx.Site()` 仍可映射为「默认推荐」语境（文档表），但 MUST NOT 单独决定 splice。

### D5：`ApplyExpandResult` 签名

```go
func ApplyExpandResult(file *ast.File, call *ast.CallExpr, result macro.ExpandResult) error
```

移除对 `site` 参数的依赖；内部 `macro.ValidateExpandResult(call, file, result)` + 锚点检查。

错误文案示例：

```text
macro: SpliceReplaceCallExpr invalid at assign RHS; legal targets: SpliceReplaceAssignRHS, SpliceReplaceAssignStmt
```

### D6：`mactest`

- 保留 `mactest.Expand(...)` 返回 `(ExpandResult, error)`
- 新增 `mactest.Validate(ctx, result) error` 调用与引擎相同的 `macro.ValidateExpandResult`
- 文档建议：`result, err := mactest.Expand(...); mactest.Validate(ctx, result)`

**理由**：满足「单测尽早失败」且不与 Expand 语义耦合（Expand 仍允许 provider 返回 error）。

### D7：破坏性迁移（E1）

- 删除「按 Site 选字段」分支；所有 `ExpandResult` 构造处加 `Target`
- `init provider` 脚手架生成的占位 Expander 设置合理默认 `Target`（如 `SpliceReplaceCallExpr`）
- contrib：`Try` 类 assign 若需保留 LHS，改用 `SpliceReplaceAssignRHS`；语句化 return 用 `SpliceReplaceReturnStmt`

### D8：`CallSiteKind` 保留

不删除 `ctx.Site()`；provider 仍可用其做语义分支（如禁止某 Site）。spec 明确：**贴回不以 Site 为准**。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 作者选 `SpliceReplaceAssignStmt` 却只想换 RHS | 文档 + `LegalSpliceTargets` 引导；assign 场景优先推荐 RHS Target |
| `ReplaceAssignRHS` 替换后类型不匹配 | 保持现有 types.Info 时机；文档说明责任在 provider；可选 follow-up 再 Check |
| contrib 与本仓版本漂移 | tasks 单列 contrib PR；README 注明最低 go-macro 版本 |
| 测试量大 | 表驱动 `splice_test.go` 按 Target 分节 |

## Migration Plan

1. 本仓：先 `macro` API + `ValidateExpandResult`，再 `ApplyExpandResult`，再修全仓测试与 `cmd/macro` 示例。
2. 发布带 **BREAKING** 的 go-macro tag（或文档声明 main 破坏性变更）。
3. contrib 跟进：各 `*Expand` 设 `Target`；Try assign 路径评估是否改为 `SpliceReplaceAssignRHS`。
4. 归档 change 合并 spec 至 `openspec/specs/`。

## Open Questions

（无。C / D1 / E1 已在 proposal Boundaries 闭合。）
