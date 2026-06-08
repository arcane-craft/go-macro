# macro-quote Specification

## Purpose

定义 `macro/quote` 子包的模板化 AST 组装能力：根 kind 语法、`#` 填洞、四种产出形状校验、注释保留，以及与 `CallExpandResult` / `DeclExpandResult` 贴回载荷的对应关系。`macro/quote` 为可选依赖，不改变 expand 引擎契约。
## Requirements
### Requirement: Quote 子包位置与可选性

系统 MUST 在 `github.com/arcane-craft/go-macro/macro/quote` 提供 Quote 能力。`macro` 根包 MUST NOT 将 Quote API 作为必填依赖暴露给所有 provider。未 import `macro/quote` 的 provider MUST 仍可仅用手写 AST 实现 Expander。

#### Scenario: 简单 provider 不依赖 quote

- **WHEN** provider 仅实现最小 Call Expander 且未 import `macro/quote`
- **THEN** MUST 可 `go test` 且 expand 行为与引入 quote 前一致

### Requirement: 根 kind 语法

Quote 支持四种 kind：`expr`、`exprs`、`stmts`、`decls`。kind 与产出形状 MUST 与下表一致。

**Typed API**（`Expr` / `Exprs` / `Stmts` / `Decls`）：函数名已指定 kind，模板 MAY 直接写 **body**（不必再包 `@kind{ }`）。MAY 仍使用显式 `@kind{ }` 包装；若使用包装，其 kind MUST 与所调 API 一致。`@kind` 后 MUST 使用大括号 `{ }`；闭括号 `}` 之后 MUST NOT 含其它非空白内容。

**`Quote(tpl, args)`**：无 typed 入口，模板 MUST 要求最外层为且仅为以下四种之一：

- `@expr{ ... }`
- `@exprs{ ... }`
- `@stmts{ ... }`
- `@decls{ ... }`

MUST NOT 支持多个并列根 kind。

#### Scenario: typed API body-only

- **WHEN** 调用 `quote.Stmts` 且模板为 `x := 1`（无 `@kind{ }`）
- **THEN** MUST 成功并返回非空 `[]ast.Stmt`

#### Scenario: 合法 stmts 根（显式包装）

- **WHEN** 模板为 `@stmts{ x := 1 }` 且调用 `quote.Stmts`
- **THEN** MUST 成功并返回非空 `[]ast.Stmt`

#### Scenario: Quote 缺少根 kind

- **WHEN** 调用 `quote.Quote` 且模板为 `x := 1`（无 `@kind{ }`）
- **THEN** MUST 返回错误

#### Scenario: 使用小括号而非大括号

- **WHEN** 模板为 `@stmts( x := 1 )` 且调用 `quote.Quote` 或带显式包装的 typed API
- **THEN** MUST 返回错误

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

### Requirement: 嵌套 Quote

绑定值 MAY 为嵌套 `quote.Expr` / `Exprs` / `Stmts` / `Decls` / `Quote` 的结果。父模板 `#hole` MUST 接受与子模板产出类型一致的绑定。

#### Scenario: 嵌套 expr

- **WHEN** 父模板 `#inner` 且 `inner` 来自 `quote.Expr("1", nil)`
- **THEN** 父 `Expr` MUST 成功且内层为字面 `1`

### Requirement: 语句与声明列表洞

当 `#name` 绑定 `[]ast.Stmt` 或 `[]ast.Decl` 时，Quote MUST 支持在 stmts/decls kind 模板中将该洞展开为对应列表（首版 MUST 至少支持 **整个 body 仅为单一 `#name` 洞** 的 fast path）。

#### Scenario: 单一 stmts 洞

- **WHEN** 调用 `quote.Stmts` 且模板为 `#block`，`block` 为含两条 assign 的 `[]ast.Stmt`
- **THEN** MUST 返回长度为 2 的语句列表

### Requirement: 注释保留

Quote 实现 MUST 使用 `parser.ParseComments`（或等价）解析合成 Go 源文。填洞与 Clone MUST NOT 主动剥离产出 AST 上已挂载的 `Doc`/`Comment`。模板 body 中、经 parse 进入产出 AST 的注释 SHOULD 在 `go/printer` 输出中仍可呈现。

#### Scenario: 行注释保留

- **WHEN** 调用 `quote.Stmts` 且模板含行注释 `// hello` 与语句 `x := 1`
- **THEN** 合成解析产出的 AST 经 printer 输出的字符串 MUST 包含 `hello`

### Requirement: 公开 API 与错误

`macro/quote` MUST 提供：

- `Quote(tpl string, args map[string]any) ([]ast.Node, error)`（或等价 Syntax 类型）
- `Expr`、`Exprs`、`Stmts`、`Decls` 四个 typed 入口

常规失败路径 MUST 返回 `error`；MUST NOT 以 panic 作为默认错误机制。

#### Scenario: parse 失败返回 error

- **WHEN** 调用 `quote.Stmts` 且模板为 `:= broken`
- **THEN** MUST 返回非 nil error

### Requirement: 与 go-macro 贴回载荷对应

文档 MUST 说明 Quote 产出与 `CallExpandResult` / `DeclExpandResult` 字段的推荐对应：

- `@expr` → `Expr`
- `@exprs` → `Exprs`（`SpliceReplaceReturnResults`）
- `@stmts` → `Stmts`
- `@decls` → Decl Expander 的 `Methods` 等（全量合并由 Expander 负责）

Quote MUST NOT 设置 `SpliceTarget` 或调用 splice 引擎。

#### Scenario: quote 不贴回

- **WHEN** 调用 `quote.Stmts` 成功
- **THEN** MUST NOT 修改任何用户宏主文件 AST
