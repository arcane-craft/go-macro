## 1. Call API 重命名（BREAKING）

- [x] 1.1 将 `Context`/`ExpandResult`/`Expander` 重命名为 `CallContext`/`CallExpandResult`/`CallExpander`
- [x] 1.2 重命名 `NewContext`、`ValidateExpandResult`、`ValidateExpandResultForCall` 等辅助函数
- [x] 1.3 更新 `internal/expander`、`macro/expandtool`、`cmd/macro` 模板与全仓引用
- [x] 1.4 更新 `macro/mactest` 为 `ExpandCall`/`ValidateCall`（或保留别名并标记废弃——spec 要求新名）
- [x] 1.5 更新本仓测试与 `examples/`；在 tasks 备注中跟踪 `go-macro-contrib` `TryExpand`/`InlineExpand` 迁移（contrib 需在独立 PR 将 `TryExpand(ctx CallContext, ...)` 等同步）

## 2. Decl API 与注册表

- [x] 2.1 新增 `DeclContext`、`DeclSite`、`DeclExpandResult`、`DeclExpander`、`MacroTag`
- [x] 2.2 实现 `NewDeclContext`、`ValidateDeclExpandResult`、`DeclContext.TargetMethods()`
- [x] 2.3 扩展 `ScanProviderFiles`：扫描类型 doc `//macro:` 登记 marker；支持**每包多 syntax-id**
- [x] 2.4 实现 `RegisterCall`/`RegisterDecl`（按 syntax-id）及 `expandtool` 双注册表
- [x] 2.5 更新 `cmd/macro expand` link 生成：为每个 syntax-id 分别 link Call/Decl Expander

## 3. Decl 展开引擎

- [x] 3.1 实现 `RecognizeDeclSites`：匿名嵌入 + 已注册 marker + import/link 校验
- [x] 3.2 实现 `ExpandDeclMacros`：按字段顺序、每 embed 一次 `DeclExpander`
- [x] 3.3 实现 `ApplyDeclExpandResult`：全量替换 Fields、替换 Target 方法集
- [x] 3.4 在 `expandOnePackage` 中接入顺序：**先 Decl 后 Call**
- [x] 3.5 为 Decl 管线添加单元测试与集成测试

## 4. Codegen 扩展

- [x] 4.1 扩展 `WriteGenFile` 输出 `TypeSpec`（struct）及 Decl 生成的方法
- [x] 4.2 为 Decl 方法体生成 `//line`（`MacroPos` 取自嵌入字段）
- [x] 4.3 更新 `macro-codegen` 相关测试；验证 `go build`（无 macro tag）使用 gen 中类型+方法

## 5. 试点 syntax

- [x] 5.1 实现 `derive-stringer`：`DeriveStringer` marker + `DeriveStringerExpand`（生成 `String()`、删桩、冲突报错）
- [x] 5.2 实现 `wire-json`：`WireJSON` marker + `WireJSONExpand`（`macro` tag 驱动 json tag、全量 Methods）
- [x] 5.3 在 contrib 添加 `derivestringer` / `wirejson` 与单测；`go test` 验证（端到端 expand 示例可后续补）

## 6. mactest 与文档

- [x] 6.1 实现 `mactest.ExpandDecl` 与 `ValidateDecl`
- [x] 6.2 更新 `docs/author-guide.md`：Call/Decl 双轨、Marker 模板、全量 Fields/Methods
- [x] 6.3 更新 `README.md` breaking 说明（Call 重命名 + Decl 简介）

## 7. 验收

- [x] 7.1 `go test ./...` 通过
- [x] 7.2 `openspec validate decl-macro-and-call-rename` 通过
- [x] 7.3 归档：合并框架类 delta 至 `go-macro/openspec/specs/`（`macro-core`、`macro-expander`、`macro-codegen`、`decl-macro`、`author-guide`、`macro-repo-layout`）；**勿**将 `syntax-derive-stringer`/`syntax-wire-json` 并入 go-macro（权威 spec 在 contrib，已实现）
- [x] 7.4 contrib OpenSpec：`syntax-derive-stringer`、`syntax-wire-json` 主 spec；更新 `macro-contrib`；`syntax-try`/`syntax-inline` Call API 命名
