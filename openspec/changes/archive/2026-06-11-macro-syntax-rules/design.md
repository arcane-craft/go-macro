## Context

go-macro 已有稳定 expand 引擎（识别、SpliceTarget 贴回、codegen）与 **`macro/quote`**（`#` 模板写 AST）。宏作者仍大量手写 `go/ast`、Inspect enclosing 语句，且 **CallExpander / DeclExpander**、**CallExpandResult**、**CallContext** 与 **DeclContext** 双轨 API 增加心智负担。探索结论：以 **Scheme syntax-rules / syntax-case** 为参照，用 **`Syntax` 统一读写**，**Rule 表为默认 Expander**，**Match 时确定 SplicePlan**（无运行时 InferTarget），**Context 极简**，**外层函数签名** 通过通用 Types API 获取（不暴露 FuncDecl AST，不 Try 特化）。

## Goals / Non-Goals

**Goals:**

- `Syntax` + Match（`$`）+ Quote（`#`）对称 API
- `SyntaxRules`（pattern + template）与 `SyntaxCase`（+ fender + transform）
- 统一 `Expander(ctx Context, site Syntax) (Syntax, error)`
- `Context` 仅 `FileSet`、`Types`、`TempIdent`
- `site` 含 `MacroPos`；Match 隐含 anchor（同句多宏、展开顺序对作者透明）
- **`MatchedSpan` + `SplicePlan`**：pattern match 划定语义边界并在 Match 时产出 `[]SpliceStep`；`ValidateSplice(out, meta)` + `Apply(file, meta, out)`；可选 `Clause.Plan` override
- **`MatchMeta` 经 `site` 内部槽传递**（D15）：公开 `Expander` 只返回 `Syntax`；引擎从 `site` 读 meta
- `EnclosingSignature` / `EnclosingResults` / `ZeroSyntax` 通用 API
- 合并 `macro/quote` 入 `macro.Quote`；短期 adapter 后 **BREAKING** 删旧 API

**Non-Goals:**

- 编译期 macro / 完全 Scheme hygiene
- Try 专用 built-in（`ErrorReturn`、`ZerosFor` 框架注入）
- `Clause.Literals` 字段（裸 ident = literal，见 macro-pattern）
- 首版强制 contrib 全量迁移
- `@kind{ }` Quote 包装（已由 `To*` 取代）

## Decisions

### D1：统一 `Syntax` 为读写与贴回载荷

**选择**：`Syntax` 提供 `Match`、`To*`、`Underlying()`、`MacroPos()`。

**理由**：一种类型贯穿 Match → Quote → Apply；与 Scheme stx 心智一致。

**备选**：保留 `CallExpandResult` — 拒绝，与 syntax-rules 模型冲突。

### D2：Pattern / Template 符号分离

**选择**：Match 用 `$`（`$name`、`$_`、`$field ...`）；Quote 用 `#`（`#name`、`#field ...`）。裸标识符（非 `$` 前缀）在 pattern 中为 **literal**。

**理由**：读写对称且无二义；**不需要 `Clause.Literals` 字段**（方案 A）。

### D3：Ellipsis

**选择**：`$field ...`（空格 + 三点），Quote 侧 `#field ...`。

**理由**：贴近 Scheme；与讨论定稿一致。

### D4：Quote 无 `@kind`

**选择**：`Quote(template, binds map[string]Syntax) Syntax`；产出形状由 **`ToExpr` / `ToStmts` / …** 校验。

**理由**：减少重复；与 Q4 决策一致。

### D5：统一 Expander，不区分 Call / Decl

**选择**：单一 `Expander`；`site` 构造覆盖 assign/return/struct type 等锚点。

**理由**：Derive 与 Try 同用 `SyntaxCase`；注册仍用 `//macro:`。

### D6：`Context` 三字段

**选择**：`FileSet`、`Types`、`TempIdent` only。

**理由**：`File`、`EnclosingFunc`、`MacroPos` 等与 site 或引擎重复；见 D7/D8。

### D7：`site` 与 anchor

**选择**：`ResolveSite(file, anchor)` → `Syntax`；anchor 类型因 Call/Decl 而异（见 **D18**）；内部含本轮 anchor、enclosing 根与 `fileRef`；**anchor 不导出**。

