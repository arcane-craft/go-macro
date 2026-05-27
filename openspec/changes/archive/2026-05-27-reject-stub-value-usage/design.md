# Design: reject-stub-value-usage

## Context

`internal/expander.RecognizeMacroCalls` 仅遍历 `*ast.CallExpr`，通过 `resolveStubCall` 判定 `Fun` 是否为已注册 provider **包级**桩，再进入 `Expander` 语义阶段。宏主文件中若出现 `f(tr.Try)`、`fn := tr.Try`、`reflect.ValueOf(tr.Try)` 等，桩符号出现在**非 callee** 位置，识别器不会收录，expand 可能成功并写出 `*_macro_gen.go`，运行时在 `!macro` 构建下可能执行到桩的 `panic`。

规格要求：在识别/展开之前增加**引擎级值用法校验**；仅针对**当前注册表**已登记的桩；与 shadow、未 link、死代码等边界行为见 change spec 与 proposal。

现有可复用代码：`recognize.go` 中的 `objectStub`、`isPackageSelector`、`unwrapParen`；错误报告使用 `macro.ErrorAt`。

## Goals / Non-Goals

**Goals:**

- 宏主文件 expand 前，对已注册桩的「非直调」引用失败并报行列位置。
- 与宏识别共用 `go/types` + `Registry.HasStub` 判定，避免误报 shadow / 未 link。
- 直调 `pkg.Stub(...)` / `(pkg.Stub)(...)` / dot-import `Stub(...)` 不受影响。
- 测试覆盖 spec 中全部 Scenario；更新 `docs/author-guide.md` 与归档后 `openspec/specs/*`。

**Non-Goals:**

- 跨语句数据流（如仅 `fn(1)` 且 `fn` 来源不可静态见桩符号）。
- 修改 `TryExpand` / contrib 载荷或 Site 规则。
- 对未 link 的 import 做法向值用法诊断。
- 单独的 `go vet` 分析器（首版仅在 expand 路径执行）。

## Decisions

### D1：新函数 `ValidateStubValueUsage`，挂在 `ExpandFile` 最前

```text
ExpandFile:
  1. ValidateStubValueUsage(file, fset, info, reg)  // 失败即 return，不写回
  2. RecognizeMacroCalls → expand loop（不变）
```

**理由**：规格要求校验先于识别；失败时不应部分展开或写 gen 文件（`load.go` 在 `ExpandFile` 成功后才 `WriteGenFile`，无需改 load 顺序）。

**备选**：与 `RecognizeMacroCalls` 合并为一次遍历 → 省一次 AST 遍历，但混合「收集宏调用」与「报错值用法」职责，违背 spec 中的阶段分离表述；首版保持两函数。

### D2：复用 `objectStub` / `isPackageSelector`，抽到包内共享

将 `objectStub`、`isPackageSelector`（及必要时 `resolveStubExpr`）保留在 `recognize.go` 或移至 `stub_resolve.go`，供 `resolveStubCall` 与值用法校验共用。

对 `*ast.Ident` / `*ast.SelectorExpr` 节点：

1. 用与 `resolveStubCall` 相同的规则解析 `(stubName, importPath)`；
2. `reg.HasStub(importPath, stubName)` 为 false → 跳过。

**理由**：保证 shadow（本包 `*types.Func`）、方法（`Recv != nil`）、未登记桩与识别器行为一致。

### D3：判定「直调 callee」——父节点链 + `unwrapParen`

对每个命中已注册桩的 `Ident` / `SelectorExpr` 节点 `e`：

```text
isDirectMacroCallee(e) :=
  存在祖先 *ast.CallExpr `call`，使得 unwrapParen(call.Fun) == e
```

实现：一次 `ast.Inspect` 构建 `map[ast.Node]ast.Node` 父指针表（或维护访问栈），再对候选桩节点向上查找。

**理由**：覆盖 `(tr.Try)(x)`、`return tr.Try(x)`、嵌套 `outer(tr.Try(1))`（内层 Selector 是内层 `CallExpr.Fun）。

**备选**：仅检查 immediate parent → 无法处理多层 `ParenExpr`；不采用。

**注意**：只访问**整棵** `SelectorExpr`（如 `tr.Try`），不要对 `Sel` 单独报（避免重复报错）；`Inspect` 回调里对 `*ast.SelectorExpr` 已注册桩处理一次即可。dot-import 裸 `Ident` 同理。

### D4：全文件 Inspect，不做可达性分析

凡 AST 中出现已注册桩且非 callee → 报错，包括 `if false { _ = tr.Try }`。

**理由**：与产品决策一致；实现简单、行为可预测。

### D5：错误文案

统一使用 `macro.ErrorAt(fset, e.Pos(), ...)`，建议格式：

```text
macro stub %q must be invoked directly (e.g. %s.%s(...)), not used as a function value
```

`%s.%s` 优先用本地 import 名 + 桩名（从 `SelectorExpr` / `BuildImportMap` 反查），dot-import 时仅桩名。

`reflect.ValueOf` / `TypeOf` 参数中的桩：**不**单独分支（一视同仁）；若需更易读，MAY 在 message 中追加 “including reflect” 半句，但 MUST NOT 对 reflect 豁免。

### D6：reflect 与传参、赋值同一规则

不解析 `reflect` 包名做特殊逻辑；`reflect.ValueOf(tr.Try)` 中 `tr.Try` 已是「非 callee 的 SelectorExpr」→ 自然失败。

### D7：测试布局

新文件 `internal/expander/stub_value_test.go`（或 `validate_stub_value_test.go`）：

- 复用 `recognize_test.go` / `recognize_helpers_test.go` 的 provider 注册与 `typecheckWithProvider` 辅助；
- 表驱动或分用例对应 spec Scenario；
- 对 `ValidateStubValueUsage` 单测 + 可选 `ExpandFile` 端到端一例（校验失败不写 gen）。

未 link：registry 不 `RegisterProvider` → `HasStub` false → 用例期望 `Validate` 无错；直调仍由现有 expand 测覆盖。

### D8：文档

按 `specs/author-guide/spec.md` delta 更新 `docs/author-guide.md`「编写宏库」「宏使用方」——自然语言说明直调约束与 expand 失败，不引入新工具链。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 父指针表遗漏新 AST 节点类型 | 仅对 `Ident`/`SelectorExpr` 查父链到 `CallExpr`；其它父节点继续向上 |
| 与 `RecognizeMacroCalls` 重复遍历 | 宏主文件通常较小；两趟 Inspect 可接受；后续可合并优化 |
| 间接调用 `fn(1)` 未赋值行含桩 | 规格 Non-goal；文档说明须直调 |
| 误将普通同包函数当桩 | 已由 `HasStub` + provider 包路径约束；shadow 测例锁定 |
| 报错多次（同一表达式重复访问） | 每个 `SelectorExpr`/`Ident` 只处理一次；同一节点不重复入队 |

## Migration Plan

1. 实现并合并后，用户若曾在宏主文件传桩作值，**首次 expand 即失败**——属预期破坏性变更，无数据迁移。
2. 归档 change 时 `openspec archive` 将 delta 合并进 `openspec/specs/macro-expander/spec.md` 与 `author-guide/spec.md`。
3. 无需 contrib / examples 行为变更；examples 若无非法用法则测试不变。

## Open Questions

（无。边界已在 proposal 中闭合。）
