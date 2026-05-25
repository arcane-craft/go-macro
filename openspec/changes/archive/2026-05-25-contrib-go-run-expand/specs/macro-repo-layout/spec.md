## ADDED Requirements

### Requirement: 多 module 仓库布局

本仓库 MUST 以三个独立 Go module 组织，路径如下：

| 目录 | module 路径 | 职责 |
|------|-------------|------|
| 仓库根 | `github.com/arcane-craft/go-macro` | 核心库：`macro/`、`internal/expander/`、`internal/codegen/` 等；`cmd/macro`（仅 `init provider`） |
| `contrib/` | `github.com/arcane-craft/go-macro/contrib` | 官方宏库 `inline/`、`try/`；`register/`（`expandtool.Register`） |
| `examples/` | `github.com/arcane-craft/go-macro/examples` | 示例与官方 expand 二进制 `cmd/macroexpand/` |

根 module MUST NOT 再包含顶层 `inline/`、`try/` 或根级 `cmd/macroexpand/`。

#### Scenario: 根 module 仅承载框架库

- **WHEN** 查看根目录 `go.mod`
- **THEN** MUST 仅 `require` 核心构建所需依赖（如 `golang.org/x/tools`），MUST NOT `require github.com/arcane-craft/go-macro/contrib`

### Requirement: 官方 expand 二进制位于 examples module

官方 expand 入口 MUST 为 `examples/cmd/macroexpand`，另 module 路径：

`github.com/arcane-craft/go-macro/examples/cmd/macroexpand`

另实现 MUST 仅 blank import `contrib/register` 并调用 `expandtool.Main()`。`contrib` 依赖 MUST 落在 **examples**（及 contrib）module，而非根 module。

宏使用方 canonical generate 一行：

```go
//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .
```

#### Scenario: go run 官方 macroexpand

- **WHEN** 用户于宏主文件所在 module 执行上述 `go run`
- **THEN** MUST 编译 examples 下的 macroexpand 并成功展开已 import 且已 link 的宏

### Requirement: 根 module 测试不依赖 contrib

根 module（`github.com/arcane-craft/go-macro`）内所有 `*_test.go` MUST NOT import `github.com/arcane-craft/go-macro/contrib/...` 任何包（含 `inline`、`try`、`register`）。

依赖真实 `TryExpand` / `InlineExpand` 的宏库单测在 `contrib` module；`examples` 仅保留示例包内 golden 等轻量测试（不强制 module 级 expand 集成测试）。

#### Scenario: 根 go test 无 contrib 边

- **WHEN** 于仓库根执行 `GOWORK=off go test ./...`
- **THEN** MUST 通过且根 `go.mod` MUST NOT 因测试而引入 `contrib` require

#### Scenario: examples 轻量测试

- **WHEN** 在 `examples` 目录执行 `go test ./...`
- **THEN** MUST 通过（含 `readfile` 对已提交 `*_macro_gen.go` 的 golden 校验等）

### Requirement: 本地开发 workspace

仓库根 MUST 提供 `go.work`，且 `use` MUST 至少包含根 module、`./contrib`、`./examples`，以便同仓联调与跨 module 测试。

#### Scenario: workspace 联调

- **WHEN** 开发者在仓库根执行 `go test ./...`（默认启用 `go.work`）
- **THEN** MUST 能同时测试根、contrib、examples 三个 module

### Requirement: 示例包路径

`examples/readfile` 包路径 MUST 为 `github.com/arcane-craft/go-macro/examples/readfile`（属于 examples module）。示例宏主文件 MUST import `contrib/try`（或所需官方库）并使用 examples module 下的 macroexpand generate。

#### Scenario: readfile 示例路径

- **WHEN** expander 或 examples 测试加载 readfile
- **THEN** 包路径 MUST 为 `github.com/arcane-craft/go-macro/examples/readfile`
