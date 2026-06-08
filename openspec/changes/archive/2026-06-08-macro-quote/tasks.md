## 1. 包骨架与模板解析

- [x] 1.1 创建 `macro/quote/` 包骨架（errors、公开 API 签名 stub）
- [x] 1.2 实现 `@kind{ }` 根解析（四种 kind、大括号配平、根后无多余内容）
- [x] 1.3 实现 body 内 `#hole` 扫描（跳过字符串/注释）
- [x] 1.4 为模板解析添加单元测试（合法/非法根、注释内 `#`）

## 2. 合成与 Go 解析

- [x] 2.1 实现四种 kind 的 Go 源文合成（expr / exprs return 包装 / stmts func 包装 / decls package 包装）
- [x] 2.2 使用 `parser.ParseComments` 解析合成源文并按 kind 提取 AST
- [x] 2.3 添加 parse 失败与 kind 提取的单测

## 3. 填洞与 Clone

- [x] 3.1 实现 string 洞（ident 字面）与 `_q_*` 占位 expr 洞
- [x] 3.2 使用 `astutil.Apply`（或等价）替换占位 ident 为绑定 `ast.Expr`
- [x] 3.3 实现 `[]ast.Stmt` / `[]ast.Decl` 单一洞 fast path 展开
- [x] 3.4 实现 AST Clone，避免共享树变异
- [x] 3.5 填洞与列表展开单测（含嵌套 Quote）

## 4. 公开 API 与校验

- [x] 4.1 实现 `Quote`、`Expr`、`Exprs`、`Stmts`、`Decls` 及根 kind 与 API 一致性校验
- [x] 4.2 错误信息包含 kind / hole 名 / 模板上下文；常规路径不 panic
- [x] 4.3 golden 测试：`@expr`、`@exprs`、`@stmts`、`@decls` 各至少一例
- [x] 4.4 注释保留测试：`@stmts{ // hello ... }` printer 含注释

## 5. 文档

- [x] 5.1 更新 `docs/author-guide.md` Quote 子节（根 kind、# 洞、贴回对应、StampStmtPos）
- [x] 5.2 在 author-guide 中注明 `macro/quote` 为可选依赖

## 6. 集成验证（可选）

- [x] 6.1 （可选）在 go-macro-contrib 选一个 Try 分支用 Quote 重写并跑 mactest
- [x] 6.2 根目录 `go test ./...` 全绿
