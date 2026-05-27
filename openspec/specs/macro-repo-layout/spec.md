# macro-repo-layout Specification

## Purpose
TBD - created by archiving change contrib-go-run-expand. Update Purpose after archive.
## Requirements
### Requirement: 多 module 仓库布局

`go-macro` 仓库 MUST 以两个独立 Go module 组织，路径如下：

| 目录 | module 路径 | 职责 |
|------|-------------|------|
| 仓库根 | `github.com/arcane-craft/go-macro` | 核心库：`macro/`、`internal/expander/`、`internal/codegen/` 等；`cmd/macro`（`init provider`、`expand`） |
| `examples/` | `github.com/arcane-craft/go-macro/examples` | 示例宏调用方工程（宏主文件 + `go:generate`） |

官方宏库（`inline`、`try`）MUST 位于独立仓库 `github.com/arcane-craft/go-macro-contrib`，不在本仓库目录树内；其 OpenSpec 规范 MUST 位于 contrib 仓库（见「官方宏库 OpenSpec 归属 contrib 仓库」）。

根 module MUST NOT 再包含顶层 `inline/`、`try/`、`contrib/`，也 MUST NOT 要求维护 `cmd/macroexpand/`。

#### Scenario: 根 module 仅承载框架库

- **WHEN** 查看 `go-macro` 根目录 `go.mod`
- **THEN** MUST 仅 `require` 核心构建所需依赖（如 `golang.org/x/tools`），MUST NOT `require github.com/arcane-craft/go-macro-contrib`

### Requirement: 调用方项目通过 cmd/macro expand 承载展开

宏展开入口 MUST 由 `cmd/macro expand` 提供；宏调用方项目通过 `//go:generate` 或命令行调用该子命令触发展开。MUST NOT 要求调用方手写 `cmd/macroexpand` 与 `register` 接线代码。

官方宏库依赖 MUST 落在调用方工程 module（例如 examples）与 `go-macro-contrib` module，而非 `go-macro` 根 module。

宏使用方推荐 generate 一行：

```go
//go:generate go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .
```

#### Scenario: 使用 cmd/macro expand

- **WHEN** 用户于宏主文件所在 module 执行 `go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .`
- **THEN** MUST 成功展开已 import 的宏调用并写回生成文件

#### Scenario: 仅维护宏主文件与 generate

- **WHEN** 用户项目不包含 `cmd/macroexpand` 目录，仅在宏主文件维护 `go:generate`
- **THEN** MUST 仍可通过 `cmd/macro expand` 完成展开

### Requirement: 根 module 测试不依赖 contrib

`go-macro` 根 module 内所有 `*_test.go` MUST NOT import `github.com/arcane-craft/go-macro-contrib/...` 任何包（含 `inline`、`try`）。

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

### Requirement: 官方宏库 OpenSpec 归属 contrib 仓库

`go-macro` 仓库的 `openspec/specs/` MUST NOT 包含 `macro-contrib`、`syntax-inline`、`syntax-try` 能力的主规范文件。上述能力的权威 OpenSpec MUST 位于 `github.com/arcane-craft/go-macro-contrib` 仓库的 `openspec/specs/<capability>/spec.md`。

`go-macro` 的 `macro-repo-layout`、`macro-expander` 等规范 MAY 引用 contrib 仓规范名称与 import 路径，MUST NOT 重复抄写 provider 级展开语义全文。

#### Scenario: core 主规范列表无 contrib 能力

- **WHEN** 维护者于 `go-macro` 执行 `openspec list`（或等价方式列出主 spec）
- **THEN** MUST NOT 出现 `macro-contrib`、`syntax-inline`、`syntax-try` 作为主 spec 条目

#### Scenario: 查找 Try 载荷规则

- **WHEN** 贡献者需修改 `TryExpand` 的 k 校验或桩名规则
- **THEN** MUST 在 `go-macro-contrib` 仓库的 `syntax-try` spec 中修改，而非 `go-macro`

### Requirement: readfile Try 端到端示例

`go-macro` 仓库 MUST 在 **examples module** 包含 `readfile` 示例（`Try` 用于 k=1），包路径 `github.com/arcane-craft/go-macro/examples/readfile`。示例 MUST import `github.com/arcane-craft/go-macro-contrib/try`，经 `go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .` generate 后 `go test` 与 golden 一致。

#### Scenario: ReadFile 黄金测试

- **WHEN** 在 `examples/readfile` 执行 `go test`（对照已提交的 `readfile_macro_gen.go` 与 golden）
- **THEN** 测试 MUST 通过且 golden 匹配