**Match（Call）**：pattern 中 stub Call **必须 unify** 本轮 anchor `*ast.CallExpr`；每轮 expand 重算 site。

**Match（Decl）**：anchor 为 embed `*ast.Field`；`DeclPattern` 以 enclosing `*ast.TypeSpec` 为 match 根；marker embed 经无序约束集 match，**不**走 Call unify 规则。

**理由**：同句多宏无需作者关心顺序；整句 stmt pattern 每轮只 match 待展开宏；Decl 与 Call 共用 `site Syntax` 类型但 anchor 语义分轨。

### D8：外层函数签名 — Types + 内部 scope

**选择**：`EnclosingSignature(ctx, site)` / `EnclosingResults` / `ZeroSyntax`；**internal** 用 `*ast.File` + Pos 定位 `*types.Func`，**不**在公开 API 返回 `*ast.FuncDecl`。

**理由**：Try error 分支 `return ...` 需 **完整 Results**（非 assign lhs  alone）；provider 自 compose；框架不 Try 特化。

### D9：MatchedSpan 与 SplicePlan（方案 C）

**选择**：`site.Match(pattern)` MUST 记录 **`MatchedSpan`**（语义边界）与 **`Plan []SpliceStep`**（贴回计划）。`Plan` MUST 在 Match 时完全确定；**无** normative `InferTarget`；**不**区分 Call/Decl。`SpliceStep` 首版两类原语：`ReplaceInContainer`（BlockStmts / AssignRhs / ReturnResults / GenDeclSpecs 等）、`InsertAfterInFileDecls`（Derive 追加新 methods）。`MatchRoot`（`Stmt` | `Call` | `Decl`）MAY 供错误信息；**MUST NOT** 作为贴回边界或 Plan 的权威来源。若 pattern 静态分析可推导多种 Plan → 注册 fatal；`Clause.Plan` MAY override（adapter / 极端 Transform）。

**理由**：贴回第一性是 **父槽位操作**，不是 Call/Decl 种类；与 syntax-rules 一致——pattern 决定 match 哪段 **以及** 如何替换；Try stmt 级与 Derive type 级同一 `[]SpliceStep` 执行模型。

**示例**（摘要）：

| pattern 示例 | MatchedSpan | Plan（摘要） | out 示例 | Apply 效果 |
|--------------|-------------|--------------|----------|------------|
| `$lhs ... := Try($inner)` | AssignStmt | BlockStmts 1→N | 3 Stmts | 1 stmt → 3 stmts |
| `Try($inner)` | anchor Call | AssignRhs 1→1 | 1 Expr | RHS call → expr |
| `type $item struct { ... }` | TypeSpec | GenDeclSpecs + InsertAfter | TypeSpec' + FuncDecls | 换 type + 插 methods |

### D10：`SyntaxCase` clause 行为

**选择**：pattern **解析**失败 → fatal（加载时）；**match**失败 → 下一 clause；fender 失败 → 下一 clause；全失败 → `no matching syntax rule`。

**理由**：Q8:B。

### D11：Apply 执行 SplicePlan

**选择**：`Apply(file, meta, out)` 按 `meta.Plan` 顺序执行；**仅**影响 Plan 所描述的槽位；**未** match 的 AST（如 pattern 未覆盖的 methods）**MUST NOT** 被修改或删除。**无** Decl 全量 merge/patch。

**out 节点数 MAY 大于 MatchedSpan**：Try 为 `ReplaceInContainer` 1→N；Derive 为 `GenDeclSpecs` 1→1 + `InsertAfterInFileDecls` 追加 **新生成** FuncDecls——同属 `Plan` 内步骤，**不是**作者级第二种 API。

**理由**：Derive 不必返回全量 Methods；Call/Decl 共享 `ValidateSplice` + `Apply(plan)`。

### D12：合并 `macro/quote`

**选择**：实现迁入 `macro/quote.go`（或同包）；`macro/quote` deprecated → 删除；绑定类型改为 `map[string]Syntax`。

**理由**：Q7:A。

### D13：Call 迁移 adapter（`TargetToPlan`）

**选择**：**仅 Call** 提供短期 adapter：`CallExpander` 包装为 `Expander`；`CallExpandResult` 经 **`TargetToPlan(file, call, result)`** 编译为 `MatchMeta.Plan`，载荷经 **`CallExpandResultToSyntax`** 转为 `out`；下一 major 删除。**不提供** `DeclExpander` / `DeclExpandResult` 语义等价 adapter（见 D14）。

