## MODIFIED Requirements

### Requirement: 快速上手保留 normative 要点

`快速上手` MUST 包含且保持与 `macro-codegen` 一致的下列语义：

- **首要步骤**：在使用宏的项目中，通过 `go run github.com/arcane-craft/go-macro/cmd/macro@latest expand`（或 `//go:generate` 等价一行）展开宏
- 以 `examples/readfile` 等作为对照示例
- 宏主文件（`//go:build macro`）中 RECOMMENDED generate：`go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .`
- 须 import 所用宏库
- 对外发布 SHOULD 提交 `*_macro_gen.go`
- 日常 `go build` / `go test` 使用生成侧，不依赖长期 `-tags macro`

#### Scenario: 快速上手与 macro-codegen 对齐

- **WHEN** 读者对照 `macro-codegen` 中 expand 子命令要求
- **THEN** README `快速上手` MUST 以 `cmd/macro expand` 为默认入口，且 MUST NOT 要求先创建项目内 `cmd/macroexpand`

#### Scenario: 不再推荐 examples macroexpand 为默认

- **WHEN** 读者阅读 README `快速上手` 全文
- **THEN** MUST NOT 将任何旧参考入口作为宏使用方默认推荐命令

## MODIFIED Requirements

### Requirement: README 命令节与 cmd/macro 调用方式

README `命令` 节 MUST 列出：

- `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>` — 初始化宏库
- `go run github.com/arcane-craft/go-macro/cmd/macro@latest expand [patterns...]` — 展开宏（默认展开入口）

README MUST NOT 将 `go tool macro` 作为文档中的主要调用方式。

#### Scenario: expand 命令可见

- **WHEN** 读者在 README `命令` 节查找如何展开宏
- **THEN** MUST 看到 `cmd/macro expand` 子命令说明
