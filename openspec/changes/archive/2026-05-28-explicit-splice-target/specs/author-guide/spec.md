## MODIFIED Requirements

### Requirement: 编写宏库保留 normative 契约要点

`编写宏库` MUST 包含且保持与 `macro-core`、`macro-codegen`、`macro-directive` 一致的下列语义（正文 MAY 使用自然语言，不必逐字保留 RFC 2119 英文关键词或设计文档内部代号如「方案 C」「首版」）：

- 每个语法桩与 Expander 的 doc MUST 含 `//macro: <syntax-id>`；同一 syntax-id 下多桩共享一个 Expander
- Provider：`Expand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error)` 签名约定
- 宏主文件 MUST import provider；`cmd/macro expand` 仅对已 import 且已生成 link 的 provider 展开
- MUST NOT 要求宏作者维护 `register/` 包
- 语法桩为包级 `panic` 函数，运行时不可调用
- **宏主文件中，已 link 注册的语法桩 MUST 仅以 `pkg.Stub(...)`（或 dot-import 下 `Stub(...)`）直接调用；MUST NOT 将桩作为函数值传递、赋值、返回或传入 `reflect.ValueOf` / `reflect.TypeOf`；违反时 expand MUST 失败（见 `macro-expander`）**
- **`ExpandResult` MUST 设置显式 `Target`（`SpliceTarget`）**；MUST 说明各 `Target` 替换的 AST 范围及对应载荷字段（`Stmts` / `Expr` / `Exprs`）；MUST 说明 `ctx.LegalSpliceTargets()` 与 `mactest.Validate`（或等价）用于单测校验；MAY 保留「调用处语境」表（assign/return/语句/表达式）作阅读辅助，但 MUST NOT 暗示仅凭填哪个字段即可贴回
- MUST 说明 `SpliceReplaceAssignRHS`：保留赋值左侧，仅替换含宏调用的右侧表达式
- 展开时 `Context` MUST 提供 `EnclosingFunc()`（`*ast.FuncDecl` 或 `*ast.FuncLit`），供 provider 读取外层函数语境
- `init provider` 文档入口为 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>`
- 纯 Expand 单测使用 `macro/mactest` 的示例（含 `Validate`）

#### Scenario: 契约与 macro-directive 对齐

- **WHEN** 读者对照 `macro-directive` 与 `macro-core` 要求
- **THEN** author-guide `编写宏库` MUST 说明 per-function `//macro:`，且 MUST NOT 将 `register/` 列为作者职责

#### Scenario: 值用法约束可查

- **WHEN** 宏库作者或宏使用方阅读 `编写宏库` 或 `宏使用方`
- **THEN** MUST 能找到「桩须直调、不可作函数值」的说明，且 MUST 指向 expand 期报错行为

#### Scenario: 显式 Target 可查

- **WHEN** 宏库作者阅读 `编写宏库` 中 ExpandResult 说明
- **THEN** MUST 能找到 `Target` 与「替换整条语句 / 仅 RHS / 仅 CallExpr」的对照，且 MUST NOT 仅以「宏写在哪里 → 填 Stmts 或 Expr」隐式规则作为唯一说明