**`TargetToPlan` 映射**（与 today `ApplyExpandResult` 行为对齐）：

| `SpliceTarget` | `Plan` | `MatchedSpan` | `out` |
|----------------|--------|---------------|-------|
| `SpliceReplaceAssignStmt` | `BlockStmts[i]` OneToMany | `*AssignStmt` | `ToStmts()` |
| `SpliceReplaceAssignRHS` | `AssignRhs[j]` OneToOne | `*CallExpr` | `ToExpr()` |
| `SpliceReplaceReturnStmt` | `BlockStmts[i]` OneToMany | `*ReturnStmt` | `ToStmts()` |
| `SpliceReplaceReturnResults` | `ReturnResults` **ReplaceAll** | `*ReturnStmt` | `ToExprs()` |
| `SpliceReplaceExprStmt` | `BlockStmts[i]` OneToMany | `*ExprStmt` | `ToStmts()` |
| `SpliceReplaceCallExpr` | 动态 `ExprSlot` OneToOne（同 `replaceCallExpr` parent 定位） | `*CallExpr` | `ToExpr()` |

**理由**：旧 Decl 的 `removeTargetMethods` + 全量 merge 与新 MatchedSpan 模型 **不可忠实编译**；Call 六种 Target 可定位 parent/index，适合 Plan 编译。

### D14：Decl 无 adapter，强制迁移

**选择**：Decl 宏 **MUST NOT** 提供 `DeclExpander` → `Expander` 或 `DeclExpandResult` → `SplicePlan` 的过渡 adapter。Derive 等 Decl provider **MUST** 在本 change 落地时改写为 **`SyntaxCase` + pattern 划定 TypeSpec + `out.ToDecls()`**；旧 API 与 `ApplyDeclExpandResult` 在引擎切换时 **直接删除**（与 Call adapter 清理可同 major，但无 Decl 过渡层）。

**理由（方案 C）**：旧 Decl 语义（全量 Fields/Methods、删光 receiver methods）与新模型根本冲突；虚假 adapter 会掩盖行为差异（未 match methods 被误删）。强制一次性迁移，spec 与实现更简单。

**Migration**：Derive `SyntaxCase` canonical 示例见下文「Derive SyntaxCase 示例」；contrib Decl provider 列为 **blocking**（非 optional）若仓库内建 Derive。

### Derive SyntaxCase 示例

本节展示 Decl 宏从旧 `DeclExpander` 迁移到新 `SyntaxCase` 的完整路径（对应 D14/D16/D18）。实现时 author-guide MUST 以此为准。

#### 使用方源码

```go
type Item struct {
    provider.Derive[fmt.Stringer]
    Name string `json:"name"`
}

func (Item) Foo() {} // 既有 method — 展开后 MUST 保留（与旧 API 不同，见下文「与旧 API 对比」）
```

#### Provider：marker 与 Expander 注册

```go
//macro: derive
type Derive[T any] struct{}

// DeriveExpander 注册为统一 Expander（expandtool.Register("derive", DeriveExpander)）
var DeriveExpander = macro.SyntaxCase(macro.Clause{
    Pattern: `type $item struct { Derive[$iface] $field ... }`,
    Transform: deriveTransform,
})
```

pattern 说明：

- `$item`：绑定 enclosing `*ast.TypeSpec`（类型名 `Item`）
- `Derive[$iface]`：按 **invoked name** 匹配 embed；`$iface` 绑定 `fmt.Stringer` 类型实参 `ast.Expr`
- `$field ...`：绑定 struct 内**全部具名字段**（0+），**不含** Derive embed；与 embed 在源码中的书写顺序无关
- `{ Derive[$iface] $field ... }` 与 `{ $field ... Derive[$iface] }` 语义等价

#### 引擎展开流水线（作者无需手写）

