# Proposal: explicit-splice-target

## Why

当前 `ExpandResult` 通过「调用处 `CallSiteKind` + 填哪个字段（`Stmts`/`Expr`/`Exprs`）」隐式决定 AST 贴回方式，宏作者需背对照表；`SiteReturn` 还允许 `Stmts` 与 `Exprs` 二选一，易错且 `mactest` 难以在 splice 前校验。产品需要 **显式贴回目标**、**赋值仅换 RHS（D1）**，并在不跑全链路 expand 时尽早发现 Target 与锚点/载荷不匹配。

## What Changes

- **BREAKING**：`macro.ExpandResult` 增加必填 `Target`（`SpliceTarget`），贴回由 `Target` + 载荷字段决定；移除「仅凭 `ctx.Site()` 推断用哪个字段」的引擎行为。
- 新增 `SpliceReplaceAssignRHS`：仅替换 `AssignStmt.Rhs` 中含宏调用的那一项，保留 `Lhs`。
- `macro.Context` 增加 `LegalSpliceTargets()`（或等价 API），枚举当前调用处在 AST 上合法的贴回目标，供 Expander 与 `mactest` 校验。
- `internal/expander.ApplyExpandResult` 重写为按 `Target` 贴回；`Target` 与真实锚点不一致时 MUST 报错并尽量列出合法 Target。
- 更新 `docs/author-guide.md` 中 `ExpandResult` 说明（显式 Target 表，替代隐式 Site×字段表）。
- **BREAKING（contrib）**：`go-macro-contrib` 中 `TryExpand`、`InlineExpand` 等 MUST 迁移为新 `ExpandResult`（本仓 tasks 含联调说明；contrib 规范在 contrib 仓维护）。

## Capabilities

### New Capabilities

（无独立新 capability；行为归入现有 macro 能力。）

### Modified Capabilities

- `macro-core`：`ExpandResult` 形状、`SpliceTarget` 类型、`Context` 合法 Target 查询
- `macro-expander`：splice 规则、校验与错误信息；新增 `ReplaceAssignRHS`
- `author-guide`：编写宏库契约与 `ExpandResult` 文档表

## Impact

| 区域 | 影响 |
|------|------|
| `macro/expand.go`、`macro/context.go` | **BREAKING** API |
| `internal/expander/splice.go`、`site.go` | 贴回与锚点解析 |
| `macro/mactest` | 可选 `ValidateResult`；示例更新 |
| `cmd/macro` 内置示例 Expander | 设置 `Target` |
| 全仓 `ExpandResult{...}` 测试与桩 | 全部加 `Target` |
| `go-macro-contrib` | **BREAKING** 同步迁移（独立仓库） |
| `openspec/specs/*` | 归档时合并 delta |

## Boundaries（已定）

| 议题 | 决策 |
|------|------|
| 体验优先级 | 写 Expander 可读性与 `mactest` 早失败 **同等重要**（C） |
| 硬需求贴回 | **D1**：仅换 assign RHS |
| API 演进 | **E1**：破坏性改动，不保留旧隐式字段推断 |
| `ReplaceAssignRHS` 载荷 | 仅 `Expr` 替换 `Rhs` 中对应槽位；多值 `a, b := Macro()` 用单个返回值表达式替换原 `CallExpr` |
| 非目标 | 条件位展开为语句块、任意 AST Patch、改 LHS/函数签名 |

## Success criteria

- `openspec validate explicit-splice-target` 通过
- `go test ./...`（本仓）通过；examples 若受影响一并更新
- author-guide 与 spec delta 一致
- contrib 迁移在 tasks 中跟踪（可在本 PR 后单独 PR）
