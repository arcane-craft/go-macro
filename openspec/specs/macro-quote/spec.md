# macro-quote Specification

## Purpose
定义 `macro.Quote` 模板化 AST 组装能力：`#` 填洞、`Syntax` 绑定、产出形状由 `To*` 校验、注释保留。`macro/quote` 子包已移除，能力迁入 `macro` 根包。
## Requirements
### Requirement: Unquote 填洞语法

模板体内 MUST 使用 `#` 加 Go 标识符表示填洞（`#name`）。`#` 扫描 MUST 忽略字符串字面量、字符字面量、行注释与块注释内的 `#`。

绑定表 MUST 为每个出现的 hole 名提供值；缺失 MUST 报错。

#### Scenario: 字符串绑定为 ident

- **WHEN** 调用 `quote.Expr` 且模板为 `#x`，args 中 `"x": "file"`（string）
- **THEN** 产出 MUST 语义等价于 ident `file` 的表达式

#### Scenario: AST 表达式洞

- **WHEN** 调用 `quote.Expr` 且模板为 `#inner`，args 中 `inner` 为 `ast.Expr`
- **THEN** 产出 Expr MUST 与绑定节点结构等价（允许 Clone），且 MUST NOT 要求经 printer 再 parse

#### Scenario: 注释内 hash 非洞

- **WHEN** 调用 `quote.Stmts` 且模板含行注释 `// #not_a_hole` 与语句 `x := 1`
- **THEN** MUST NOT 将 `not_a_hole` 视为 hole 名

### Requirement: 四种 kind 产出与校验

四种 kind 的产出形状与校验规则 MUST 符合下表：

| 根 kind | API | 产出 | 校验 |
|---------|-----|------|------|
| `@expr` | `Expr` | 单个 `ast.Expr` | 恰好一个 Expr |
| `@exprs` | `Exprs` | `[]ast.Expr` | 长度 ≥ 1 |
| `@stmts` | `Stmts` | `[]ast.Stmt` | 长度 ≥ 1 |
| `@decls` | `Decls` | `[]ast.Decl` | 长度 ≥ 1 |

`Quote(tpl, args)` MUST 解析根 kind 并返回 `[]ast.Node`（或等价 Syntax 类型）。`Expr`/`Exprs`/`Stmts`/`Decls` MUST 在根 kind 与 API 不一致时返回错误。

#### Scenario: exprs 产出 return 列表

- **WHEN** 调用 `quote.Exprs` 且模板为 `#v, nil`，`#v` 绑定为 `ast.Expr`
- **THEN** MUST 返回长度为 2 的 `[]ast.Expr`

#### Scenario: API 与根 kind 不匹配

- **WHEN** 调用 `quote.Stmts` 且模板为 `@expr{ 1 }`（显式包装 kind 与 API 不一致）
- **THEN** MUST 返回错误

### Requirement: 注释保留

Quote 实现 MUST 使用 `parser.ParseComments`（或等价）解析合成 Go 源文。填洞与 Clone MUST NOT 主动剥离产出 AST 上已挂载的 `Doc`/`Comment`。模板 body 中、经 parse 进入产出 AST 的注释 SHOULD 在 `go/printer` 输出中仍可呈现。

#### Scenario: 行注释保留

- **WHEN** 调用 `quote.Stmts` 且模板含行注释 `// hello` 与语句 `x := 1`
- **THEN** 合成解析产出的 AST 经 printer 输出的字符串 MUST 包含 `hello`