```
ResolveSite(file, embed *ast.Field)     anchor = Derive 的匿名 embed 字段
        │
        ▼
SyntaxCase.Expander(ctx, site)
        │
        ├─ site.Match(pattern)  →  meta 槽：
        │      Bindings = { item, iface, field... }
        │      MatchedSpan = *ast.TypeSpec "Item"     ← 贴回边界
        │      Plan = [ GenDeclSpecs OneToOne,        ← 替换 TypeSpec
        │               InsertAfterInFileDecls ]       ← 追加生成 methods
        │
        └─ deriveTransform(ctx, site, binds)  →  out Syntax
                │
                ▼
ValidateSplice(out, meta)  →  Apply(file, meta, out)
        │
        ├─ TypeSpec' 替换 MatchedSpan 中的 type Item struct { … Derive … }
        └─ String() method 插入 GenDecl 之后
        │
        ▼
文件中 func (Item) Foo() 未动（不在 MatchedSpan 内）
```

#### Transform 实现

```go
func deriveTransform(ctx macro.Context, site macro.Syntax, binds macro.Bindings) (macro.Syntax, error) {
    // ── 1. 类型名（旧 DeclSite.Target / TypeSpec）──
    itemSyn, ok := binds.Get("item")
    if !ok {
        return nil, fmt.Errorf("derive: missing $item")
    }
    itemTS := itemSyn.Underlying().(*ast.TypeSpec)
    typeName := itemTS.Name.Name

    // ── 2. 接口类型实参（旧 DeclSite.MarkerTypeArgs[0]）──
    ifaceSyn, ok := binds.Get("iface")
    if !ok {
        return nil, fmt.Errorf("derive: missing $iface")
    }
    ifaceExpr := ifaceSyn.Underlying().(ast.Expr)
    ifaceType := ctx.Types().TypeOf(ifaceExpr)
    if ifaceType == nil {
        return nil, fmt.Errorf("derive: cannot typecheck interface %s", typeName)
    }
    // ifaceType 用于判断需生成哪些 methods（如 fmt.Stringer → String）

    // ── 3. embed struct tag（旧 DeclSite.MacroTag）──
    // Decl anchor 即 embed *ast.Field（D18）；site.Underlying() 即该 field
    embedField := site.Underlying().(*ast.Field)
    tag := macro.ParseMacroTag(embedField.Tag)
    _ = tag // 例：tag["format"] 影响 String 实现

    // ── 4. 具名字段（旧：从 Target struct 手动 filter 匿名字段）──
    // $field ... 已排除 embed；每项 Underlying() 为 *ast.Field（含 Tag）
    fieldElems, _ := binds.Elems("field")

    // ── 5. Quote 产出 TypeSpec' + 生成 methods ──
    // 关键：TypeSpec' 只含 $field ... 绑定的字段，不含 Derive embed
    out, err := macro.Quote(`type #name struct {
    #field ...
}

