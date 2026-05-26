## MODIFIED Requirements

### Requirement: 可选官方宏库与引入方式

`try` 包 MUST 在 `go-macro-contrib` 仓库内作为官方宏库发布，路径为 `github.com/arcane-craft/go-macro-contrib/try`。使用方 MUST 在宏主文件中 import 该路径，且 expand 工具的 `linked` map MUST 包含该 import path 与 `TryExpand`，方可展开 `Try` 族调用。

#### Scenario: 未 import 时不展开

- **WHEN** 宏主文件使用 `Try(...)` 但未 import `github.com/arcane-craft/go-macro-contrib/try`
- **THEN** 展开管线 MUST NOT 注册 `syntax-try`

#### Scenario: import 但未 link 时不展开

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro-contrib/try`，但 expand 工具 `linked` 未含该 path
- **THEN** 对 `Try(...)` 的调用 MUST NOT 被展开

### Requirement: Try 端到端示例

仓库 MUST 在 **examples module** 包含 `readfile` 示例（`Try` 用于 k=1），包路径 `github.com/arcane-craft/go-macro/examples/readfile`。示例 MUST import `github.com/arcane-craft/go-macro-contrib/try`，经 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand` generate 后 `go test` 与 golden 一致。

#### Scenario: ReadFile 黄金测试

- **WHEN** 在 `examples/readfile` 执行 `go test`（对照已提交的 `readfile_macro_gen.go` 与 golden）
- **THEN** 测试 MUST 通过且 golden 匹配
