## ADDED Requirements

### Requirement: 语法桩禁止值用法（expand 期校验）

在宏调用识别与展开之前，展开引擎 MUST 对**宏主文件**执行语法桩**值用法**校验：凡 `go/types` 解析为**当前注册表**中已登记 `(importPath, stubName)` 的 provider **包级**语法桩，且该 AST 节点**不是**宏直调 `CallExpr` 的 callee（剥除外层 `ParenExpr` 后等于 `CallExpr.Fun`），expand MUST 失败。

校验 MUST 使用与宏调用识别相同的符号解析规则（包级 `*types.Func`、`SelectorExpr` 的 `X` 为 `*types.PkgName`、dot-import 裸 `Ident` 等）。引擎 MUST NOT 仅依据函数名匹配。

校验 MUST 在**同一宏主文件**内一视同仁：MUST NOT 为 provider 作者、死代码、反射等场景单独豁免。

错误信息 MUST 含文件名、行号、列号、桩名（及 import path 或本地 import 名若可解析），并 MUST 提示宏桩须以 `pkg.Stub(...)`（或 dot-import 下 `Stub(...)`）**直接调用**，MUST NOT 作为函数值使用。引擎 MUST NOT 静默跳过。

#### Scenario: 桩作函数实参传递

- **WHEN** 宏主文件已 link 并注册 `example.com/try` 的桩 `Try`，源码为 `apply(try.Try)` 且 `apply` 为普通函数
- **THEN** expand MUST 在 `try.Try` 处失败，且 MUST NOT 展开该文件

#### Scenario: 桩赋值给变量

- **WHEN** 已注册桩 `Try`，源码为 `fn := try.Try`
- **THEN** expand MUST 在 `try.Try` 处失败

#### Scenario: 桩作为 return 值

- **WHEN** 已注册桩 `Try`，源码为 `return try.Try`
- **THEN** expand MUST 在 `try.Try` 处失败

#### Scenario: reflect 获取桩函数值

- **WHEN** 已注册桩 `Try`，源码为 `reflect.ValueOf(try.Try)` 或 `reflect.TypeOf(try.Try)`
- **THEN** expand MUST 在 `try.Try` 处失败

#### Scenario: 死代码中的桩值引用

- **WHEN** 已注册桩 `Try`，源码为 `if false { _ = try.Try }`
- **THEN** expand MUST 在 `try.Try` 处失败（引擎 MUST NOT 做可达性分析而跳过）

#### Scenario: 直调仍为合法宏调用

- **WHEN** 已注册桩 `Try`，源码为 `return try.Try(expr)` 或 `(try.Try)(expr)`
- **THEN** 值用法校验 MUST NOT 对该 `try.Try` 报错；引擎 MUST 按现有规则识别为宏调用并展开

#### Scenario: 未 link 的宏库不校验值用法

- **WHEN** 宏主文件 `import` 了 `example.com/try` 但本次 expand 的 `linked` 未包含该 path（注册表无 `Try` 桩）
- **THEN** 引擎 MUST NOT 因 `var _ = try.Try` 等值用法失败；对 `try.Try(...)` 直调 MUST 仍按现有「未注册桩 / 展开失败」规则处理

#### Scenario: shadow 同名不误报

- **WHEN** 宏主文件包内定义 `func Try(int) int` 且 `return Try(1)` 或 `_ = Try`
- **THEN** `go/types` 将 `Try` 解析为本包函数，非 provider 桩
- **THEN** 值用法校验 MUST NOT 报错，且宏识别 MUST NOT 将此类调用识别为宏（与现有 shadow 行为一致）

#### Scenario: 方法名与桩同名不误报

- **WHEN** 源码为 `s.Try(1)` 且 `Try` 为类型 `S` 的方法
- **THEN** 值用法校验 MUST NOT 报错，且 MUST NOT 识别为宏调用

#### Scenario: 嵌套直调中的内层桩

- **WHEN** 源码为 `outer(try.Try(1))` 且 `Try` 已注册
- **THEN** 内层 `try.Try` 作为 `CallExpr` callee MUST NOT 触发值用法错误；外层 `outer(...)` MUST NOT 被识别为宏

#### Scenario: 校验失败阻断写回

- **WHEN** 值用法校验在宏主文件任一处失败
- **THEN** expand MUST 失败，且 MUST NOT 写回 `*_macro_gen.go` 或部分展开结果

## MODIFIED Requirements

### Requirement: 宏调用识别与语义校验分离

展开引擎 MUST 将「是否为宏调用」（识别）、「语法桩是否被当作函数值使用」（引擎级值用法校验）与「宏调用是否合法」（provider 语义校验）分离。识别阶段与值用法校验阶段 MUST NOT 依赖特定宏的实参或返回类型规则。

#### Scenario: 识别不校验 Try 载荷

- **WHEN** 引擎识别到对已注册 `Try` 桩的调用
- **THEN** 识别阶段 MUST 仅依据 `go/types` 符号与注册表判定为宏调用，载荷合法性 MUST 由 `TryExpand` 在语义校验阶段处理

#### Scenario: 值用法校验不校验 Try 载荷

- **WHEN** 宏主文件含 `try.Try(badPayload)` 且实参在 `TryExpand` 语义下非法
- **THEN** 值用法校验 MUST NOT 因载荷非法而失败；失败 MUST 由 `TryExpand` 在展开阶段返回

### Requirement: 扫描范围

引擎 MUST 仅在宏主文件（含 `macro` build tag 的源文件）上识别宏调用**并**执行语法桩值用法校验，且 MUST NOT 将 `*_macro_gen.go` 作为上述扫描输入。

#### Scenario: 生成侧不参与识别

- **WHEN** 包内同时存在 `foo.go`（`macro` tag）与 `foo_macro_gen.go`（`!macro` tag）
- **THEN** 引擎 MUST 仅在 `foo.go` 上扫描宏调用与桩值用法
