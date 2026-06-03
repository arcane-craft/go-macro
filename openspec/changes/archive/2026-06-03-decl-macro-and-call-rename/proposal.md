# Proposal: decl-macro-and-call-rename

## Why

当前 go-macro 仅支持**过程宏**（函数体内 `Stub(...)` 调用点展开），无法表达「在类型声明上贴标记、改写 struct 与方法」的**声明宏**需求（Contract、Wire 等）。同时，过程宏 API 使用泛称 `Context` / `ExpandResult` / `Expander`，与将引入的声明宏 API 无法区分。项目尚未发布正式版，适合一次性完成 **Call API 重命名** 与 **Decl 宏能力** 的规格与实现，避免后续破坏性演进成本。

## What Changes

- **BREAKING**：过程宏 API 重命名为 `CallContext`、`CallExpandResult`、`CallExpander`（及 `NewCallContext`、`ValidateCallExpandResult` 等）；`SpliceTarget` 保留现名，仅用于 Call 宏。
- **新增**嵌入**声明宏**（Decl macro）：通过 struct **匿名嵌入**已注册的 marker 类型触发；`//macro:` 标在 marker **类型**定义上。
- **新增** `DeclContext`、`DeclSite`、`DeclExpandResult`、`DeclExpander`；成功时 MUST 返回 **全量** `Fields` 与 **全量** `Methods`（该 Target 类型的完整声明形态）。
- **新增**声明宏展开管线：同一宏主文件内 **先 Decl 后 Call**；每 embed **一次** `DeclExpander` 调用；多 Marker 按 **字段声明顺序** 展开。
- **修改**注册表：同一 provider 包 **允许多个** `syntax-id`；Call 与 Decl Expander **分别 link**。
- **修改** `WriteGenFile`：生成侧 MUST 写出展开后的 **类型定义**（`TypeSpec`）与 **函数**（含 Decl 生成的方法与 Call 展开后的函数体）。
- **修改** `//line`：Decl 生成的方法与 Call 生成代码同级要求，指向 macro 主文件嵌入处。
- **试点 syntax**（`go-macro-contrib`）：`derive-stringer`、`wire-json`（`derivestringer`、`wirejson` 包）。
- **范围外**：包级 `const`/`var`、其它类型/函数、独立 mock/test 文件、Metadata-only 独立能力、Test/Mock 声明宏；桩 struct 字段仅 godoc 提示、**永不读取**。

## Capabilities

### New Capabilities

- `decl-macro`：嵌入声明宏的识别、Context/Result、展开顺序、作用域 MUST NOT、注册与 link

（`syntax-derive-stringer`、`syntax-wire-json` 的**权威 spec** 在 `go-macro-contrib`；本 change delta 中对该两能力的条目为 REMOVED，归档时不并入 `go-macro/openspec/specs/`。）

### Modified Capabilities

- `macro-core`：Call API 重命名；Decl API 新增；多 syntax-id 注册；expandtool 双 Expander 注册
- `macro-repo-layout`：官方库列表含 `derivestringer`、`wirejson`；syntax OpenSpec 归属 contrib
- `macro-expander`：Decl 展开引擎；Expand 顺序；Call API 名称更新
- `macro-codegen`：gen 文件写出类型定义；Decl 方法 `//line`
- `author-guide`：过程宏/声明宏双轨文档；Marker 模板；Call/Decl Expander 签名

## Impact

| 区域 | 影响 |
|------|------|
| `macro/*.go` | **BREAKING** 重命名 + Decl 新类型 |
| `internal/expander/` | Decl 扫描/apply；Expand 顺序 |
| `internal/codegen/writeback.go` | 输出 `TypeSpec` + `FuncDecl` |
| `macro/registry.go` | 多 syntax-id；类型桩扫描 |
| `macro/expandtool` | `RegisterCall` / `RegisterDecl`（或等价） |
| `macro/mactest` | `ExpandCall` / `ExpandDecl` |
| `cmd/macro` | link 生成、脚手架 |
| `docs/author-guide.md`、`README.md` | 文档更新 |
| `go-macro-contrib` | **BREAKING** Call API；新增 `derivestringer`/`wirejson`；OpenSpec 增 `syntax-derive-stringer`、`syntax-wire-json` |
| `openspec/specs/*` | 归档时合并框架类 delta；**不**合并 syntax-derive-stringer / syntax-wire-json 至 go-macro |

## Boundaries（已定）

| 议题 | 决策 |
|------|------|
| Call 命名 | `CallContext` / `CallExpandResult` / `CallExpander` |
| `SpliceTarget` | 保留现名 |
| `CallSiteKind` / `Site()` | 保留 |
| 多 syntax-id | 每 provider 包允许多个 |
| link | Call 与 Decl Expander 分别 link |
| 嵌入 | 仅匿名嵌入 |
| 可选参数 | 仅 `` `macro:"k=v"` `` tag |
| `//macro:` 位置 | marker 类型 |
| `Marker[T]` 语义 | 各 syntax-id 自定 |
| 宏识别 | 用户可写形似类型；仅符合注册规则的算宏 |
| 多 Marker 顺序 | 字段声明顺序 |
| Decl 调用粒度 | 一 embed 一次 Expander |
| Decl 成功结果 | Fields 与 Methods **均全量** |
| 冲突处理 | 由宏作者决定（报错或替换） |
| 纯 Contract | 允许仅 `error` 失败 |
| 首期场景 | Contract & Lint、Wire |
| 桩字段 | 永不读取 |
| 作用域 | MUST NOT 包级 const/var、其它类型/函数、外部测试文件 |
| gen 内容 | 类型 + 方法 + 函数均进 `*_macro_gen.go` |
| Expand 顺序 | 先 Decl 后 Call |
| `//line` | Decl 与 Call 同级 |
| 试点 | `derive-stringer` + `wire-json` |

## Success criteria

- `openspec validate decl-macro-and-call-rename` 通过
- spec delta 与 design 一致
- 实现阶段 `go test ./...` 通过（实现 tasks 完成后）
