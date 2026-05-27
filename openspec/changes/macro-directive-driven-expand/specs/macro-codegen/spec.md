## MODIFIED Requirements

### Requirement: go generate 集成

工具链 MUST 支持在宏主文件中通过**一行** generate 触发 expand，**无需**用户项目内 `cmd/macroexpand` 或手写 `register` 包。

RECOMMENDED generate 指令：

```go
//go:generate go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .
```

（整模块可用 `expand ./...`；按包展开用 `expand .`。）

`expand` 子命令 MUST：根据待展开包的 import 与 provider 上的 `//macro:` Expander 注解，生成或更新模块内 link 源码（含 `expandtool.Register` 与 blank import），再执行 `expandtool.Run` 写回 `*_macro_gen.go`。

#### Scenario: generate 零自建 expand 入口

- **WHEN** 用户项目无 `cmd/macroexpand`、无 `tools/macroexpand`，宏主文件含上述 RECOMMENDED generate，且已 import 所用 provider
- **THEN** `go generate` MUST 成功写回 `*_macro_gen.go`

#### Scenario: 按包 generate

- **WHEN** 宏主文件含 `//go:generate go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .`
- **THEN** MUST 仅展开该 generate 所在包（或指令指定的 patterns）

## ADDED Requirements

### Requirement: expand 子命令与 link 代码生成

`github.com/arcane-craft/go-macro/cmd/macro` MUST 提供 `expand` 子命令，作为宏使用方默认展开入口。

`expand` MUST 在目标 Go module 根目录（含 `go.mod`）下生成或更新 link 辅助源码（实现路径由 design 约定，如 `.gomacro/expand_link.go`），其 `init` MUST 对本次展开所需的 provider import path 调用 `expandtool.Register`。

文档 RECOMMENDED 将 link 生成目录加入 `.gitignore`，并在 CI 中于 `go test` 前执行 `expand`（与提交 `*_macro_gen.go` 策略一致）。

#### Scenario: 首次 expand 生成 link

- **WHEN** 模块尚无 link 文件，用户执行 `cmd/macro expand .` 且宏主文件 import 含 `//macro:` Expander 的 provider
- **THEN** MUST 创建 link 文件并完成展开，且 MUST 写回 `*_macro_gen.go`

#### Scenario: import 变更后更新 link

- **WHEN** 宏主文件新增 import 另一 provider，用户再次执行 `expand`
- **THEN** link 文件 MUST 更新以 Register 新 provider，且展开 MUST 包含新桩

## REMOVED Requirements

### Requirement: generate 依赖旧 examples 入口

**Reason**: 展开入口收敛为 `cmd/macro expand`；不再要求 blank import `contrib/register` 或运行 examples 内参考 main。

**Migration**: 将 `//go:generate` 改为 `go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .`；删除旧手写展开入口与手工注册接线代码。
