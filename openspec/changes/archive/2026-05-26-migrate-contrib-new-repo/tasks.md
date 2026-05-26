## 1. 新建 go-macro-contrib 仓库

- [x] 1.1 在本地 `../go-macro-contrib` 初始化仓库（并创建远端 `go-macro-contrib`），将当前 `contrib/` 目录内容移至该仓根（`inline/`、`try/`、`register/`、`go.mod`、`go.sum`、测试）
- [x] 1.2 将新仓 `go.mod` module 改为 `github.com/arcane-craft/go-macro-contrib`；`require` 已发布的 `go-macro`；本地联调 `replace github.com/arcane-craft/go-macro => ../go-macro`
- [x] 1.3 批量替换新仓内 import 与 `expandtool.Register` 键为 `github.com/arcane-craft/go-macro-contrib/...`
- [x] 1.4 在新仓根执行 `go test ./...` 并修复至通过
- [x] 1.5 添加新仓 README（module 路径、与 `go-macro` 同级目录 `../go-macro` 的 `replace` 联调说明、最低 `go-macro` 版本）

## 2. 更新 go-macro 仓库

- [x] 2.1 删除 `go-macro` 内 `contrib/` 目录
- [x] 2.2 更新 `go.work`：移除 `./contrib`，保留 `.` 与 `./examples`
- [x] 2.3 更新 `examples/go.mod`：`require github.com/arcane-craft/go-macro-contrib`；`replace` 为 `../go-macro-contrib`（移除原 `../contrib` replace）
- [x] 2.4 更新 `examples/cmd/macroexpand`、`examples/readfile` 及已生成 `readfile_macro_gen.go` 中的 import 路径
- [x] 2.5 更新根 module 测试 fixture（如 `internal/expander/load_unit_test.go`、`internal/codegen/*_test.go`）中的硬编码路径
- [x] 2.6 更新 `README.md`、`docs/author-guide.md` 中的模块路径与官方宏库说明
- [x] 2.7 在 `go-macro` 仓库根执行 `go test ./...`（workspace）并修复至通过

## 3. 规范与发布

- [x] 3.1 归档本变更：将 delta 合并至 `openspec/specs/{macro-contrib,macro-repo-layout,macro-codegen,syntax-inline,syntax-try,macro-expander}/spec.md`
- [x] 3.2 在两仓 CHANGELOG/README 添加 **BREAKING** 迁移表（旧路径 → 新路径、`go get` 示例）
- [x] 3.3 为 `go-macro-contrib` 打首个独立 tag；更新 `examples/go.mod` 引用版本；确认与 `go-macro` 核心 tag 兼容矩阵
