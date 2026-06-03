# Design: decl-macro-and-call-rename

## Context

今日数据流（仅 Call 宏）：

```text
ScanProviderFiles → 每包单一 syntax-id，函数桩 + Expander
RecognizeMacroCalls → *ast.CallExpr
ExpandFile → Expander(Context, call) → ExpandResult.Target + splice
WriteGenFile → 仅 *ast.FuncDecl
```

问题：

1. 无法表达 struct 嵌入标记（derive、wire tag 等）。
2. `Context` / `ExpandResult` / `Expander` 命名无法与将引入的 Decl API 区分。
3. gen 文件不含类型定义，与「类型与方法均在 macro 文件、整段进 gen」的目标不符。

已确认约束见 `proposal.md` Boundaries 表。

## Goals / Non-Goals

**Goals:**

- **BREAKING** 重命名 Call API：`CallContext`、`CallExpandResult`、`CallExpander`
- 新增 Decl API：`DeclContext`、`DeclSite`、`DeclExpandResult`、`DeclExpander`
- 嵌入匿名 marker 识别；每 embed 一次 DeclExpander；多 Marker 按字段顺序
- Decl 成功：全量 `Fields` + 全量 `Methods` 替换 Target 声明
- 注册表支持 **多 syntax-id**；Call/Decl Expander **分别 link**
- Expand 顺序：**Decl → Call**
- `WriteGenFile` 输出 **TypeSpec + FuncDecl**；Decl 方法带 `//line`
- 试点：`syntax-derive-stringer`、`syntax-wire-json`（`go-macro-contrib`）

**Non-Goals:**

- 具名嵌入 `m Marker`
- 读取 marker 类型 struct 字段作为配置
- 包级 `const`/`var`、其它类型、其它函数、mock/test 独立文件
- Metadata-only 独立 syntax、Test/Mock 声明宏
- expand 后全文件二次 typecheck（可选 follow-up）
- 字符串字面量作泛型类型实参（非 Go 语法）

## Decisions

### D1：Call API 重命名（E1，不兼容）

| 旧名 | 新名 |
|------|------|
| `Context` | `CallContext` |
| `ExpandResult` | `CallExpandResult` |
| `Expander` | `CallExpander` |
| `NewContext` | `NewCallContext` |
| `ValidateExpandResult` | `ValidateCallExpandResult` |
| `LegalSpliceTargetsForCall` | 保留或别名 `CallLegalSpliceTargets`（实现可选别名过渡一周，spec 要求新名） |

保留：`SpliceTarget`、`CallSiteKind`、`Site()`。

**理由**：与 `Decl*` 对称；项目未发布，一次性替换成本低。

### D2：Decl API 形状

```go
type DeclExpander func(ctx DeclContext, site DeclSite) (DeclExpandResult, error)

type DeclExpandResult struct {
    Fields  []ast.Field   // 全量 struct 字段列表（必填，成功时）
    Methods []*ast.FuncDecl // 全量 Target 方法（必填，成功时）；receiver 必须为 Target
}

type DeclSite struct {
    Target     *ast.TypeSpec
    TargetType types.Type
    EmbedIndex int
    EmbedField *ast.Field
    MarkerImportPath string
    MarkerTypeName   string
    MarkerTypeArgs   []types.Type // 泛型实例化实参，可为空
    MacroTag     MacroTag       // 解析自 EmbedField.Tag 的 macro 键
}
```

**成功语义**：`Fields` 与 `Methods` **均 MUST 非 nil 且表达完整目标形态**（含删桩、改 tag、含未改方法的原样带回）。禁止零值 result 表示成功。

**失败语义**：仅 `error != nil`；不写 gen。

**理由**：引擎替换简单（整表 swap）；作者责任清晰；Fields/Methods 一视同仁。

**备选**：patch API（按 index 改 tag）→ 作者易漏方法，与「全量 Methods」决议冲突，不采用。

### D3：Marker 语法模板

| 形态 | 合法 Go 示例 |
|------|----------------|
| 无参 | `DeriveStringer` |
| 必选类型实参 | `JSONMarker[Item]`（`Item` 为类型名） |
| 可选 KV | `` `macro:"omitempty=Role"` `` 写在嵌入字段上 |
| 必选 + 可选 | `Wire[T] `macro:"..."`` |

`//macro: <syntax-id>` MUST 在 provider 的 **marker 类型** `type X struct{}` doc 上。

桩类型内字段 **仅 godoc**，引擎 **MUST NOT** 读取。

### D4：宏识别规则

一次嵌入算作宏，当且仅当：

1. 字段为 **匿名嵌入**；
2. 嵌入类型解析为 **已 import** provider 包中的类型；
3. 该类型在注册表中登记为 **marker 类型**，且对应 syntax-id 已 **link DeclExpander**；
4. 泛型实例 `Marker[T]` 的 base name + type args 与注册条目匹配（由 `go/types` 判定）。

用户自定义同名空 struct **不**满足 (2)(3) 时不触发。

### D5：注册表与多 syntax-id

从「每包单一 syntax-id」改为：

