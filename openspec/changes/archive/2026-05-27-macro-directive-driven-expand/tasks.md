## 1. Registry 与指令解析

- [x] 1.1 实现 per-function `//macro:` 解析（桩与 Expander），移除 `isPanicStub` 作为登记条件
- [x] 1.2 重写 `RegisterProvider`：syntax-id 冲突检测、桩↔Expander 关联、`(importPath, stubName)` 查找
- [x] 1.3 更新 `macro/registry_test.go` 与 `registry_more_test.go` 覆盖新契约与错误路径
- [x] 1.4 更新 `internal/expander/recognize.go` 使用 `(importPath, stubName)` 查找

## 2. cmd/macro expand 与 link 生成

- [x] 2.1 新增 `expand` 子命令：解析 module root、扫描 patterns 下 import 与 provider Expander 注解
- [x] 2.2 实现 `.gomacro/expand_link.go`（或 design 选定路径）生成与增量更新
- [x] 2.3 集成 `expandtool.Run` 完成 `*_macro_gen.go` 写回
- [x] 2.4 为 expand 子命令添加集成测试（临时 module fixture）

## 3. 脚手架与 expandtool

- [x] 3.1 更新 `init provider` 模板：per-function `//macro:`，删除 `register/` 生成
- [x] 3.2 调整 `expandtool`/`RegisterLinked` 适配新 `ProviderSyntaxID` 语义（按 Expander doc 取 id）
- [x] 3.3 更新 `internal/expander` 全部测试 fixture 为 per-function 注释

## 4. 文档与 examples

- [x] 4.1 更新 `docs/author-guide.md`（角色分工、契约、使用方 expand 一行命令）
- [x] 4.2 更新根 `README.md` 快速上手与命令节
- [x] 4.3 更新 `examples/`：`go:generate` 改 `cmd/macro expand`；移除旧展开入口
- [x] 4.4 在 author-guide 或 README 说明 `.gomacro/` gitignore 与 CI expand 建议

## 5. go-macro-contrib 协调（可并行 PR）

- [x] 5.1 为 inline/try 各桩与 Expander 添加 per-function `//macro:`
- [x] 5.2 删除 `register/` 包并更新 contrib 文档与测试
- [x] 5.3 验证 `go run .../cmd/macro@latest expand` 对 contrib 消费方示例可用

## 6. 收尾

- [x] 6.1 全仓 `GOWORK=off go test ./...` 与 examples 测试通过
- [x] 6.2 运行 `openspec validate macro-directive-driven-expand`（若 CLI 支持）并修复 spec 问题
