## Why

`go tool macro expand` 无法在编译期由用户选择要 link 的 `Expander`，而 `expander/official_providers.go` 又将官方宏库硬编码进引擎，与「contrib 独立、框架提供 expand 入口」的目标冲突。Go 要求 `Expander` 必须在 expand 二进制编译期链接；**极简 expand 入口与注册表 MUST 由 `macro` 库与 examples 提供的 expand 二进制实现**，宏作者（provider）只写 `Expand` + 可选 `register` 子包，**不**承担 expand 工具或 `tools/macroexpand` 维护。

## What Changes

- **BREAKING**：删除 `go tool macro expand` 子命令；`cmd/macro` 仅保留 `init provider`
- **BREAKING**：根目录 `inline/`、`try/` 迁入 `contrib/` 独立子 module
- **BREAKING**：删除 `expander/official_providers.go`；根 module 的 `expander` 不再 import 任何宏库
- **BREAKING**：官方 `cmd/macroexpand` 迁至 **`examples/cmd/macroexpand`**（examples 独立 module）；根 `go.mod` **不** require contrib
- `ExpandPackages` API 改为 `linked map[string]macro.Expander`
- **新增 `macro/expandtool`**：`Register`、`Registered`、`Run`、`Main`
- **`contrib/register`**：在 `init` 中 `expandtool.Register` 官方 inline/try
- 宏主文件 generate：**一行** `//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .`
- `init provider` 为宏作者生成 **`register/register.go`**（非 `tools/macroexpand`）
- **根 module 测试不 import contrib**；readfile 等集成测试在 **examples** module
- 仓库根 **`go.work`** 联调根 / contrib / examples
- 通用化 `internal/codegen/imports.go`

## Capabilities

### New Capabilities

- `macro-contrib`：contrib 子 module、官方宏库路径、`contrib/register` 注册官方 Expander
- `macro-repo-layout`：三 module 布局、examples 承载 macroexpand、根测试边界

### Modified Capabilities

- `macro-core`：新增 `macro/expandtool`；移除官方目录；linked 注册模型
- `macro-codegen`：删除 go tool expand；canonical generate 指向 `examples/cmd/macroexpand`
- `macro-expander`：`ExpandPackages` + linked 激活；根测试不依赖 contrib
- `syntax-inline` / `syntax-try`：路径迁至 contrib；示例在 examples module

## Impact

- **API**：`macro/expandtool`；`expander.ExpandPackages(patterns, linked)`
- **用户（宏使用方）**：contrib import 路径 + generate 指向 `examples/cmd/macroexpand`；**无需** `tools/macroexpand`
- **宏作者（provider）**：可选 `register` 子包由脚手架生成；**不**实现 expand main
- **根 go.mod**：仅核心库依赖；**contrib 由 examples module 引用**
- **开发**：根目录 `go.work` 覆盖三 module
