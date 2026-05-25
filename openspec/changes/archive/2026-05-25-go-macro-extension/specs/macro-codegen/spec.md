## ADDED Requirements

### Requirement: go tool macro CLI

项目 MUST 提供可通过 `go tool macro` 调用的命令行工具，支持 `expand` 子命令对指定包路径执行宏展开。

`cmd/macro` 的 `expand` **MUST NOT** 在编译期硬编码或默认注册本仓库官方宏库（`inline`、`try`）；官方库由宏主文件 `import` 触发，经 `expander` 官方目录衔接（见 `macro-core` / `macro-expander` spec）。

#### Scenario: 展开当前模块包

- **WHEN** 用户在模块根目录执行 `go tool macro expand ./...`
- **THEN** 工具 MUST 扫描所有包、执行展开并写回生成文件

#### Scenario: expand 不携带内置 provider 列表

- **WHEN** `cmd/macro` 调用 `expander.ExpandPackages`
- **THEN** 传入的额外 provider 列表 MUST 为空（`nil` 或零长度），不得将 `inline`/`try` 的 `Expand` 作为 CLI 默认参数

### Requirement: go generate 集成

工具 MUST 支持在源文件中通过 `//go:generate go tool macro expand` 触发与 CLI 等价的展开流程。

#### Scenario: generate 钩子展开

- **WHEN** 用户执行 `go generate ./...` 且包内含 generate 指令
- **THEN** 对应包的宏调用 MUST 被展开且生成文件被更新

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

对同一输入源码重复执行展开，工具 MUST 产生相同的生成文件内容（除时间戳注释外可配置忽略）。

#### Scenario: 重复执行 expand

- **WHEN** 连续两次执行 `go tool macro expand` 且主文件未变
- **THEN** `foo_macro_gen.go` 内容 MUST 一致

### Requirement: 仅展开当前主模块

`go tool macro expand` MUST 仅处理**当前主模块**内的包（如 `expand ./...`）。MUST NOT 修改 module cache 或依赖模块源码树。

#### Scenario: 不展开依赖模块

- **WHEN** 某依赖库在 module cache 中含 macro 主文件但未提交 `*_macro_gen.go`
- **THEN** expand 命令 MUST NOT 写入该依赖路径；文档 MUST 要求库作者在发布前于**其自身仓库**内 expand 并提交生成物

#### Scenario: 本模块 expand

- **WHEN** 用户于主模块根目录执行 `go tool macro expand ./...`
- **THEN** 工具 MUST 仅更新本模块内的 `*_macro_gen.go`

### Requirement: 对外库须提交生成物

文档与示例 MUST 规定：若模块会被他人在 `go.mod` 中 `require` 引用，维护者 **MUST** 在版本库中提交与宏主文件配对的 `*_macro_gen.go`（发布前已执行 expand）。

#### Scenario: 下游无需 macro 工具链

- **WHEN** 库维护者已提交 `foo_macro_gen.go` 并发布，下游 `go get` 该版本
- **THEN** 下游在未安装 `go-macro` 时 MUST 能仅通过 `go build` 编译该依赖（默认 build 使用 `!macro` 生成侧）

#### Scenario: 应用项目可选策略

- **WHEN** 项目不作为对外库发布
- **THEN** 文档 MAY 允许仅在本模块 CI 中 `expand` 而不提交 gen，但 MUST 明确此方式不适用于被依赖的库
