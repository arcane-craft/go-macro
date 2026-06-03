## MODIFIED Requirements

### Requirement: 多 module 仓库布局

`go-macro` 仓库 MUST 以两个独立 Go module 组织，路径如下：

| 目录 | module 路径 | 职责 |
|------|-------------|------|
| 仓库根 | `github.com/arcane-craft/go-macro` | 核心库：`macro/`、`internal/expander/`、`internal/codegen/` 等；`cmd/macro`（`init provider`、`expand`） |
| `examples/` | `github.com/arcane-craft/go-macro/examples` | 示例宏调用方工程（宏主文件 + `go:generate`） |

官方宏库（`inline`、`try`、`derivestringer`、`wirejson`）MUST 位于独立仓库 `github.com/arcane-craft/go-macro-contrib`，不在本仓库目录树内；其 OpenSpec 规范 MUST 位于 contrib 仓库（见「官方宏库 OpenSpec 归属 contrib 仓库」）。

根 module MUST NOT 再包含顶层 `inline/`、`try/`、`contrib/`、`syntax/` 或官方语法实现包副本。

#### Scenario: 根 module 仅承载框架库

- **WHEN** 查看 `go-macro` 根目录 `go.mod`
- **THEN** MUST 仅 `require` 核心构建所需依赖（如 `golang.org/x/tools`），MUST NOT `require github.com/arcane-craft/go-macro-contrib`

### Requirement: 官方宏库 OpenSpec 归属 contrib 仓库

`go-macro` 仓库的 `openspec/specs/` MUST NOT 包含 `macro-contrib`、`syntax-inline`、`syntax-try`、`syntax-derive-stringer`、`syntax-wire-json` 能力的主规范文件。上述能力的权威 OpenSpec MUST 位于 `github.com/arcane-craft/go-macro-contrib` 仓库的 `openspec/specs/<capability>/spec.md`。

`go-macro` 的 `macro-repo-layout`、`macro-expander` 等规范 MAY 引用 contrib 仓规范名称与 import 路径，MUST NOT 重复抄写 provider 级展开语义全文。

#### Scenario: core 主规范列表无 contrib 能力

- **WHEN** 维护者于 `go-macro` 执行 `openspec list`（或等价方式列出主 spec）
- **THEN** MUST NOT 出现 `macro-contrib`、`syntax-inline`、`syntax-try`、`syntax-derive-stringer`、`syntax-wire-json` 作为主 spec 条目

#### Scenario: 查找 derive-stringer 规则

- **WHEN** 贡献者需修改 `DeriveStringerExpand` 语义
- **THEN** MUST 在 `go-macro-contrib` 仓库的 `syntax-derive-stringer` spec 中修改，而非 `go-macro`
