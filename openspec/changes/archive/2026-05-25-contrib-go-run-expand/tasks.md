## 1. macro/expandtool 与 cmd/macroexpand

- [x] 1.1 实现 `macro/expandtool`（`Register`、`Registered`、`Run`、`Main`）
- [x] 1.2 为 `expandtool` 添加单测（默认 patterns、nil linked 使用 Registered、错误传播）
- [x] 1.3 实现 `examples/cmd/macroexpand/main.go`（blank import `contrib/register` + `expandtool.Main()`）
- [x] 1.4 根 `go.mod` 不 require contrib（macroexpand 与 contrib link 仅在 examples module）

## 2. contrib 子 module

- [x] 2.1 创建 `contrib/go.mod`，迁入 `inline/`、`try/`（含测试）
- [x] 2.2 实现 `contrib/register/register.go`（init 中 `expandtool.Register` 官方库）
- [x] 2.3 删除根目录 `inline/`、`try/`
- [x] 2.4 contrib 目录 `go test ./...` 通过

## 3. expander API 去耦

- [x] 3.1 删除 `expander/official_providers.go` 及测试
- [x] 3.2 `ExpandPackages(patterns, linked map[string]macro.Expander)`
- [x] 3.3 删除 `expander.Provider`；实现 import ∩ linked 过滤
- [x] 3.4 修复 `load.go` active append bug
- [x] 3.5 通用化 `internal/codegen/imports.go`
- [x] 3.6 更新 expander 测试（显式 linked 或 expandtool.Registered）

## 4. CLI 与脚手架

- [x] 4.1 删除 `cmd/macro` 的 `expand` 子命令
- [x] 4.2 `init provider` 生成 `register/register.go`，**不**生成 `tools/macroexpand`
- [x] 4.3 更新 `cmd/macro` usage 与测试

## 5. 示例与文档

- [x] 5.1 更新 `examples/readfile`：contrib import + `go run .../examples/cmd/macroexpand` generate；regenerate golden
- [x] 5.2 README：一行 generate，明确**无需** tools/macroexpand
- [x] 5.3 author-guide：宏作者 vs 使用方职责；第三方 register 附录
- [x] 5.4 **不要**添加根 module `tools/macroexpand`

## 6. 验证

- [x] 6.1 核心 `go test ./...`（根测试不 import contrib；`GOWORK=off` 时根 go.mod 无 contrib）
- [x] 6.2 contrib `go test ./...`
- [x] 6.3 仓库无 `go tool macro expand`、无 canonical `tools/macroexpand` 要求

## 7. 仓库结构（examples module）

- [x] 7.1 `examples/go.mod` 独立 module；`examples/cmd/macroexpand`；删除根 `cmd/macroexpand`
- [x] 7.2 根 `go.work`（use 根、contrib、examples）
- [x] 7.3 依赖 contrib 的集成测试迁至 `examples`；根 expander 测试不 import contrib
- [x] 7.4 specs：`macro-repo-layout` 及关联 capability 路径更新
