## ADDED Requirements

### Requirement: Quote 模板与 Syntax 绑定

系统 MUST 提供 `Quote(template string, binds map[string]Syntax) (Syntax, error)`。模板 MUST 使用 `#name` 表示填洞；MUST 支持 `#field ...` ellipsis 注入列表。`#` 扫描 MUST 忽略字符串、字符字面量与注释内的 `#`。

#### Scenario: 单洞 expr

- **WHEN** template 为 `#v` 且 binds 中 `v` 为 ident Syntax
- **THEN** `Quote` 成功且 `ToExpr()` 成功

#### Scenario: 注释内 hash 非洞

- **WHEN** template 含 `// #not_a_hole` 与 `#x`
- **THEN** MUST NOT 将 `not_a_hole` 视为洞名

### Requirement: 产出形状由 To 校验

Quote MUST NOT 要求 `@expr{ }` / `@stmts{ }` 根包装。调用方 MUST 通过 `ToExpr` / `ToStmts` / `ToDecls` 等校验产出形状；形状不符 MUST 返回 error。

#### Scenario: 语句 template

- **WHEN** template 为 `#x := 1` 且 binds 合法
- **THEN** `ToStmts()` MUST 成功

### Requirement: 注释保留

Quote 实现 MUST 使用 `parser.ParseComments`（或等价）。填洞 MUST NOT 剥离已挂载 `Doc`/`Comment`。

#### Scenario: 行注释保留

- **WHEN** template 含 `// hello` 与 `#x := 1`
- **THEN** printer 输出 MUST 仍含 `hello`

### Requirement: 合并 macro-quote 子包

`macro/quote` 独立子包 MUST deprecated；能力 MUST 迁入 `macro.Quote` 与 `Syntax`。绑定类型 MUST 为 `map[string]Syntax`（非 `map[string]any`）。迁移完成后 MUST 删除 `macro/quote` 子包。

#### Scenario: 旧 quote 导入 deprecated

- **WHEN** provider import `macro/quote`
- **THEN** MUST 可编译（deprecated 期）或文档指向 `macro.Quote`
