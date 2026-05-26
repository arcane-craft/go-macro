## MODIFIED Requirements

### Requirement: 多 module 仓库布局

本仓库 MUST 以三个独立 Go module 组织，路径如下：

| 目录 | module 路径 | 职责 |
|------|-------------|------|
| 仓库根 | `github.com/arcane-craft/go-macro` | 核心库：`macro/`、`internal/expander/`、`internal/codegen/` 等；`cmd/macro`（仅 `init provider`） |
| `contrib/` | `github.com/arcane-craft/go-macro/contrib` | 官方宏库 `inline/`、`try/`；`register/`（`expandtool.Register`） |
| `examples/` | `github.com/arcane-craft/go-macro/examples` | 示例宏调用方工程；包含参考 `cmd/macroexpand/` 接线实现 |

根 module MUST NOT 再包含顶层 `inline/`、`try/` 或根级 `cmd/macroexpand/`。

#### Scenario: 根 module 仅承载框架库

- **WHEN** 查看根目录 `go.mod`
- **THEN** MUST 仅 `require` 核心构建所需依赖（如 `golang.org/x/tools`），MUST NOT `require github.com/arcane-craft/go-macro/contrib`

### Requirement: 调用方项目承载 expand 入口，examples 提供参考实现

宏展开入口 MUST 由宏调用方项目承载（如项目内 `cmd/macroexpand`），并且实现 MUST 通过 blank import 所需 `register` 包后调用 `expandtool.Main()`（或与 `Run` 等价的接线模式）。

`examples/cmd/macroexpand` MUST 作为参考实现存在于 examples module；其路径 `github.com/arcane-craft/go-macro/examples/cmd/macroexpand` 为 RECOMMENDED 默认用法，但 MUST NOT 被解释为唯一允许路径。

`contrib` 依赖 MUST 落在调用方工程 module（例如 examples）与 contrib module，而非根 module。

宏使用方推荐 generate 一行：

```go
//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .
```

#### Scenario: 使用 examples 参考入口

- **WHEN** 用户于宏主文件所在 module 执行 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .`
- **THEN** MUST 编译 examples 下参考入口并成功展开已 import 且已 link 的宏

#### Scenario: 使用调用方自建入口

- **WHEN** 用户在自身项目内实现等价 `cmd/macroexpand`（blank import register 并调用 `expandtool.Main()`）
- **THEN** MUST 在行为上与参考入口等价，并成功展开已 import 且已 link 的宏
