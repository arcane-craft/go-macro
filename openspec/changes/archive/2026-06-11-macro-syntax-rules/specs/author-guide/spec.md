## ADDED Requirements

### Requirement: syntax-rules 作者指引

作者指南 MUST 说明 **`SyntaxRules` / `SyntaxCase`** 为默认 Expander 形态；pattern 使用 `$`、Quote 使用 `#`；裸标识符为 literal（无 `Literals` 字段）。MUST 说明 **MatchedSpan**：pattern match 哪部分，Apply 就只替换哪部分（Call/Decl 同一规则）；out 可含比 match 更多的 stmt/decl（Try 的 `if err`、Derive 的生成 methods）。

MUST 说明 **pattern 首版子集**（design D17、macro-pattern）：

- 顶层：`CallPattern` / `StmtPattern` / `DeclPattern`
- stub **invoked name** 匹配（`Try($inner)` 匹配 `tr.Try(...)`）
- assign：`$lhs ... :=`、`$lhs ... =`、`var $lhs ... =`（首版不支持包级 `var`）；return：`return $vals ... , Try($inner)`；ExprStmt：`Try($inner);` 为 stmt 级语法糖
- Decl anchor：embed `*ast.Field`；`MacroPos` 为 embed 位置（design D18）
- 裸 `Expander`：MUST 在返回前 `site.Match(pattern)`（design D19）
- Decl struct：**顺序无关**；`Derive[$iface]` 与 `$field ...` 书写顺序不影响 match
- ellipsis 经 `Bindings.Elems`；SyntaxCase 宽 Stmt clause 应排在窄 Call clause 前

#### Scenario: Inline 示例

- **WHEN** 读者查看 Call 宏最小示例
- **THEN** MUST 展示 `SyntaxRules` 单 clause，而非手写 `go/ast`

### Requirement: 统一 Expander 与 Context

作者指南 MUST 说明 **统一** `Expander(ctx Context, site Syntax) (Syntax, error)`；**不**区分 Call/Decl 宏章节为 normative 两套签名。`Context` MUST 文档化为三字段；`MacroPos` 自 **`site.MacroPos()`** 获取。

#### Scenario: Try EnclosingResults 文档

- **WHEN** 读者查看 Try 实现指引
- **THEN** MUST 说明 error 分支 `return` 使用 `EnclosingResults` + `ZeroSyntax`，不得仅依据 assign lhs

#### Scenario: Derive MatchedSpan 与生成 methods

- **WHEN** 读者查看 Derive 实现指引
- **THEN** MUST 说明 pattern 划定 type 边界、out 含新 TypeSpec 与 **新生成** methods 即可，**不必**复制 Target 既有未 match methods；并说明与 Try「替换载荷多 stmt」的对称性

#### Scenario: Derive SyntaxCase 完整示例

- **WHEN** 读者查看 Decl 宏迁移或 Derive 实现指引
- **THEN** MUST 提供与 design.md「Derive SyntaxCase 示例」等价的完整 walkthrough：使用方源码、pattern、`deriveTransform` 读取 `$item`/`$iface`/`Elems("field")`/`site.Underlying()` embed tag、Quote 产出 `[TypeSpec', FuncDecl...]`、引擎 Plan 两步贴回、与旧 `DeclExpander` 对比表、常见错误（embed 未移除、TypeSpec 改名、method 同名冲突）

#### Scenario: Decl embed 元数据与 Underlying

- **WHEN** 读者查看 Derive / Decl 宏实现指引
- **THEN** MUST 说明 **不**恢复 `DeclContext` / `TargetMethods()`；MUST 提供旧 API → 新路径对照（`MarkerTypeArgs` → `binds.Get("iface")` + `Types().TypeOf`；`MacroTag` → `*ast.Field.Tag` + `ParseMacroTag`；`$field ...` → `binds.Elems("field")` + `Underlying()` 为 `*ast.Field`）；MUST 说明 Decl pattern 不依赖 struct 字段顺序

### Requirement: BREAKING 迁移

作者指南 MUST 含迁移表：

| 旧 API | 新 API | 过渡 |
|--------|--------|------|
| `CallExpander` | `Expander` | MAY `TargetToPlan` adapter（短期） |
| `CallExpandResult` | `Syntax` + `SplicePlan` | 同上 |
| `DeclExpander` / `DeclExpandResult` | `SyntaxCase` + `Syntax` | **无 adapter**，必须改写 |
| `DeclContext.Site()` / `MacroTag` / `MarkerTypeArgs` | `Bindings` + `Underlying()` + `ParseMacroTag` / `Types().TypeOf` | 无专用 accessor |
| `DeclContext.TargetMethods()` | 不恢复；仅生成新 methods | 读已有 methods 仅 escape |
| `macro/quote` | `macro.Quote` | 直接迁移 |

#### Scenario: 迁移入口

- **WHEN** 现有 provider 作者阅读指南
- **THEN** MUST 区分 Call adapter 期限与 Decl 强制改写说明

## MODIFIED Requirements

### Requirement: Quote 编写指引

原 `macro/quote` 章节 MUST 替换为 **`macro.Quote` + SyntaxRules**；MUST NOT 再要求 `@kind{ }` 或 import `macro/quote`。

#### Scenario: 无 quote 子包

- **WHEN** 读者按指南实现模板化宏
- **THEN** MUST 仅 import `macro` 根包（及必要 internal 无）

## REMOVED Requirements

### Requirement: CallExpandResult Target 对照表（作为作者必填）

**Reason**: 贴回计划在 Match 时确定；Rule 可选 `Plan` override。

**Migration**: 文档改为 SplicePlan / ValidateSplice 规则摘要；见 macro-splice-apply。
