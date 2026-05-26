# macro-codegen Specification

## Purpose
TBD - created by archiving change go-macro-extension. Update Purpose after archive.
## Requirements
### Requirement: go generate 集成

工具链 MUST 支持在宏主文件中通过**一行** generate 触发 expand，**无需**用户项目内 `tools/macroexpand`。触发的 expand 进程 MUST 通过 blank import 所需 `register` 包并调用 `macro/expandtool.Main()`（或与 `Run` 等价的接线模式）。

文档与快速上手 **RECOMMENDED** 使用：

```go
//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .
```

（整模块可用 `./...`；按包展开用 `.`。）

上述 RECOMMENDED 命令在本仓库中编译并运行 examples module 下的参考 `cmd/macroexpand`（其内部 blank import `github.com/arcane-craft/go-macro-contrib/register`）。宏使用方 MAY 将 generate 改为 `go run <本项目>/cmd/macroexpand` 等等价入口。

#### Scenario: generate 零项目 expand 文件

- **WHEN** 用户仅使用官方宏库（`go-macro-contrib`），宏主文件含 RECOMMENDED 上述 generate（或等价自建入口的 generate），且项目内无 `tools/macroexpand`
- **THEN** `go generate` MUST 成功写回 `*_macro_gen.go`

#### Scenario: 按包 generate

- **WHEN** 宏主文件含 `//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .`（或指向本项目等价入口的 generate）
- **THEN** MUST 仅展开该 generate 所在包（或指令指定的 patterns）

### Requirement: 方案 C 主文件 + 生成侧写回

带宏的源文件 MUST 保持主文件名 `foo.go`（用户编辑，含 `Try`）。工具 MUST 生成/更新 `foo_macro_gen.go` 作为展开产物。用户 MUST NOT 为使用宏而重命名主文件（例如改为 `foo.macro.go`）。

#### Scenario: 主文件命名不变

- **WHEN** 开发者在 `readfile.go` 中使用 `Try` 并执行 expand
- **THEN** 工具 MUST 更新 `readfile_macro_gen.go` 且 MUST NOT 要求将主文件改名为其它后缀

#### Scenario: 默认构建仅使用生成侧

- **WHEN** 宏展开成功且用户执行 `go build`（无 `-tags macro`）
- **THEN** 编译 MUST 使用 `readfile_macro_gen.go` 且 MUST NOT 编译未带 `-tags macro` 的宏源主文件

#### Scenario: 用户不编辑生成侧

- **WHEN** 开发者修改 `readfile.go` 并执行 expand
- **THEN** 工具 MUST 更新 `readfile_macro_gen.go` 且开发者无需手改生成文件

### Requirement: 主文件 build tag 由用户维护

工具 MUST NOT 修改宏源主文件 `foo.go` 上的 `//go:build` / `// +build` 行。用户 MUST 自行将 `macro` 与已有平台/自定义 constraint 合并。若主文件 constraint 不含标识符 `macro`，expand MUST 失败并提示用户添加。

#### Scenario: 用户自行合并平台 tag

- **WHEN** `readfile.go` 含 `//go:build macro && linux` 且宏展开成功
- **THEN** 工具 MUST 保持 `readfile.go` 的 constraint 不变，且生成的 `readfile_macro_gen.go` MUST 含 `//go:build !macro && linux`（或等价推导结果）

#### Scenario: 主文件缺少 macro tag

- **WHEN** `readfile.go` 含宏调用但 constraint 为 `linux` 或为空且不含 `macro`
- **THEN** expand MUST 失败并提示用户在主文件自行加入 `macro` 与现有 constraint 的合并方式

#### Scenario: 禁止 ignore 作为主文件唯一排除方式

- **WHEN** 主文件仅含 `//go:build ignore` 且无 `macro`
- **THEN** expand MUST 失败并说明应改用 `macro` 约束或参考方案 A

### Requirement: 生成侧 constraint 推导

工具 MUST 从主文件 constraint 推导生成侧：将表达式中的 `macro` 替换为 `!macro`，其余子表达式保持不变，并写入 `foo_macro_gen.go` 头部。

#### Scenario: 仅 macro 约束

- **WHEN** 主文件含 `//go:build macro`
- **THEN** 生成文件 MUST 含 `//go:build !macro`

#### Scenario: 复杂表达式互补

- **WHEN** 主文件含 `//go:build macro && (linux || darwin)`
- **THEN** 生成文件 MUST 含 `//go:build !macro && (linux || darwin)`（或规范化后的等价形式）

### Requirement: 生成代码 line 指令

写入 `*_macro_gen.go` 时，工具 MUST 在展开插入的语句前生成 `//line <macro-main-file>:<line>`，使运行态堆栈与调试映射至宏主文件行号，而非生成文件行号。

