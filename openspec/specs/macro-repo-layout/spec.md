# macro-repo-layout Specification

## Purpose
TBD - created by archiving change contrib-go-run-expand. Update Purpose after archive.
## Requirements
### Requirement: 多 module 仓库布局

`go-macro` 仓库 MUST 以两个独立 Go module 组织，路径如下：

| 目录 | module 路径 | 职责 |
|------|-------------|------|
| 仓库根 | `github.com/arcane-craft/go-macro` | 核心库：`macro/`、`internal/expander/`、`internal/codegen/` 等；`cmd/macro`（仅 `init provider`） |
| `examples/` | `github.com/arcane-craft/go-macro/examples` | 示例宏调用方工程；包含参考 `cmd/macroexpand/` 接线实现 |

官方宏库（`inline`、`try`、`register`）MUST 位于独立仓库 `github.com/arcane-craft/go-macro-contrib`，不在本仓库目录树内。

根 module MUST NOT 再包含顶层 `inline/`、`try/`、`contrib/` 或根级 `cmd/macroexpand/`。

#### Scenario: 根 module 仅承载框架库

- **WHEN** 查看 `go-macro` 根目录 `go.mod`
- **THEN** MUST 仅 `require` 核心构建所需依赖（如 `golang.org/x/tools`），MUST NOT `require github.com/arcane-craft/go-macro-contrib`

### Requirement: 调用方项目承载 expand 入口，examples 提供参考实现

宏展开入口 MUST 由宏调用方项目承载（如项目内 `cmd/macroexpand`），并且实现 MUST 通过 blank import 所需 `register` 包后调用 `expandtool.Main()`（或与 `Run` 等价的接线模式）。

`examples/cmd/macroexpand` MUST 作为参考实现存在于 examples module；其路径 `github.com/arcane-craft/go-macro/examples/cmd/macroexpand` 为 RECOMMENDED 默认用法，但 MUST NOT 被解释为唯一允许路径。

官方宏库依赖 MUST 落在调用方工程 module（例如 examples）与 `go-macro-contrib` module，而非 `go-macro` 根 module。

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

### Requirement: 根 module 测试不依赖 contrib

`go-macro` 根 module 内所有 `*_test.go` MUST NOT import `github.com/arcane-craft/go-macro-contrib/...` 任何包（含 `inline`、`try`、`register`）。

依赖真实 `TryExpand` / `InlineExpand` 的宏库单测在 `go-macro-contrib` 仓库；`examples` 仅保留示例包内 golden 等轻量测试（不强制 module 级 expand 集成测试）。

#### Scenario: 根 go test 无 contrib 边

- **WHEN** 于 `go-macro` 仓库根执行 `GOWORK=off go test ./...`
- **THEN** MUST 通过且根 `go.mod` MUST NOT 因测试而引入 `go-macro-contrib` require

#### Scenario: examples 轻量测试

- **WHEN** 在 `go-macro` 的 `examples` 目录执行 `go test ./...`（examples `go.mod` 已 require 兼容版本的 `go-macro-contrib`）
- **THEN** MUST 通过（含 `readfile` 对已提交 `*_macro_gen.go` 的 golden 校验等）

### Requirement: 本地开发 workspace

`go-macro` 仓库根 **MAY** 提供 `go.work`；若提供，其 `use` **MUST** 仅包含根 module（`.`）与 `./examples`，**MUST NOT** `use` 已迁出的 `contrib` 或 sibling `go-macro-contrib` 路径。

仓库 **MUST NOT** 要求根目录必须存在 `go.work` 才能满足本规范。根 module 与 `examples` module 的测试 **MUST** 可按 module 边界分别执行（见下方场景）。

`examples/go.mod` 的**已提交**依赖 **MUST** 通过版本化 `require` 引用已发布的 `github.com/arcane-craft/go-macro` 与 `github.com/arcane-craft/go-macro-contrib` tag（与当前仓库 tag 策略兼容的版本号）。**MAY** 在已提交文件中保留 `replace github.com/arcane-craft/go-macro => ../`，以便在本仓开发核心时联调 examples。

对 `go-macro-contrib` 的本地并行开发，开发者 **MAY** 在本地（含未提交变更）向 `examples/go.mod` 添加 `replace github.com/arcane-craft/go-macro-contrib => ../go-macro-contrib`（contrib checkout 位于 `go-macro` 仓库根同级目录 `../go-macro-contrib`，路径相对 `examples/go.mod` 所在目录）。**MUST NOT** 要求该 contrib `replace` 必须出现在已提交的 `examples/go.mod` 中。

#### Scenario: 根 module 独立测试

- **WHEN** 于 `go-macro` 仓库根执行 `GOWORK=off go test ./...`
- **THEN** MUST 通过且仅覆盖根 module 包

#### Scenario: examples module 独立测试

- **WHEN** 于 `go-macro/examples` 目录执行 `go test ./...`，且 `examples/go.mod` 已 `require` 兼容的已发布 `go-macro-contrib` 版本（或开发者本地已添加 contrib `replace`）
- **THEN** MUST 通过（含 `readfile` golden 等）

#### Scenario: 可选 workspace 联调本仓两 module

- **WHEN** 开发者于 `go-macro` 根提供 `go.work`（`use` 为 `.` 与 `./examples`）并在根执行 `go test ./...`
- **THEN** MUST 能同时测试根 module 与 examples module

### Requirement: 示例包路径

`examples/readfile` 包路径 MUST 为 `github.com/arcane-craft/go-macro/examples/readfile`（属于 examples module）。示例宏主文件 MUST import `github.com/arcane-craft/go-macro-contrib/try`（或所需官方库）并使用 examples module 下的 macroexpand generate。

#### Scenario: readfile 示例路径

- **WHEN** expander 或 examples 测试加载 readfile
- **THEN** 包路径 MUST 为 `github.com/arcane-craft/go-macro/examples/readfile`

