## MODIFIED Requirements

### Requirement: 主文件与生成侧写回

带宏的源文件 MUST 保持主文件名 `foo.go`；工具 MUST 生成/更新 `foo_macro_gen.go`。

生成侧 MUST 包含展开后的：

- 宏主文件中经 Decl 展开后的 **类型定义**（`TypeSpec`，含 struct 字段与 tag）；
- 上述类型的 **方法**（`*ast.FuncDecl`，receiver 为该类型）；
- 经 Call 宏展开后的 **包级函数**（`*ast.FuncDecl`，无 receiver 或原有包级函数体）。

用户 MUST NOT 因宏而重命名主文件。

#### Scenario: gen 含类型定义

- **WHEN** `foo.go` 含 `type Item struct { DeriveStringer; A int }` 且 Decl 展开成功
- **THEN** `foo_macro_gen.go` MUST 含展开后的 `type Item struct { ... }`

#### Scenario: 默认构建仅使用生成侧

- **WHEN** 宏展开成功且用户执行 `go build`（无 `-tags macro`）
- **THEN** 编译 MUST 使用 `*_macro_gen.go` 中的类型与方法及函数

### Requirement: 生成代码 line 指令

写入 `*_macro_gen.go` 时，工具 MUST 对 **Call 宏展开语句** 与 **Decl 宏生成的方法体语句** 生成 `//line <macro-main-file>:<line>`，定位回宏主文件。

Decl 方法的 line 行号 MUST 取自对应 **嵌入 marker 字段**位置（或 `DeclContext.MacroPos()` 约定位置）。

#### Scenario: Decl 方法 panic 堆栈

- **WHEN** Decl 生成的方法 panic
- **THEN** 堆栈 MUST 指向宏主文件中 marker 嵌入行

#### Scenario: Call 宏 panic 堆栈

- **WHEN** Call 展开代码 panic
- **THEN** 堆栈 MUST 指向宏主文件中宏调用行

### Requirement: expand 子命令自动 link

`cmd/macro expand` 生成的 `.gomacro/expand_runner` MUST 依据 provider 上 `//macro:` 分别 link **CallExpander** 与 **DeclExpander**（按 syntax-id）。同一 provider 包含多个 syntax-id 时 MUST 为每个已发现的 Expander 生成注册调用。

#### Scenario: 多 syntax link

- **WHEN** provider 同时导出 `TryExpand`（Call）与 `DeriveStringerExpand`（Decl）
- **THEN** expand_runner MUST 分别 `RegisterCall` 与 `RegisterDecl`（或等价）

## ADDED Requirements

### Requirement: 生成侧类型与方法顺序

`WriteGenFile` 输出 decl 顺序 MUST 保持与宏主文件一致：类型 `TypeSpec` 与其方法相邻，包级函数随后（或按源文件 `Decls` 顺序）。

#### Scenario: 类型在函数前

- **WHEN** 宏主文件先声明 `type Item` 再声明 `func Read()`
- **THEN** gen 文件中 `type Item` MUST 出现在 `func Read` 之前
