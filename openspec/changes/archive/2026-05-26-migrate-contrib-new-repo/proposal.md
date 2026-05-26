## Why

`contrib` 已是独立 Go module，但仍作为 `go-macro`  monorepo 内的 `contrib/` 目录维护。官方宏库与框架核心分属不同发布节奏与职责边界；继续同仓会增加版本耦合、干扰 `go-macro` 仓库边界，也不利于第三方单独依赖或贡献官方宏库。将 `contrib` 迁入独立仓库可完成仓库级解耦，同时保留对 `go-macro` 的 module 依赖关系。

## What Changes

- **BREAKING**：新建独立仓库 `github.com/arcane-craft/go-macro-contrib`，将当前 `contrib/` 目录整体迁入并作为该仓库根 module 发布。
- **BREAKING**：所有官方宏库 import 路径由 `github.com/arcane-craft/go-macro/contrib/...` 改为 `github.com/arcane-craft/go-macro-contrib/...`（`inline`、`try`、`register`）。
- **BREAKING**：从 `go-macro` 仓库删除 `contrib/` 目录；`go.work` 不再 `use ./contrib`。
- 更新 `examples` module：`go.mod` 通过版本化 `require` 引用新 contrib module；本地开发 `replace` 指向 **`../go-macro-contrib`**。
- 更新 `contrib/register` 内 `expandtool.Register` 注册的 import path 字符串。
- 更新 README、author-guide、OpenSpec 与根 module 内仍引用旧路径的测试/fixture。
- `go-macro` 仓库布局由「三 module monorepo」变为「根 + examples」两 module；contrib 测试与发布在独立仓库执行。

## Capabilities

### New Capabilities

（无。contrib 能力由既有 `macro-contrib` 承载，仅变更仓库边界与 module 路径。）

### Modified Capabilities

- `macro-contrib`：contrib 为独立 Git 仓库与 module 根，不再位于 `go-macro/contrib/`；更新官方路径与 `register` 约定。
- `macro-repo-layout`：本仓库仅保留根 module 与 `examples/`；`go.work`、依赖边界、联调方式随 contrib 外迁调整。
- `macro-codegen`：文档与 RECOMMENDED generate 中涉及 contrib import 的表述对齐新路径。
- `syntax-inline`：`inline` 包发布路径迁至新 module。
- `syntax-try`：`try` 包发布路径迁至新 module。
- `macro-expander`：测试/fixture 中硬编码的旧 contrib import path 对齐（若规范层有引用）。

## Impact

- **用户（宏使用方）**：宏主文件与 `cmd/macroexpand` 的 blank import 须改用新 module 路径；需 `go get github.com/arcane-craft/go-macro-contrib@<version>`。
- **go-macro 仓库**：删除 `contrib/`；`examples/go.mod`、`go.work`、文档、部分 `internal/*_test.go` 路径字符串。
- **go-macro-contrib 仓库**：新仓承载原 contrib 全部代码与 CI；`go.mod` 继续 `require` 已发布的 `go-macro` 核心版本（开发期 `replace` 可选）。
- **发布**：contrib 与 core 可独立打 tag；需在迁移说明中注明最低兼容 `go-macro` 版本。
