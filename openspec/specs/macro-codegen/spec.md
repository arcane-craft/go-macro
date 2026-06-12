# macro-codegen Specification

## Purpose
定义宏展开代码生成链路：以 `cmd/macro expand` 作为统一入口，自动生成 link runner 并写回 `*_macro_gen.go`。
## Requirements
### Requirement: go generate 集成

工具链 MUST 支持在宏主文件中通过一行 generate 触发展开，且 MUST NOT 要求用户自建 `cmd/macroexpand` 或手写 `register`。

RECOMMENDED 命令：

```go
//go:generate go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .
```

#### Scenario: generate 零自建入口

- **WHEN** 用户项目仅在宏主文件包含上述 generate，并 import 所用 provider
- **THEN** `go generate` MUST 成功写回 `*_macro_gen.go`

#### Scenario: 按包 generate

- **WHEN** 宏主文件含 `//go:generate ... expand .`
- **THEN** MUST 仅展开该文件所在包（或指令指定 patterns）

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

### Requirement: 主文件 build tag 由用户维护

工具 MUST NOT 修改宏主文件 `foo.go` 上的 build tags。若主文件约束不含 `macro`，expand MUST 失败并提示用户添加。

#### Scenario: 主文件缺少 macro tag

- **WHEN** `readfile.go` 含宏调用但 constraint 不含 `macro`
- **THEN** expand MUST 失败并提示修正

### Requirement: 生成侧 constraint 推导

工具 MUST 从主文件约束推导生成侧：把表达式中的 `macro` 替换为 `!macro`，其余子表达式保持不变。

#### Scenario: 仅 macro 约束

- **WHEN** 主文件含 `//go:build macro`
- **THEN** 生成文件 MUST 含 `//go:build !macro`

### Requirement: 生成代码 line 指令

写入 `*_macro_gen.go` 时，工具 MUST 对 **Call 宏展开语句** 与 **Decl 宏生成的方法体语句** 生成 `//line <macro-main-file>:<line>`，定位回宏主文件。

Decl 方法的 line 行号 MUST 取自对应 **嵌入 marker 字段**位置（或 `DeclContext.MacroPos()` 约定位置）。

#### Scenario: Decl 方法 panic 堆栈

- **WHEN** Decl 生成的方法 panic
- **THEN** 堆栈 MUST 指向宏主文件中 marker 嵌入行

#### Scenario: Call 宏 panic 堆栈

- **WHEN** Call 展开代码 panic
- **THEN** 堆栈 MUST 指向宏主文件中宏调用行

### Requirement: 幂等展开

对同一输入重复执行 `cmd/macro expand`，生成文件 MUST 一致（时间戳除外）。

#### Scenario: 重复执行 expand

- **WHEN** 连续两次相同 expand 且主文件未变
- **THEN** `foo_macro_gen.go` MUST 一致

### Requirement: 仅展开当前主模块

`cmd/macro expand` MUST 仅处理调用方所在主 module 的包，MUST NOT 写 module cache。

#### Scenario: 本模块 expand

- **WHEN** 在某 module 根执行 `go run github.com/arcane-craft/go-macro/cmd/macro@latest expand ./...`
- **THEN** MUST 仅更新该 module 内 `*_macro_gen.go`

### Requirement: expand 子命令自动 link

`cmd/macro expand` 生成的 `.gomacro/expand_runner` MUST 依据 provider 上 `//macro:` 分别 link **CallExpander** 与 **DeclExpander**（按 syntax-id）。同一 provider 包含多个 syntax-id 时 MUST 为每个已发现的 Expander 生成注册调用。

#### Scenario: 多 syntax link

- **WHEN** provider 同时导出 `TryExpand`（Call）与 `DeriveStringerExpand`（Decl）
- **THEN** expand_runner MUST 分别 `RegisterCall` 与 `RegisterDecl`（或等价）

### Requirement: init provider 仅生成 provider 骨架

`init provider` MUST 生成桩与 Expander（含 per-function `//macro:`）及测试模板，MUST NOT 生成 `register/` 或 `tools/macroexpand`。

#### Scenario: 宏作者无 register 义务

- **WHEN** 用户执行 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac`
- **THEN** 输出 MUST 不包含 `register/register.go`

### Requirement: 生成侧类型与方法顺序

`WriteGenFile` 输出 decl 顺序 MUST 保持与宏主文件一致：类型 `TypeSpec` 与其方法相邻，包级函数随后（或按源文件 `Decls` 顺序）。

#### Scenario: 类型在函数前

- **WHEN** 宏主文件先声明 `type Item` 再声明 `func Read()`
- **THEN** gen 文件中 `type Item` MUST 出现在 `func Read` 之前

### Requirement: 生成代码行号

写入 `*_macro_gen.go` 时，引擎 MUST 在生成语句块中使用 `//line` 指向宏主文件。Call 宏行号 MUST 取自 **`site.MacroPos()`**（不再取自 `CallContext.MacroPos()`）。Decl 宏行号 MUST 取自 embed 处 **`site.MacroPos()`**（或 ResolveSite 记录的 embed 位置）。

#### Scenario: Call 宏 line 指令

- **WHEN** expand 替换 `Try(...)` 为多条 stmt
- **THEN** 生成 stmt MUST 带 `//line` 指向原宏主文件 macro 调用行

#### Scenario: 与 StampStmtPos 一致

- **WHEN** Apply 完成后 StampStmtPos
- **THEN** MUST 使用 `site.MacroPos()` 作为 stamp 输入
