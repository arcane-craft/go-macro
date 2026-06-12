## 1. Syntax 与 Context 基础

- [x] 1.1 定义 `Syntax`、`Bindings`（`Get` + `Elems`）、`Context`（三字段）与 `Expander` 签名
- [x] 1.2 实现 `Syntax` 包装 AST、`To*` 校验、`Underlying()`、`MacroPos()`；`siteImpl` 含内部 **meta 槽**（初始为空）
- [x] 1.3 实现 `internal/expander/resolve.go`：`ResolveSite(file, anchor) Syntax`（Call anchor `*ast.CallExpr`；Decl anchor embed `*ast.Field`，见 design D18；每轮新 site、空 meta 槽）
- [x] 1.4 实现 `internal/expander/meta.go`：`MatchMetaFromSite` / `SetMatchMeta`（引擎专用，不 export 给 provider）

## 2. Pattern Match（macro-pattern）

- [x] 2.1 实现 pattern 解析与扫描（design D17：`CallPattern` / `StmtPattern` / `DeclPattern`；assign `:=` / `=` / `var ... =`；`$`、`$_`、`$lhs ...` / `$field ...` / `$vals ...`；裸 ident = literal；**invoked name** callee；首版 **无**逐字段 `$name $type`）
- [x] 2.2 实现 anchored match（Call unify anchor；Decl 以 embed + TypeSpec 根 + 无序约束集）、`MatchedSpan`、`MatchRoot`、Plan 推导（Stmt 固定 BlockStmts；Call 按父槽位；Decl 两步）；`Bindings.Elems`；`site.Match` 成功时写入 meta 槽
- [x] 2.3 添加 match 单测：assign `$lhs ... :=` / `$lhs ... =` / `var $lhs ... =`、return `$vals ... ,`、ExprStmt `Try($inner)` vs `Try($inner);`、invoked name（`tr.Try`）、Decl embed 前后顺序等价、Decl ResolveSite（embed anchor + MacroPos）、同句两 macro、**$field/$iface Underlying 与 field.Tag**

## 3. Quote（macro-template，合并 macro/quote）

- [x] 3.1 实现 `Quote(template, map[string]Syntax) Syntax` 与 `#` / `#field ...` 注入
- [x] 3.2 移植/改写原 `macro/quote` 测试至新 API（无 `@kind`）
- [x] 3.3 标记 `macro/quote` deprecated；文档指向 `macro.Quote`

## 4. EnclosingSignature（macro-enclosing-signature）

- [x] 4.1 实现 `internal/expander/scope.go`：Pos → `*types.Func`（不 export FuncDecl）
- [x] 4.2 实现 `EnclosingSignature`、`EnclosingResults`、`ZeroSyntax`
- [x] 4.3 添加与 today EnclosingFunc 行为对照测试

## 5. SyntaxRules / SyntaxCase（macro-rules）

- [x] 5.1 实现 `Clause`、`SyntaxRules`、`SyntaxCase`（pattern 解析 fatal；match/fender 顺序；失败 clause 清空 meta 槽；`Clause.Plan` override）
- [x] 5.2 确认无 `Literals` 字段、无 Try 专用 built-in；裸 `Expander` 路径须 `site.Match`（design D19）
- [x] 5.3 添加 Inline SyntaxRules、Try SyntaxCase 与 **Derive SyntaxCase** 集成测试（Decl 无 adapter，须 native 新 API）

## 6. SplicePlan 校验与 Apply（macro-splice-apply）

- [x] 6.1 实现 `SpliceStep`（`ReplaceInContainer`、`InsertAfterInFileDecls`）、`ValidateSplice(out, meta)`、`Apply(file, meta, out)`
- [x] 6.2 实现 **`TargetToPlan` / `CallExpandResultToSyntax`（仅 Call adapter）**；adapter 经 `SetMatchMeta` 写入 site 槽；含 `ReplaceAll`（ReturnResults）与 `ExprSlot`（CallExpr）
- [x] 6.3 添加 ValidateSplice/Apply 单测：Try stmt/call（含 `=` 与 `var` 形）、Derive 保留未 match methods、method 同名冲突、Apply 不重复 Inspect、Validate 失败不 Apply

## 7. 引擎集成（macro-expander）

- [x] 7.1 改写 `internal/expander/expand.go` 流水线：ResolveSite → Expander → `MatchMetaFromSite` → ValidateSplice → Apply → StampStmtPos（meta 槽空则失败）
- [x] 7.2 统一 Call/Decl 识别 → 同一 `Expander` 分发（Decl 以 embed `*ast.Field` 为 ResolveSite anchor）
- [x] 7.3 更新 `expandtool.Register` 为统一 Expander（MAY 保留 `RegisterCall` adapter 别名；**无 `RegisterDecl` adapter**）

## 8. 迁移（macro-core）

- [x] 8.1 实现 **`CallExpander` → `Expander` adapter**（`TargetToPlan`）；**不**实现 DeclExpander adapter
- [x] 8.2 改写 **Derive 等 Decl provider** 为 `SyntaxCase`（blocking：`internal/expander/decl_test.go`、`macro/mactest/decl.go`）；更新 `mactest`（Call adapter 路径 + Decl 新 API）
- [x] 8.3 更新 `init provider` 脚手架为 SyntaxRules / SyntaxCase 模板

## 9. 文档与 OpenSpec archive 准备

- [x] 9.1 更新 `docs/author-guide.md`（syntax-rules、**pattern 首版子集 D17–D19**、Context 三字段、**Call adapter vs Decl 强制迁移**、Decl embed 元数据 **Bindings + Underlying + Elems** 对照表，见 design D16/D17/D18/D19）
- [x] 9.2 更新 README Breaking 说明（Call adapter 期限；Decl 无过渡）
- [x] 9.3 全量 `go test ./...`；可选 contrib spike 文档

## 10. 清理

- [x] 10.1 删除 `CallExpandResult`/`CallContext`、**`DeclExpandResult`/`DeclContext`/`DeclExpander`**、Call adapter、`ApplyDeclExpandResult`
- [x] 10.2 删除 `macro/quote` 子包
- [x] 10.3 archive change 并合并 delta specs 至 `openspec/specs/`
