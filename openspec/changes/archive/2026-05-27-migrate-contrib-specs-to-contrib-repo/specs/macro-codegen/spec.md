## MODIFIED Requirements

### Requirement: expand 子命令自动 link

`cmd/macro expand` MUST 自动生成并更新 `.gomacro/expand_runner`，依据 provider 上的 `//macro:` 指令进行 Expander link。官方宏库（`inline`、`try`）的 provider 契约与路径以 `go-macro-contrib` 仓库 OpenSpec（`macro-contrib`、`syntax-inline`、`syntax-try`）为准。

#### Scenario: import 即可 link

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro-contrib/try` 并执行 expand
- **THEN** MUST 自动 link `TryExpand`，无需 blank import `register`