```text
syntaxID → {
    callStubs:  map[stubFuncName]struct{}
    declMarkers: map[markerTypeName]struct{}  // 含泛型 base name
    callExpander: CallExpander?   // link 提供
    declExpander: DeclExpander?   // link 提供
}
```

同一 provider 包可有多个 syntax-id（如 `derive-stringer` + `wire-json`）。

`ScanProviderFiles` 扩展：

- 函数 doc `//macro:` → call stub + 绑定 CallExpander（若该函数签名符合 `CallExpander`）
- 类型 doc `//macro:` → decl marker + 绑定 DeclExpander（若存在符合签名的函数，命名约定 `XxxDeclExpand` 或同 syntax-id 的 `func ... DeclExpand`）

**link 表**（expand_runner 生成）：

```go
expandtool.RegisterCall("path", pkg.TryExpand)           // per syntax or per path
expandtool.RegisterDecl("path", pkg.DeriveStringerExpand)
```

具体键：`syntax-id` 字符串优于 import path（同一包多 syntax 必须按 syntax-id link）。**决策**：`RegisterCall(syntaxID, expand)` / `RegisterDecl(syntaxID, expand)`，import path 在发现阶段解析。

### D6：Expand 管线顺序

```text
expandOnePackage(file):
  parse + typecheck
  ExpandDeclMacros(file)   // 改 TypeSpec + file.Decls 中的 Target 方法
  ExpandCallMacros(file)   // 现有 Call splice
  WriteGenFile(file)       // 写出 types + funcs
```

同一文件内 Call 宏可出现在 Decl 展开后的方法体中；Decl 先跑保证 struct 形态已定。

### D7：ApplyDeclExpandResult

对每个成功的 `DeclExpandResult`：

1. 将 `Target` 的 `StructType.Fields.List` 设为 `result.Fields`
2. 从 `file.Decls` 移除所有 receiver 为 `Target` 的 `*ast.FuncDecl`
3. 在 `Target` 的 `TypeSpec` 之后插入 `result.Methods`（顺序保持作者返回顺序）
4. 不在此阶段写 gen 文件

**多 embed 顺序**：按 `EmbedIndex` 升序（字段声明顺序）依次调用；每次 apply 后更新 AST，下一 embed 的 index 可能需重新扫描（实现：收集 sites 时记录 type name + marker，每轮 apply 后按字段名/marker 类型重新定位，或从后往前删 embed 保 index 稳定——**采用按声明顺序正向扫描，每次 apply 后重新 collect 剩余 sites**）。

### D8：WriteGenFile 扩展

当前仅遍历 `*ast.FuncDecl`。改为：

1. 遍历 `file.Decls`
2. `*ast.GenDecl` with `token.TYPE` → 打印展开后的 `TypeSpec`（含 struct）
3. `*ast.FuncDecl` → 现有逻辑 + `//line`

macro 主文件在 `!macro` 构建不参与编译；生产构建仅 gen 含类型+方法+函数。

### D9：Decl `//line`

Decl 生成的每个方法体语句前插入 `//line <main>:<line>`，行号取自 **嵌入 marker 字段**类型名或 tag 的位置（与 Call 宏 `MacroPos()` 一致由 `DeclContext.MacroPos()` 提供）。

### D10：冲突与 Contract

框架 **不** 内置「已有 `String()` 必报错」。`derive-stringer` syntax spec 可规定报错；`wire-json` 仅改 tag。

纯 Contract：`DeclExpander` 返回 `error`；或返回全量 Fields（删桩）+ 全量 Methods（不变）。

### D11：试点 syntax 归属

- `syntax-derive-stringer`、`syntax-wire-json` 的**权威 OpenSpec** 位于 `go-macro-contrib/openspec/specs/`；
- 实现位于 `derivestringer`、`wirejson` 包；`go-macro` change delta 对该两能力为 REMOVED（归档时不写入 go-macro 主 spec）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 全量 Methods 作者易漏 | `DeclContext.TargetMethods()` 返回当前切片副本；文档 + mactest 模板 |
| 多 embed 正向 apply 下标漂移 | 每轮 apply 后重新扫描 sites，或首轮收集 stable key（type名+marker语法） |
| WriteGenFile 打类型导致 gen 变大 | 与 macro 双文件模型一致；主文件仅 macro 构建 |
| 多 syntax-id 破坏现有 ScanProviderFiles | 全量改注册表 + 测试 |
| contrib BREAKING | Call API 重命名 + 新增 Decl 语法包；spec 已同步 |

## Migration Plan

1. 合并本 change spec + 实现 tasks
2. 本仓：重命名 Call API → 实现 Decl 管线 → 扩展 codegen → 更新 examples
3. contrib：`TryExpand`/`InlineExpand` 使用 `CallContext`；`derivestringer`/`wirejson` 与 OpenSpec 已落地
4. 无运行时兼容层（未发布）

## Open Questions

（无阻塞项；以下实现期细化）

- `RegisterCall`/`RegisterDecl` 是否保留 per-import-path 便捷重载用于测试
- `ExpandDecl` mactest 是否一次测单 embed 或支持多 marker 顺序集成测
