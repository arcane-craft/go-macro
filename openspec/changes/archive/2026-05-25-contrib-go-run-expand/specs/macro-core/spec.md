## ADDED Requirements

### Requirement: expandtool 展开入口

`macro` 包 MUST 提供子包 `macro/expandtool`，供 expand 二进制与高级集成使用，至少包含：

- `Register(importPath string, expand Expander)` — 注册可 link 的 Expander
- `Registered() map[string]Expander` — 返回当前已注册副本
- `Run(args []string, linked map[string]Expander) error` — `args` 为空时 MUST 默认 `[]string{"./..."}`；`linked` 为 `nil` 时使用 `Registered()`；内部调用 `expander.ExpandPackages`
- `Main()` — `Run(os.Args[1:], nil)`，错误时 MUST 写 stderr 并以非零状态退出

宏作者（provider 包作者）MUST NOT 被要求实现或维护 expand main；`Main`/`Run` 由 **examples/cmd/macroexpand**（或用户自建等价 cmd）调用。

#### Scenario: Main 使用 Registered 注册表

- **WHEN** `contrib/register` 已在进程内 `init` 中 Register 官方宏库，且用户执行 `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .`
- **THEN** MUST 展开宏主文件中已 import 且已 Register 的官方宏调用

#### Scenario: Run 默认包路径

- **WHEN** 调用 `Run(nil, linked)` 或 `Run([]string{}, linked)`
- **THEN** MUST 等价于对 `./...` 执行 expand

## MODIFIED Requirements

### Requirement: 通用宏注册与查找

系统 MUST 支持：对**宏主文件所在包已 import 的**宏库包扫描 `//macro: <syntax-id>` 与 panic 桩，并结合 **`linked map[string]macro.Expander`**（来自 `expandtool.Registered()` 或显式传入）构建注册表。`internal/expander` 与 `macro/expandtool` MUST NOT 硬编码任何具体宏库 Expander。

#### Scenario: 仅注册已 import 且已 link 的宏库

- **WHEN** 宏库已 `expandtool.Register`，但宏主文件未 import 该包
- **THEN** 注册表 MUST NOT 包含该包的桩

#### Scenario: 注册多桩名到同一 syntax-id

- **WHEN** 宏库包中 `//macro: syntax-try` 绑定 `TryExpand`，存在桩 `Try` 与 `Try2`，且 `linked` 含该 import path
- **THEN** 注册表 MUST 将 `Try`、`Try2` 均映射到同一 `TryExpand`

## REMOVED Requirements

### Requirement: 官方宏库（可选依赖）

**Reason**: 官方宏库迁至 contrib；Expander 链接改由 `expandtool.Register` + `contrib/register` + **examples/cmd/macroexpand** 承担。

**Migration**:

1. import 改为 `github.com/arcane-craft/go-macro/contrib/try` 等
2. `//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .`
3. 删除项目内 `tools/macroexpand`（若存在）
