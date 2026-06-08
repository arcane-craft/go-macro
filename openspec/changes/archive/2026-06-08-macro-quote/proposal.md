## Why

宏作者 today 在 Expander 中大量手写 `go/ast` 结构体（Try、Inline 等官方宏均为数百行 AST 拼装），可读性差、易错、难测。需要一套 **模板化组装 AST** 的能力：作者写接近最终 Go 的模板，用 `#` 填洞，由框架 parse 并产出 `ast.Expr` / `[]ast.Stmt` / `[]ast.Decl` 等，再交给现有 splice 与 codegen。这与 `macro-core`「core 保持薄、provider 可选增强」的方向一致，且不影响 expand 引擎契约。

## What Changes

- 新增 **`macro/quote`** 子包，提供 Quote 首版 API：`Quote`、`Expr`、`Exprs`、`Stmts`、`Decls`
- 模板语法：根 **`@expr{ }` / `@exprs{ }` / `@stmts{ }` / `@decls{ }`** 之一；体内 **`#name`** 填洞；支持嵌套 Quote 与 AST 注入
- 实现路径：解析模板根 → 按 kind 合成 Go 源文 → `go/parser`（含 `ParseComments`）→ 占位符替换 → 校验产出形状
- 模板 body 中的 **注释 SHOULD 保留** 至产出 AST（经 printer 进入生成代码）
- **`docs/author-guide.md`** 增加 Quote 编写指引（可选子包、与 `CallExpandResult` / `DeclExpandResult` 衔接）
- **不**改变 `internal/expander` splice、`SpliceTarget`、link 或 codegen 语义
- **不**引入 Matcher / `$pattern`、编译期 Quote、展开期二次 typecheck
- 首版 **不强制** contrib（Try/Inline）迁移；可在 tasks 中作为可选 follow-up

## Capabilities

### New Capabilities

- `macro-quote`: Quote 模板语法、`macro/quote` 公开 API、绑定参数语义、四种根 kind 校验、注释保留、与 go-macro 贴回载荷的对应关系

### Modified Capabilities

- `macro-core`: 更新「轻薄 AST 辅助」要求——core 仍不强制 astbuilder；**MAY** 文档指向可选的 `macro/quote`；简单宏仍可不依赖 quote
- `author-guide`: 增加宏作者使用 Quote 的章节与示例

## Impact

- **代码**：新包 `macro/quote/`（template 解析、synth、parse、bind、validate）；`docs/author-guide.md`；`openspec/specs/macro-quote/spec.md`
- **API**：新公开 API，无 **BREAKING** 变更
- **依赖**：复用已有 `golang.org/x/tools/go/ast/astutil`（Apply）；不新增第三方模块
- **测试**：`macro/quote` 单元测试 + golden；可选 contrib spike
- **系统**：expand 引擎、codegen、`mactest` 行为不变；Expander 作者 opt-in 使用 Quote