#### Scenario: panic 堆栈指向宏源码

- **WHEN** 展开后 gen 文件内代码 panic
- **THEN** 堆栈 MUST 显示宏主文件（如 `readfile.go`）中的行号，当且仅当 `//line` 已正确生成

### Requirement: 幂等展开

对同一输入重复执行同一 expand 入口（RECOMMENDED 为 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand`，或调用方自建等价入口），生成文件 MUST 一致（时间戳除外）。

#### Scenario: 重复执行 expand

- **WHEN** 连续两次相同 expand 且主文件未变
- **THEN** `foo_macro_gen.go` MUST 一致

### Requirement: 仅展开当前主模块

由 `expandtool` 驱动的 expand 入口进程 MUST 仅处理**调用方所在**主 module 内包，MUST NOT 写 module cache。

#### Scenario: 本模块 expand

- **WHEN** 于某 module 根通过 expand 入口执行 expand（RECOMMENDED：`go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand ./...`）
- **THEN** MUST 仅更新该 module 内 `*_macro_gen.go`

### Requirement: 对外库须提交生成物

文档与示例 MUST 规定：若模块会被他人在 `go.mod` 中 `require` 引用，维护者 **MUST** 在版本库中提交与宏主文件配对的 `*_macro_gen.go`（发布前已执行 expand）。

#### Scenario: 下游无需 macro 工具链

- **WHEN** 库维护者已提交 `foo_macro_gen.go` 并发布，下游 `go get` 该版本
- **THEN** 下游在未安装 `go-macro` 时 MUST 能仅通过 `go build` 编译该依赖（默认 build 使用 `!macro` 生成侧）

#### Scenario: 应用项目可选策略

- **WHEN** 项目不作为对外库发布
- **THEN** 文档 MAY 允许仅在本模块 CI 中 `expand` 而不提交 gen，但 MUST 明确此方式不适用于被依赖的库

### Requirement: examples 参考 expand 入口（本仓库）

本仓库 MUST 在 **examples** 子 module 提供参考实现 `cmd/macroexpand`（路径 `github.com/arcane-craft/go-macro/examples/cmd/macroexpand`）。该实现 MUST 仅 blank import 所需 `register` 包（含 `github.com/arcane-craft/go-macro-contrib/register`）并调用 `expandtool.Main()`，MUST NOT 包含其它业务逻辑。

根 module MUST NOT 包含 `cmd/macroexpand`（布局细节以 `macro-repo-layout` 为准）。

该路径为宏使用方 **RECOMMENDED** 快速上手命令，MUST NOT 被解释为唯一允许的 expand 入口。

#### Scenario: 参考入口与 expandtool Main 等价

- **WHEN** 用户 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .` 且进程已 link 所需 `register`
- **THEN** 行为 MUST 与在同一进程内调用 `expandtool.Main()` 一致

#### Scenario: 调用方自建等价入口

- **WHEN** 宏使用方在项目内提供等价 `cmd/macroexpand`（blank import 所需 register 并调用 `expandtool.Main()`）
- **THEN** expand 行为 MUST 与使用本仓库 examples 参考入口等价（就 expand 语义而言）

### Requirement: init provider 生成 register 而非 expand 工具

`go tool macro init provider` MUST 为**宏作者**生成 provider 骨架，含 `register/register.go`：在 `init` 中 `expandtool.Register(<module>/provider/import/path>, ProviderExpand)`。

MUST NOT 生成 `tools/macroexpand` 或要求宏作者实现 expand main。README MUST 说明宏**使用方**须承载 expand 入口（register + `expandtool.Main()`），并 **RECOMMENDED** 提供 `examples/cmd/macroexpand` 的 generate 一行作为快速上手模板。

#### Scenario: 宏作者无 expand main 义务

- **WHEN** 用户执行 `go tool macro init provider mymac`
- **THEN** 输出 MUST 含 `register/register.go` 且 MUST NOT 含 `tools/macroexpand/main.go`

### Requirement: 消费第三方宏库的附录路径

当宏使用方除 `go-macro-contrib` 外还依赖其它带 `register` 子包的宏库时，文档 MAY 说明：复制 `examples/cmd/macroexpand` 为项目内 `cmd/macroexpand` 并**仅**追加 blank import 该宏库的 `register` 包，仍调用 `expandtool.Main()`。该路径 MUST NOT 作为默认快速上手内容。

#### Scenario: 附录不增加宏作者负担

- **WHEN** 第三方宏作者按脚手架发布 `register` 子包
- **THEN** 宏作者 MUST NOT 需要维护 expand 二进制；链接责任在使用方可选 cmd 或未来框架扩展

