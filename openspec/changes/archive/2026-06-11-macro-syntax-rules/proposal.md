## Why

宏作者 today 在 Expander 中手写 `go/ast`、手动查找 enclosing 语句与贴回 Target，成本高且与同句多宏、嵌套展开容易出错。`macro/quote` 仅解决「写 AST」，未解决「读调用点 pattern」与统一 Expander 模型。需要以 Scheme **`syntax-rules` / `syntax-case`** 为参照，用 **`Syntax` 统一 Match（`$`）与 Quote（`#`）**，**统一 Expander 签名**，**Match 时确定 SplicePlan**（无运行时 InferTarget），并 **BREAKING** 收敛 Call/Decl 双轨 API。

## What Changes

- 引入 **`macro.Syntax`**、**`macro.Bindings`**：统一读写 AST 片段；`Syntax.Underlying()` 供逃生；**Decl embed 元数据**（MacroTag、类型实参等）经 **Bindings + Underlying()** + `ParseMacroTag` / `Types().TypeOf`（无 `DeclSite` / `site.MacroTag()`，见 design D16）
- 引入 **Pattern Match**（normative 首版子集见 **design D17–D19**：`CallPattern` / `StmtPattern` / `DeclPattern`；assign 覆盖 `:=` / `=` / `var ... =`；`$capture`、`$_`、`$lhs ...` / `$field ...` / `$vals ...` ellipsis；裸 ident = literal；**invoked name** callee 匹配；Decl struct **顺序无关约束集**；**Decl anchor = embed `*ast.Field`**（D18））
- **`macro.Quote(template, map[string]Syntax)`** 合并 **`macro/quote` 子包**（无 `@kind{ }`，形状由 `To*` 决定）
- 引入 **`SyntaxRules` / `SyntaxCase`**（`Clause`：Pattern + Template / Transform / 可选 Fender / 可选 Plan override；**无 Literals 字段**）
- **BREAKING**：统一 **`Expander func(ctx Context, site Syntax) (Syntax, error)`**；不再区分 CallExpander / DeclExpander
- **BREAKING**：**`Context` 极简三字段**（`FileSet`、`Types`、`TempIdent`）；`MacroPos` 迁至 **`site`**
- 通用 **`EnclosingSignature` / `EnclosingResults` / `ZeroSyntax`**（Types + 内部 scope；不暴露 `*ast.FuncDecl`；**不** Try 特化 built-in）
- **`MatchedSpan` + `SplicePlan`**：pattern match 划定语义边界并在 Match 时产出 `[]SpliceStep`；**`ResolveSite` → Expander → 从 `site` 读 `MatchMeta` → `ValidateSplice` → `Apply`**；`MatchMeta` 经 **`site` 内部槽**传递（非作者 API）；pattern 推导多 Plan **注册 fatal**
- **BREAKING**：废弃 `CallExpandResult` 作者 API；**Call** 短期 **`TargetToPlan` adapter** 后删除
- **BREAKING**：废弃 `DeclExpander` / `DeclExpandResult`；**无 Decl adapter**，Decl provider **必须**迁移 `SyntaxCase`（方案 C）
- **BREAKING**：贴回 **仅替换 MatchedSpan**；out 节点数 MAY 大于 match（Try 多 stmt、Derive 多 decl）；**无** Insert API、**无** Decl 全量 Methods merge
- 更新 **`docs/author-guide.md`** 与 **`mactest`**（Call adapter + Decl 强制迁移指引）

## Capabilities

### New Capabilities

- `macro-syntax`: `Syntax`、`Bindings`、`To*`、`Underlying`、`site.MacroPos`、**site 内部 MatchMeta 槽**（引擎专用）
- `macro-pattern`: Match pattern 语法、anchor、ellipsis、`MatchedSpan`、`MatchRoot`、`SplicePlan`
- `macro-template`: `Quote`（`#` 洞）、合并原 `macro-quote`、注释保留
- `macro-rules`: `SyntaxRules`、`SyntaxCase`、`Clause` 语义与 clause 顺序
- `macro-splice-apply`: `SpliceStep`、`ValidateSplice`、`Apply(plan)`；取代 `InferTarget` / normative `SpliceTarget`
- `macro-enclosing-signature`: `EnclosingSignature`、`EnclosingResults`、`ZeroSyntax`

### Modified Capabilities

- `macro-core`: **BREAKING** Expander/Context；统一宏模型；移除 Call/Decl 双 Expander
- `macro-expander`: **BREAKING** 流水线、ResolveSite、ValidateSplice/Apply(plan)；adapter 过渡
- `macro-splice-inference`: **REMOVED**；合并至 `macro-splice-apply`
- `macro-quote`: **REMOVED** 独立子包要求；能力迁入 `macro-template`
- `decl-macro`: **MODIFIED** 统一 Expander/site；MatchedSpan 贴回；移除全量 Methods 与引擎 merge 要求
- `author-guide`: syntax-rules / syntax-case 作者指引；迁移说明
- `macro-codegen`: `StampStmtPos` 使用 `site.MacroPos()`

## Impact

- **代码**：`macro/syntax.go`、`match.go`、`quote.go`（合并 quote）、`rules.go`、`types.go`、`splice.go`；`internal/expander/resolve.go`、**`meta.go`**、`scope.go`、`expand.go`；adapter 层
- **API**：**BREAKING** 公开 Expander/Context；`macro/quote` deprecated → 删除
- **测试**：pattern/quote/rules/splice-apply/enclosing 单测；adapter + golden；contrib 迁移 optional follow-up
- **系统**：expand/codegen 行为变更；**Call** provider 可经 adapter 过渡；**Decl** provider **必须**改写（无 adapter）
