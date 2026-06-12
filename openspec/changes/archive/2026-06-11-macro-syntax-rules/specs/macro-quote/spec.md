## REMOVED Requirements

### Requirement: Quote 子包位置与可选性

**Reason**: Quote 能力迁入 `macro.Quote` 与 `Syntax`（macro-template）；`macro/quote` 子包删除。

**Migration**: `import macro/quote` → `macro.Quote`；`map[string]any` → `map[string]Syntax`；移除 `@kind{ }`，用 `To*` 校验形状。

### Requirement: 根 kind 语法

**Reason**: 由 `Syntax.ToExpr` / `ToStmts` / `ToDecls` 决定产出形状，不再使用 `@expr{ }` 等包装。

**Migration**: `quote.Stmts(tpl, args)` → `Quote(tpl, binds).ToStmts()`。

### Requirement: Unquote 填洞语法（# 在独立子包）

**Reason**: 合并至 macro-template；绑定类型改为 Syntax。

**Migration**: 见 macro-template ADDED Requirements。

### Requirement: 四种 kind 产出与校验（typed API Expr/Stmts/Decls）

**Reason**: 移除 `quote.Expr`/`Stmts`/… typed 入口；统一 `Quote` + `To*`。

**Migration**: `quote.Expr("#x", m)` → `Quote("#x", m).ToExpr()`。

### Requirement: 嵌套 Quote

**Reason**: 嵌套通过 Syntax 值传递；由 macro-template 重新定义。

**Migration**: 嵌套 Quote 结果为 Syntax，放入 binds。

### Requirement: 语句与声明列表洞

**Reason**: 由 `#field ...` ellipsis 与 Syntax 列表取代 fast path 描述。

**Migration**: 见 macro-pattern / macro-template。

### Requirement: 公开 API 与错误

**Reason**: API 面变更为 `Quote(template, map[string]Syntax) (Syntax, error)`。

**Migration**: 同上。

### Requirement: 与 go-macro 贴回载荷对应

**Reason**: 贴回由 Match 产出的 `SplicePlan` + `Apply` 完成；Quote 不再对应 CallExpandResult 字段表。

**Migration**: SyntaxRules + SplicePlan；见 author-guide。
