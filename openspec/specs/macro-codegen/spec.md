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

带宏的源文件 MUST 保持主文件名 `foo.go`；工具 MUST 生成/更新 `foo_macro_gen.go` 作为展开产物。用户 MUST NOT 因宏而重命名主文件。

#### Scenario: 主文件命名不变

- **WHEN** 开发者在 `readfile.go` 使用宏并执行 expand
- **THEN** MUST 更新 `readfile_macro_gen.go` 且 MUST NOT 要求改主文件后缀

#### Scenario: 默认构建仅使用生成侧

- **WHEN** 宏展开成功且用户执行 `go build`（无 `-tags macro`）
- **THEN** 编译 MUST 使用 `*_macro_gen.go`

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

写入 `*_macro_gen.go` 时，工具 MUST 生成 `//line <macro-main-file>:<line>`，以便错误与调试定位回宏主文件。

#### Scenario: panic 堆栈指向宏源码

- **WHEN** 展开后生成代码 panic
- **THEN** 堆栈 MUST 指向宏主文件行号

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

`cmd/macro expand` MUST 自动生成并更新 `.gomacro/expand_runner`，依据 provider 上的 `//macro:` 指令进行 Expander link。官方宏库（`inline`、`try`）的 provider 契约与路径以 `go-macro-contrib` 仓库 OpenSpec（`macro-contrib`、`syntax-inline`、`syntax-try`）为准。

#### Scenario: import 即可 link

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro-contrib/try` 并执行 expand
- **THEN** MUST 自动 link `TryExpand`，无需 blank import `register`

### Requirement: init provider 仅生成 provider 骨架

`init provider` MUST 生成桩与 Expander（含 per-function `//macro:`）及测试模板，MUST NOT 生成 `register/` 或 `tools/macroexpand`。

#### Scenario: 宏作者无 register 义务

- **WHEN** 用户执行 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac`
- **THEN** 输出 MUST 不包含 `register/register.go`

