## Context

go-macro 的 expand 引擎（识别、SpliceTarget 贴回、codegen）已稳定；宏作者的主要成本在 **Expander 内手写 `go/ast`**。`macro-core` 要求 core 保持薄，且 provider **MUST NOT 被强制** 引入厚重 astbuilder。本变更在 **`macro/quote` 可选子包** 中提供模板化 AST 组装，不改动 `internal/expander` 或 `CallExpandResult` 契约。

## Goals / Non-Goals

**Goals:**

- 提供首版 Quote：`@expr{ }` / `@exprs{ }` / `@stmts{ }` / `@decls{ }` 根 kind、`#name` 填洞、嵌套 Quote
- 产出与 go-macro 贴回载荷对齐：`ast.Expr`、`[]ast.Expr`、`[]ast.Stmt`、`[]ast.Decl`
- 模板 body 注释经 `ParseComments` 保留至产出 AST
- 公开 API 返回 `error`（非常规路径不 panic）
- author-guide 说明用法与 `StampStmtPos` / `Target` 衔接

**Non-Goals:**

- Matcher / `$pattern`、编译期 Quote、展开期 `go/types` 二次检查
- 修改 splice 引擎、codegen、`mactest` 契约
- 首版强制迁移 go-macro-contrib（Try/Inline）
- `@decls{ }` 内复杂「手写 decl + 多洞混合」的完整语义（首版可限制为全手写或单一 decl 列表洞）

## Decisions

### D1：独立子包 `macro/quote`

**选择**：Quote 放在 `github.com/arcane-craft/go-macro/macro/quote`，不并入 `macro` 根包。

**理由**：满足 macro-core「core 薄」；简单 provider 可不 import quote。

**备选**：并入 `macro` 根包 — 拒绝，会膨胀 core API。

### D2：根 kind 自描述，typed API 可省略包装

**选择**：四种 kind（`expr` / `exprs` / `stmts` / `decls`）决定产出形状。`Expr`/`Exprs`/`Stmts`/`Decls` 由函数名携带 kind，模板直接写 body；`@kind{ }` 包装可选。`Quote(tpl, args)` 无 typed 入口，仍要求显式 `@kind{ }`。

**理由**：与 author 心智及 `CallExpandResult` 四种载荷一致；typed API 避免重复标注；`Quote` 保留自描述根以便泛用入口。

### D3：占位 ident + `astutil.Apply`，不用 token Pos interval Replace

**选择**：`#name` 在合成 Go 中变为 `_q_name`（或等价前缀）；parse 后按名 Apply 替换为绑定 AST；`[]ast.Stmt` / `[]ast.Decl` 洞通过占位 stmt/decl 再展开为列表。

**理由**：实现量小、行为可测、不依赖自定义 template Pos 算术。

**备选**：自定义 `#` lexer + interval Replace — 拒绝（复杂度高，与 go-macro-design 路线解耦）。

### D4：四种 kind 的固定 parse 包装

| kind | 合成形状 | 提取 |
|------|----------|------|
| `@expr` | `ParseExpr(body)` 或等价单表达式 parse | 一个 `ast.Expr` |
| `@exprs` | `package _; func _() { return BODY }` | `ReturnStmt.Results` |
| `@stmts` | `package _; func _() { BODY }` | `Body.List` |
| `@decls` | `package _; BODY` | `File.Decls` |

**理由**：利用 Go 语法自然表达 expr 列表（return）与 stmt/decl 块。

### D5：绑定参数

**选择**：首版 `map[string]any`，文档约定允许类型：`string`（ident 字面）、`ast.Expr`、`[]ast.Expr`、`ast.Stmt`、`[]ast.Stmt`、`ast.Decl`、`[]ast.Decl`、嵌套 Quote 返回值；后续可加 typed wrapper。

**理由**：与 mactest/Expander 现有风格一致；首版不引入大套 Args builder。

### D6：注释保留

**选择**：合成 body 原样嵌入包装；`parser.ParseComments`；填洞与 Clone MUST NOT 剥离 `Doc`/`Comment`。

**理由**：宏作者常在展开代码中保留说明性注释。

### D7：行号

**选择**：Quote 使用内部 `token.FileSet`；**不**在 quote 内打宏行号。Expander 在返回前调用 `macro.StampStmtPos(ctx.MacroPos(), stmts)`（与 today 一致）。

**理由**：`//line` 语义属于 macro codegen，Quote 保持纯 AST 组装。

### D8：与 macro-core「无 astbuilder 依赖」的关系

**选择**：更新 macro-core 场景——provider **MUST NOT 被强制** import `macro/quote`；import quote **MAY** 用于模板化 Expander。

## Risks / Trade-offs

- **[Risk] `#` 在字符串/注释中被误识别为洞** → body 扫描 MUST 跳过字符串与注释；单测覆盖
- **[Risk] stmt/decl 列表洞展开破坏相邻注释** → 占位 stmt 替换为列表时保留前后 stmt 的 Doc；golden 测试
- **[Risk] `map[string]any` 误绑类型** → 校验绑定类型；错误信息含 hole 名
- **[Risk] Clone 共享 AST 被多次 Quote 篡改** → 注入前 Clone 绑定树
- **[Trade-off] `@decls` 混合模板受限** → 首版文档标明；后续迭代

## Migration Plan

- 纯新增子包与文档；无运行时 **BREAKING**
- 部署：合并后 `go test ./...`；contrib 迁移为可选 PR
- 回滚：不 import `macro/quote` 即可；Expander 可继续手写 AST

## Open Questions

- 是否在 M6 后为 `init provider` 脚手架增加可选 `@stmts` 模板片段（tasks 可列 follow-up）
- typed `quote.Args` wrapper 是否纳入首版（建议 follow-up，首版用 map + 文档）
