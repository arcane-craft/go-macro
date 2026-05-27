## MODIFIED Requirements

### Requirement: 编写宏库保留 normative 契约要点

`编写宏库` MUST 包含且保持与 `macro-core`、`macro-codegen`、`macro-directive` 一致的下列语义（正文 MAY 使用自然语言，不必逐字保留 RFC 2119 英文关键词或设计文档内部代号如「方案 C」「首版」）：

- 每个语法桩与 Expander 的 doc MUST 含 `//macro: <syntax-id>`；同一 syntax-id 下多桩共享一个 Expander
- Provider：`Expand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error)` 签名约定
- 宏主文件 MUST import provider；`cmd/macro expand` 仅对已 import 且已生成 link 的 provider 展开
- MUST NOT 要求宏作者维护 `register/` 包
- 语法桩为包级 `panic` 函数，运行时不可调用
- **宏主文件中，已 link 注册的语法桩 MUST 仅以 `pkg.Stub(...)`（或 dot-import 下 `Stub(...)`）直接调用；MUST NOT 将桩作为函数值传递、赋值、返回或传入 `reflect.ValueOf` / `reflect.TypeOf`；违反时 expand MUST 失败（见 `macro-expander`）**
- `ExpandResult` 的 `Stmts` / `Expr` / `Exprs` 及宏出现位置与返回字段的对应关系（MAY 以表格呈现，标题 MAY 不使用「Site」术语）
- 展开时 `Context` MUST 提供 `EnclosingFunc()`（`*ast.FuncDecl` 或 `*ast.FuncLit`），供 provider 读取外层函数语境
- `init provider` 文档入口为 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>`
- 纯 Expand 单测使用 `macro/mactest` 的示例

#### Scenario: 契约与 macro-directive 对齐

- **WHEN** 读者对照 `macro-directive` 与 `macro-core` 要求
- **THEN** author-guide `编写宏库` MUST 说明 per-function `//macro:`，且 MUST NOT 将 `register/` 列为作者职责

#### Scenario: 值用法约束可查

- **WHEN** 宏库作者或宏使用方阅读 `编写宏库` 或 `宏使用方`
- **THEN** MUST 能找到「桩须直调、不可作函数值」的说明，且 MUST 指向 expand 期报错行为

### Requirement: 宏使用方节保留 codegen 要点

`宏使用方` MUST 说明（语义与 `macro-codegen` 一致，文案 MAY 简化）：

- 宏主文件与 `*_macro_gen.go` 的 build tag 分工
- `go:generate` 调用 `cmd/macro expand`（或等价）生成侧文件
- 发布前在带/不带 `macro` tag 下测试的建议
- **已 link 的语法桩在宏主文件中 MUST 直接调用；将桩作为参数、变量或反射对象会导致 expand 失败**

#### Scenario: 使用方知悉 expand 约束

- **WHEN** 仅使用已有宏库的读者阅读 `宏使用方`
- **THEN** MUST 能理解 `import` 宏库后应写 `try.Try(...)` 一类直调，且 MUST NOT 依赖将 `try.Try` 当作普通函数值
