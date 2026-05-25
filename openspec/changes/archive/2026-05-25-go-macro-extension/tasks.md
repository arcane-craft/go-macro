> **验收切割（已定）**：§2–§4 + §4.6（`syntax-inline`）= **P0**；§5 = **P2**（`syntax-try`）。引擎 MUST NOT 为 Try 硬编码。

## 1. 项目初始化

- [x] 1.1 初始化 `go.mod`（`github.com/arcane-craft/go-macro`）与基础目录结构
- [x] 1.2 添加 README，说明宏工作流与 `go generate` / `go tool macro` 用法

## 2. macro-core 核心库

- [x] 2.1 实现 `Context`、`CallSiteKind`、`ExpandResult`（含 `Exprs`）、`Expander` 类型
- [x] 2.2 实现 `Context` 填充（Types、StubName、Site、**EnclosingFunc 必选**、TempIdent）
- [x] 2.3 实现宏注册表：扫描 `//macro:` 注释、绑定 syntax-id 与展开函数
- [x] 2.4 为注册表与 Context 编写单元测试
- [x] 2.5 实现轻薄辅助：`ErrorAt`、`MacroPos`（供 //line）；`macro/mactest` 纯 Expand 单测辅助

## 3. macro-expander 展开引擎

- [x] 3.1 实现包扫描与 `go/parser` 解析管线
- [x] 3.2 实现通用宏识别：macro 主文件扫描、Ident/SelectorExpr、`go/types` 包级桩判定、排除方法与 shadow
- [x] 3.2a 实现 provider 注册表：仅从宏主文件包的 import 集合加载 provider，扫描 `//macro:` + 桩名
- [x] 3.2b 识别/分发测试矩阵（design 3e，含方法调用、shadow、别名）
- [x] 3.3 实现 `Expander` 分发与 `ExpandResult` splice（Stmts/Expr/Exprs 按 Site 矩阵；**无 syntax-try 特判**）
- [x] 3.4 实现展开错误报告（文件名、行号、原因）
- [x] 3.5 为调用识别与替换逻辑编写单元测试（可用假 Expander 或 `syntax-inline`）

## 4. macro-codegen 代码生成

- [x] 4.1 实现 `cmd/macro`：`expand`（仅本模块）与 `init provider`（**最小单桩骨架**）
- [x] 4.1a 官方宏库（`inline`/`try`）：`cmd/macro` 不默认注册；`expander` 官方目录 + 宏主文件 import 激活；spec/design 已同步（Decision 8）
- [x] 4.2 实现方案 C 写回 + `//line` 指向宏主文件 + `go/format`
- [x] 4.2a 实现 `ComplementMacroConstraint`（读主文件 constraint，生成 `!macro` 侧头；不改主文件）
- [x] 4.2b 主文件缺 `macro` / 仅 `ignore` 时报错与文档提示
- [x] 4.2c 文档：主文件自行合并 `macro` 与平台 tag；gopls `buildFlags: ["-tags=macro"]`
- [x] 4.3 支持 `//go:generate go tool macro expand` 集成
- [x] 4.4 验证幂等展开（重复执行产物一致）
- [x] 4.5 为 CLI 添加集成测试（testdata 包）

## 4.6 syntax-inline 表达式宏（P0）

- [x] 4.6.1 创建 `inline` 包：单桩 `Inline` + `InlineExpand`（`SiteExpr` → `Expr`）
- [x] 4.6.2 `InlineExpand` 的 `mactest` 单测 + 非法 Site 报错
- [ ] 4.6.3 可选：极小 macro 主文件 + expand golden，验证 P0 端到端（不依赖 Try）

## 5. syntax-try 参考宏（P2）

- [x] 5.1 创建 `try` 包：语法桩族 `Try0`/`Try`/`Try2`/`Try3`（可选 `Try4`）+ 共享 `TryExpand`；注册多桩名到 `syntax-try`
- [x] 5.2 实现 `TryExpand`：返回 `ExpandResult.Stmts`；`return Try` 用完整 if+return 序列（**MUST NOT** 用 `Exprs` 简化）
- [x] 5.2a 赋值 / return / `Try0` 语句三种 Site；非法 Site 报错
- [x] 5.2b `TryExpand` 纯 `mactest` 单测 + 非法场景；可选 golden 全链路
- [x] 5.3 添加 `ReadFile` 示例与 `//go:generate` 指令
- [x] 5.4 添加 golden 测试：展开前后源码对比
- [x] 5.5 验证 `go test ./...` 在 generate 后全部通过

## 6. 收尾与文档

- [x] 6.1 编写 `docs/author-guide.md`（框架契约 vs 官方宏库 `syntax-inline` / `syntax-try`（import 启用）；mactest、init provider、//line）
- [x] 6.2 README 快速上手 + 链到作者指南；**对外库 MUST 提交 gen**
- [x] 6.3 文档：带宏库作 dependency 的发布 checklist（expand → commit gen → tag）
- [x] 6.4 Try 参考：桩族表与 error 在返回列表最后（作者指南附录，非框架 MUST）
