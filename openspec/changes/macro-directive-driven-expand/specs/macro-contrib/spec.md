## REMOVED Requirements

### Requirement: 移除官方 register 接线机制

**Reason**: Expander link 改由 `cmd/macro expand` 根据 provider 上 `//macro:` 注解生成；contrib 不再提供集中 register 包。

**Migration**:

1. 自 `go-macro-contrib` 删除 `register/` 包
2. 为 `inline`、`try` 的每个桩函数与 `InlineExpand`/`TryExpand` 添加 per-function `//macro:`
3. 使用方 generate 改为 `go run .../cmd/macro@latest expand .`，移除对 `_ ".../register"` 的 blank import

## MODIFIED Requirements

### Requirement: 官方宏库路径

官方宏库 MUST 仅通过下列 import 路径提供：

- `github.com/arcane-craft/go-macro-contrib/inline`
- `github.com/arcane-craft/go-macro-contrib/try`

`go-macro` 根 module MUST NOT 再包含 `inline/`、`try/` 或 `contrib/`。

使用方 expand MUST 通过 `cmd/macro expand` 对 import 的 provider 生成 link；MUST NOT 依赖 `contrib/register` 包。

#### Scenario: import 新路径后 expand

- **WHEN** 宏主文件 import `github.com/arcane-craft/go-macro-contrib/try` 并执行 `cmd/macro expand`
- **THEN** MUST 展开 `Try` 族调用，且 MUST NOT 要求 blank import `register`
