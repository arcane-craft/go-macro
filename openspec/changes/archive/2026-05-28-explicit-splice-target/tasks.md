## 1. macro 包 API（破坏性）

- [x] 1.1 定义 `SpliceTarget` 常量与 `String()`（错误信息用）
- [x] 1.2 `ExpandResult` 增加必填语义 `Target`；文档注释写明载荷表
- [x] 1.3 实现 `legalSpliceTargets(call, file)` 与 `Context.LegalSpliceTargets()`
- [x] 1.4 实现 `ValidateExpandResult(ctx, result) error`
- [x] 1.5 更新 `macro/expand.go` 注释与 `registry_*_test` 中所有 `ExpandResult` 构造

## 2. expander 引擎 splice

- [x] 2.1 实现 `ResolveSpliceAnchor(file, call)`（供 legal targets 与 apply 共用）
- [x] 2.2 重写 `ApplyExpandResult(file, call, result)`：按 `Target` 分支；实现 `SpliceReplaceAssignRHS`
- [x] 2.3 `ExpandFile`：expander 返回后先 `ValidateExpandResult` 再 apply；移除对 `site` 的 splice 传参
- [x] 2.4 错误信息含合法 `SpliceTarget` 列表
- [x] 2.5 `splice_test.go`：为六种 `Target` + D1 RHS + 非法 Target/载荷添加表驱动测试

## 3. mactest 与脚手架

- [x] 3.1 `mactest.Validate(ctx, result)` 封装
- [x] 3.2 更新 `mactest` 文档注释与示例
- [x] 3.3 `cmd/macro init provider` 与内置示例 Expander 设置默认 `Target`
- [x] 3.4 更新 `cmd/macro/main.go` 等本仓 Expander

## 4. 文档

- [x] 4.1 重写 `docs/author-guide.md`「ExpandResult」节：`Target` 表、`LegalSpliceTargets`、`mactest.Validate`、assign RHS 说明
- [x] 4.2 通读与 `specs/author-guide` delta 一致

## 5. contrib 与联调（可 follow-up PR）

- [x] 5.1 在 `go-macro-contrib` 为 `TryExpand`、`InlineExpand` 等设置 `Target`（assign 路径评估 `SpliceReplaceAssignRHS` vs `SpliceReplaceAssignStmt`）
- [x] 5.2 contrib OpenSpec `syntax-try` / `syntax-inline` 若有 Site×字段表述则改为 Target
- [x] 5.3 本地 `replace` 联调：`examples/` expand + `go test`（contrib `go test` 已通过本地 replace；发版 go-macro 后 examples 随依赖版本更新）

## 6. 收尾

- [x] 6.1 `go test ./...` 与 `cd examples && go test ./...` 通过
- [x] 6.2 `openspec validate explicit-splice-target` 通过
- [x] 6.3 归档前确认 BREAKING 已在 README 或 CHANGELOG 注明（若项目有）
