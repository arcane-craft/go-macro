## MODIFIED Requirements

### Requirement: expandtool 展开入口

`macro` 包 MUST 提供子包 `macro/expandtool`，供 expand 二进制与高级集成使用，至少包含：

- `Register(importPath string, expand Expander)` — 注册可 link 的 Expander
- `Registered() map[string]Expander` — 返回当前已注册副本
- `Run(args []string, linked map[string]Expander) error` — `args` 为空时 MUST 默认 `[]string{"./..."}`；`linked` 为 `nil` 时使用 `Registered()`；内部调用 `expander.ExpandPackages`
- `Main()` — `Run(os.Args[1:], nil)`，错误时 MUST 写 stderr 并以非零状态退出

框架 MUST 仅负责提供上述展开能力与接线模式；provider 作者 MUST NOT 被要求实现或维护 expand main。具体可执行入口 MUST 由宏调用方项目承载，`examples/cmd/macroexpand` 可作为参考实现，但 MUST NOT 被视为唯一允许入口。

#### Scenario: Main 使用 Registered 注册表

- **WHEN** 某调用方入口进程已通过 `register` 包 `init` 调用 `expandtool.Register` 注册宏库，且执行该入口命令
- **THEN** MUST 展开宏主文件中已 import 且已 Register 的宏调用

#### Scenario: Run 默认包路径

- **WHEN** 调用 `Run(nil, linked)` 或 `Run([]string{}, linked)`
- **THEN** MUST 等价于对 `./...` 执行 expand