func (#recv #name) String() string {
    return fmt.Sprintf("%v", #recv.Name)
}`, map[string]macro.Syntax{
        "name":  macro.LiteralIdent(typeName),
        "recv":  macro.LiteralIdent(strings.ToLower(typeName[:1])),
        "field": macro.FieldList(fieldElems), // 将 []Syntax(*ast.Field) 包装为 Quote #field ... 可注入的 Syntax
    })
    if err != nil {
        return nil, err
    }
    // out.ToDecls() → [TypeSpec', FuncDecl(String)]
    // ValidateSplice 检查：decls[0] 与 MatchedSpan 同名；decls[1:] 为新生成 FuncDecl
    return out, nil
}
```

辅助函数 `macro.FieldList` / `macro.LiteralIdent` 为 Quote 组合 helper（实现细节；等价于将 `binds.Elems("field")` 各 `*ast.Field` 注入 `#field ...`）。若 Transform 需更细控制，MAY 手写 `go/ast` 构造 TypeSpec' 再 `macro.WrapDecls(...)`——normative 路径仍须返回 `Syntax` 且 `ToDecls()` 通过 Validate。

#### 与旧 `DeclExpander` 对比

| 维度 | 旧 API | 新 SyntaxCase |
|------|--------|---------------|
| 签名 | `DeclExpander(ctx DeclContext, site DeclSite) (DeclExpandResult, error)` | `SyntaxCase` → `Expander(ctx, site) (Syntax, error)` |
| 类型实参 | `site.MarkerTypeArgs[0]` | `binds.Get("iface")` + `ctx.Types().TypeOf(expr)` |
| struct tag | `site.MacroTag` | `site.Underlying().(*ast.Field).Tag` + `ParseMacroTag` |
| 字段列表 | 手动遍历 `site.Target`，filter 匿名字段 | `binds.Elems("field")`，embed 已由 pattern 排除 |
| 既有 methods | **必须** `ctx.TargetMethods()` 全量复制进 result | **不必**复制；文件中未 match 的 `Foo()` 自动保留 |
| 贴回 | `ApplyDeclExpandResult` 替换全量 Fields + **删光** receiver methods 再插入 | `Apply(plan)` 仅替换 MatchedSpan TypeSpec + 插入**新生成** methods |
| 返回值 | `DeclExpandResult{Fields, Methods}` 全量 | `out.ToDecls()` = `[TypeSpec', 新生成 FuncDecl...]` |

#### 常见错误

1. **out 仍含 Derive embed** — `$field ...` 不含 embed，TypeSpec' 须只引用 `#field ...` 或等价构造，不得把 embed 抄回 struct。
2. **TypeSpec 改名** — `out.ToDecls()[0].Name` 须与 MatchedSpan 中 `$item` 同名，否则 `ValidateSplice` 失败。
3. **生成与既有 method 同名** — 若文件已有 `func (Item) String()` 且不在 MatchedSpan 内，Validate MUST 失败；Transform 应 skip 或报错，不得 silent 覆盖。
4. **误用 `TargetMethods()`** — 新 API 不恢复；默认 Derive 只**追加**生成 methods，不读不改既有 methods。

#### 带 MacroTag 的 embed 变体

```go
type Item struct {
    provider.Derive[fmt.Stringer] `macro:"format=compact"`
    Name string
}
```

pattern 不变（首版不支持 tag 字面量 match）。Transform 内：

```go
embedField := site.Underlying().(*ast.Field)
tag := macro.ParseMacroTag(embedField.Tag) // tag["format"] == "compact"
```

与 `$field ...` 项的 `field.Tag`（如 `` `json:"name"` ``）无关——后者来自 `binds.Elems("field")` 各项的 `Underlying().(*ast.Field).Tag`。

### D15：MatchMeta 经 `site` 内部槽传递（方案 E）

**选择**：`MatchMeta` 为**引擎内部**类型，**非**宏作者 API。公开 `Expander` 签名保持 `func(ctx Context, site Syntax) (Syntax, error)` 不变；`site.Match(pattern)` 仍只向作者返回 `(Bindings, error)`。`SyntaxRules` / `SyntaxCase` 与 Call adapter 在 expand 过程中将 `MatchMeta` 写入 **`site` 内部 meta 槽**；引擎在 `Expander` 返回后从 `site` 读取 `MatchMeta`，再执行 `ValidateSplice` / `Apply`。`out Syntax` **MUST NOT** 携带 `Plan`。

**写入时机**：

| 路径 | 写入方 | 时机 |
|------|--------|------|
| normative（SyntaxRules / SyntaxCase） | `site.Match` 成功 | 填入 `Bindings`、`MatchedSpan`、`Plan`；`Clause.Plan` override 在 match 成功后覆盖 `Plan` |
| Call adapter | `TargetToPlan` | 无 pattern match；adapter 在旧 `CallExpander` 返回后写入 meta 槽 |

**多 clause 生命周期**：runtime match 或 fender 失败 MUST 清空 meta 槽；仅最终胜出 clause 的 meta 保留至 `Expander` 返回。全部 clause 失败时 meta 槽 MUST 为空，引擎 MUST NOT 调用 `Apply`。

**读取**：`internal/expander`（或等价引擎包）提供 `MatchMetaFromSite(site Syntax) (MatchMeta, bool)`；**MUST NOT** export 给 provider。meta 槽为空时引擎 MUST 在 `ValidateSplice` 前返回 error，不得 silent 推断 `Plan`。

**裸 Expander**（见 **D19**）：MAY 实现手写 `Expander`，但 MUST 在返回 `out` 前对 `site` 调用 `Match(pattern)`（与 `SyntaxCase` 相同机制写入 meta 槽）；**不**暴露 `SetMatchMeta` 给 provider。

**理由**：`MatchMeta` 描述的是「这一展开点如何贴回」，与 `site` 语义一致；不改作者 API；避免 Go `Expander` func type 无法附带 `LastMeta()` 的问题。

**备选**：改 `Expander` 返回 `(Syntax, MatchMeta, error)` — 拒绝，增加作者心智负担且与 syntax-rules「pattern 决定贴回、作者只管 Bindings + out」冲突。

### D16：Decl embed 元数据经 Bindings + `Underlying()`（方案 F）

**选择**：Decl 宏 **不**恢复 `DeclContext` / `DeclSite` 或 `site.MacroTag()` 等专用 accessor。embed 元数据 **MUST** 经 pattern `Bindings` 与 **`Syntax.Underlying()`** 读取：

| 旧 `DeclSite` / `DeclContext` | 新 normative 路径 |
|-------------------------------|-------------------|
| `Target` / 类型名 | `binds.Get("item")`（pattern `$item`） |
| struct 字段列表 | `binds.Get("field")` ellipsis（每项 `Underlying()` 为 `*ast.Field`） |
| `MarkerTypeArgs` / 类型实参 | `binds.Get("iface")`（`Underlying()` 为 type `ast.Expr`）+ `ctx.Types().TypeOf(expr)` |
| `MacroTag` | 定位 embed 对应 `*ast.Field`（含 pattern literal `Derive[$iface]` 匹配项）→ `field.Tag` → **`macro.ParseMacroTag`** |
| `TargetMethods()` | **不**恢复；默认只生成新 methods。若须读未 match 既有 methods，MAY `site`/`MatchedSpan` 上级 `Underlying()` + `ast.Inspect`（escape，非默认路径） |
| `File()` / `Package()` | **不**恢复；复杂 AST 遍历走 `Underlying()` escape |

**`go/ast` 依据**：`Tag` 与 `Type` 平级，同属 `*ast.Field`；`$field ...` 绑定完整 field 节点（含 tag），非仅 `Type` 子树。首版 **不**在 pattern 语言增加 tag 字面量匹配（如 `` `macro:"k=v"` ``）；tag 内容在 Transform 内解析。

**理由**：与 D6 极简 Context、D15 不改 Expander 签名一致；避免 `site` accessor 与 Call/Decl 不对称膨胀；`ParseMacroTag` 已存在且 today 即用 `field.Tag`。

**备选**：`site.MacroTag()` / pattern tag literal — 拒绝，重复 `ast.Field` 已有信息。

### D17：Pattern 语言首版子集（normative）

**选择**：首版 pattern 为**封闭世界**语法，分三层：词法（`$` / `$_` / `$name ...` / 裸 ident = literal）→ **顶层形式**（决定 `MatchRoot` 与默认 `MatchedSpan`）→ **Plan 推导规则**（Call 级由 anchor 父槽位决定；Decl 级固定两步 Plan）。

**顶层形式**（解析失败 → 注册 fatal）：

| 顶层 | 示例 | `MatchRoot` | `MatchedSpan` |
|------|------|-------------|---------------|
| **CallPattern** | `Try($inner)` | `Call` | anchor `CallExpr` |
| **StmtPattern** | `$lhs ... := Try($inner)`、`$lhs ... = Try($inner)`、`var $lhs ... = Try($inner)`、`return $vals ... , Try($inner)`、`Try($inner);` | `Stmt` | 整条 stmt（`var` 形为 `*ast.DeclStmt`） |
| **DeclPattern** | `type $item struct { Derive[$iface] $field ... }` | `Decl` | `TypeSpec` |

**Callee（stub）literal 匹配 — invoked name**（Q1）：pattern 中 call 位置的 literal（如 `Try`）MUST 与 anchor 的 **invoked name** 一致——callee 为 `Ident` 时比 name；为 `SelectorExpr` 时比 `Sel.Name`。故 `Try($inner)` MUST match `Try(...)` 与 `tr.Try(...)` / `provider.Try(...)`。MAY 写限定形式 `tr.Try($inner)` 以额外约束 selector 左端（alias 敏感）。

**Assign lhs ellipsis**（Q5）：StmtPattern assign 形 MUST 支持三种 normative 写法：

| 源码形态 | pattern 示例 | `MatchedSpan` | `Plan` |
|----------|--------------|---------------|--------|
| `x, err := Try(...)` | `$lhs ... := Try($inner)` | `*ast.AssignStmt`（`Tok=DEFINE`） | `BlockStmts` OneToMany |
| `x, err = Try(...)` | `$lhs ... = Try($inner)` | `*ast.AssignStmt`（`Tok=ASSIGN`） | 同上 |
| `var x, err = Try(...)` | `var $lhs ... = Try($inner)` | `*ast.DeclStmt`（内嵌 `GenDecl{Tok:VAR}`） | `BlockStmts` OneToMany |

`$lhs ...` 绑定左侧全部名字项（0+），每项 `Underlying()` 为 `ast.Expr`；迭代顺序 MUST 为源码顺序。首版 **MUST NOT** 支持包级 `var`（`file.Decls` 内 `GenDecl`）。

**Return ellipsis**（Q3）：StmtPattern `return $vals ... , Try($inner)` 中 `$vals ...` 绑定 macro 之前的前缀 results（0+）；anchor MUST 为后缀 result 槽中的 stub call。Stmt 级 Plan 为 `BlockStmts` OneToMany。

**ExprStmt 窄替换 + 语法糖**（Q2）：CallPattern `Try($inner)` 在 `ExprStmt` 父上下文中 Plan MUST 为 `ExprSlot` OneToOne（仅换 call 子树，**不**自动提升 stmt 级）。若要 1→N stmt，MUST 使用 StmtPattern `Try($inner);`（`MatchRoot=Stmt`，`BlockStmts` OneToMany）。

**Call 级 Plan 由父槽位唯一决定**（运行时）：对 CallPattern，给定 anchor site，父链 MUST 唯一映射 Plan 一步——`AssignRhs`、`ReturnResults[j]`、`ExprSlot` 等；父不在支持集合 → match 失败（不得降级）。StmtPattern 则固定 `BlockStmts` OneToMany。

**注册期 Plan 歧义 fatal**：仅当**同一顶层形式**在**同一父上下文类**下可静态导出两种 Plan 时 fatal。`Try($inner)` 在 assign vs return 父上下文 Plan 不同**不算**歧义（运行时父链唯一）。

**Decl struct — 顺序无关约束集**（Q4 修正）：DeclPattern 的 struct `FieldList` **MUST NOT** 按书写顺序与源码字段对齐；MUST 作**无序约束集** match：

- `EmbedMarker[$iface]`（如 `Derive[$iface]`）：恰好一个匿名 embed，invoked name 与 marker 一致，绑定 `$iface`。
- `$field ...`：绑定 struct 内**全部具名字段**（`Names != nil`），**不含**已匹配 embed；0 项合法。首版 **MUST NOT** 支持逐字段 `$name $type` pattern（读 Tag 等经 `Elems("field")` + `*ast.Field`）。
- pattern 内 `Derive[$iface]` 与 `$field ...` 的**书写顺序 MUST NOT 影响** match；`{ Derive[$iface], $field ... }` 与 `{ $field ..., Derive[$iface] }` **语义等价**。
- `$field ...` 绑定迭代顺序 MUST 为源码 `Fields.List` 顺序（仅遍历稳定性，非 match 条件）。
- 首版每 struct **至多一个** marker embed。

**SyntaxCase clause 顺序**：更宽 **StmtPattern** SHOULD 排在更窄 **CallPattern** 之前（如先 `$lhs ... := Try($inner)`，再 `Try($inner)`）。

**Ellipsis 绑定 API**：`Bindings` MUST 提供 `Elems(name) ([]Syntax, bool)` 读取 `$...` 捕获列表；单项捕获仍用 `Get`。

**理由**：today 六种 `SpliceTarget` 由 pattern 宽度 + 父槽位推导；Decl 语义不应强加字段顺序；封闭子集可注册期校验、可单测。

### D18：Decl ResolveSite — embed 为 anchor

**选择**：

| 站点 | `ResolveSite` anchor | `site.MacroPos()` | match 根 |
|------|---------------------|-------------------|----------|
| **Call** | `*ast.CallExpr`（待展开 stub 调用） | `call.Pos()` | anchor Call + enclosing stmt |
| **Decl** | `*ast.Field`（匿名 embed marker 字段） | `embedField.Pos()`（与 today `DeclContext.MacroPos()` 一致） | enclosing `*ast.TypeSpec` |

Decl 展开时引擎 MUST 以 embed `*ast.Field` 构造 `site`；`DeclPattern` match 在 enclosing `TypeSpec` 上执行；`MacroPos` 供 `//line` / `StampStmtPos`（Decl 适用时）。

**理由**：embed 字段是 Decl 宏的物理锚点；`MacroPos` 对齐 today 行为；TypeSpec 为 pattern 语义边界。

### D19：裸 Expander 须经 `site.Match`

**选择**：normative 作者路径为 `SyntaxRules` / `SyntaxCase`。Provider MAY 实现手写 `Expander`，但 MUST 在返回 `out` 前对 `site` 调用 `Match(pattern)` 写入 meta 槽（机制与 `SyntaxCase` 相同）。**MUST NOT** 在 `macro` 包暴露 `SetMatchMeta` 给 provider。未调用 `Match` 即返回 → meta 槽空 → 引擎在 `ValidateSplice` 前失败。

**理由**：保留高级定制能力，同时强制贴回计划仍由 pattern 决定，避免第二套 Target API。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| **BREAKING** 面大 | Call：`TargetToPlan` adapter + author-guide；Decl：强制 SyntaxCase 迁移（无 adapter） |
| Decl provider 未迁移 | 仓库内建 Derive 随 change **blocking** 改写；文档给 Derive 示例 |
| Plan 与 pattern 不一致 | 注册期 fatal；golden + `Clause.Plan` override |
| pattern 子集不足 | **D17** 封闭子集 + 文档；L2 Transform 逃生 |
| Decl pattern 顺序误解 | **D17** 明确无序约束集；author-guide 示例 embed 在前/在后等价 |
| return/lhs ellipsis 实现复杂 | 单测覆盖；注册期解析校验 |
| Derive 生成 methods 与已有 method 冲突 | Apply 前 Validate 或 typecheck 报错；文档说明 pattern 边界 |
| 合并 quote 回归 | 移植现有 quote 单测 |
| EnclosingSignature 与 today EnclosingFunc 行为漂移 | 对照测试 Try assign/return golden |
| Decl Transform 滥用 `Underlying()` | author-guide 列 D16 映射表；Derive 示例展示 MacroTag + TypeOf |

## Migration Plan

1. **M1**：Syntax、Match（含 SplicePlan）、**site meta 槽**、Quote（合并）、ResolveSite、EnclosingSignature、ValidateSplice/Apply 骨架
2. **M2**：SyntaxRules / SyntaxCase、新 Expander 引擎路径
3. **M3**：Call adapter（`TargetToPlan`）+ mactest；**Derive 等 Decl provider 改写**；author-guide
4. **M4**：删除 Call adapter、旧 Call API、`macro/quote` 子包；删除 `DeclExpander` / `DeclExpandResult` / `ApplyDeclExpandResult`
5. contrib Call 迁移 optional；**contrib Decl 若存在则 blocking**

**回滚**：M3 阶段 MAY 保留 Call adapter 分支；Decl 无 adapter 回滚路径（须保留旧引擎分支或 revert change）。

## Open Questions

- `Text("%v")` 等 literal Syntax API 命名（`Text` vs `StringSyntax`）
- adapter 保留几个 minor release（建议 1，**仅 Call**）

（已闭合）Decl embed 元数据（MacroTag、类型实参等）→ **D16**：`Bindings` + `Underlying()` + `ParseMacroTag` / `Types().TypeOf`。
（已闭合）Pattern 语言首版子集 → **D17**（顶层形式、invoked name、Plan 推导、Decl 无序约束、ellipsis API）。
（已闭合）Ellipsis 列表读取 → `Bindings.Elems(name)`（**D17**）；Quote 侧仍用 `map[string]Syntax` + `#field ...`。
（已闭合）Decl ResolveSite / anchor → **D18**（embed `*ast.Field`；`MacroPos` = embed 位置；match 根 = `TypeSpec`）。
（已闭合）Assign 覆盖 `:=` / `=` / `var ... =` → **D17** assign 表（`var` 形 `MatchedSpan` = `*ast.DeclStmt`）。
（已闭合）逐字段 `$name $type` → 首版 **不支持**；normative 仅 `$field ...`（修正 macro-pattern scenario）。
（已闭合）裸 Expander → **D19**（MUST `site.Match`；不暴露 `SetMatchMeta`）。
（已闭合）引擎包路径 → **`internal/expander/`**（保留现有包名；更新 tasks/spec 引用）。
