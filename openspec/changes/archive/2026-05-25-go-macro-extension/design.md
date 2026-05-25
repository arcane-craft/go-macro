## Context

`go-macro` 是一个从零构建的 Go 过程宏框架。Go 编译器不接受自定义语法，因此宏必须在**构建前**将源码中的宏调用改写为标准 `go/ast` 可解析的 Go 代码。参考实现为 Rust 过程宏与 C# Source Generator 的混合模型：作者编写「语法桩 + 展开函数」，工具链在 `go build` 之前完成 AST 变换。

约束：
- 展开产物必须是合法 Go 源码，可通过 `go test` / `go vet`
- 宏桩函数在运行时 MUST panic，防止误用
- 首版以文件级展开为主，不修改 go.mod 或编译器本身

## Goals / Non-Goals

### 用户故事优先级

| 优先级 | 用户故事 | 说明 |
|--------|----------|------|
| **P0** | **宏作者能编写并注册自己的宏** | 框架目标：provider 包、`syntax-id`、`Expand`、识别与写回 |
| P1 | 宏使用方在业务包中调用已 import 的宏 | 与 P0 共用 expand / 方案 C 写回 |
| P2 | 参考宏 `Try` 演示错误处理 | **示例**，验证 P0/P1；不限制框架仅服务 error |

**Goals:**

- 提供稳定的**宏作者** API（`macro.Context`、`ast.Node`、provider 注册约定）
- 实现通用展开器分发（`syntax-id` → `Expand`）、识别（§3）、错误报告
- 支持 `go generate` 与 `go tool macro expand`
- 提供 `syntax-try` 参考实现与 `ReadFile` 级端到端测试（证明框架可用）
- 提供 `syntax-inline` 极简表达式宏示例（如 `Inline`），与 Try 并列，证明框架不绑定 error 语义

**Non-Goals:**

- 不实现编译器插件或修改 `go build` 内部流程
- 不支持声明式宏（仅过程宏）
- 首版不支持跨包宏内联、不支持在 `go test` 运行时动态展开
- **不**对依赖模块执行 macro expand（由库作者提交生成物）
- 不实现完整的 hygienic macro（宏卫生）——后续迭代

## Decisions

### 0. 框架与 Try 边界（已定，2026-05-22）

探索结论已固化，实现与文档须遵守：

| # | 决策 | 内容 |
|---|------|------|
| 1 | 规范分层 | **框架**（§8.2 splice 矩阵、`macro-core` / `macro-expander` spec）不含 Try/error 语义；`syntax-try` 的 Stmts、`if err`、桩族、k 校验等见 **Decision 7** 与 `specs/syntax-try/spec.md` |
| 2 | `ExpandResult.Exprs` | **首版保留**；`Exprs` 供罕见「仅替换 return 表达式列表」宏。`syntax-try` 在 `SiteReturn` **MUST NOT** 使用 `Exprs`（见 syntax-try spec） |
| 3 | 第二示例宏 | 首版 MUST 提供 `syntax-inline`（单桩表达式宏），与 Try 并列验收 P0 |
| 4 | `init provider` | **最小骨架**：单 panic 桩 + `Expand` 占位 + `mactest` 模板；**不**默认生成 Try 式多桩族 |
| 5 | `Context.EnclosingFunc` | **首版必选**，属通用语境 API；Try 专用规则（error 在最后、k 与桩名）**不得**进入 `macro` 包 |
| 6 | 多 error 宏 | 允许未来存在多个 error DSL（不同 syntax-id）；载荷/error 约定仅存在于各 provider 的 `Expand`，引擎无 `syntax-try` 特判 |
| 7 | 验收切割 | **P0**：假宏或 `syntax-inline` + 识别 + splice + mactest + `init provider`（最小）；**P2**：`syntax-try` 全语义 + ReadFile golden。引擎实现不得为 Try 硬编码分支 |
| 8 | 官方宏库 | `inline`、`try` 为**官方宏库**（同模块维护），非 CLI 内置；宏主文件 **import** 后由 `expander` 官方目录衔接 `Expand`；`cmd/macro expand` 不向引擎传入默认 provider 列表 |

### 1. 宏注册：syntax-id + 多桩名 + 展开器（通用）

宏框架与具体宏（如 `try` 包的 `Try` 族）**解耦**。`try` 仅为示范宏特性的参考实现。

- 宏**提供者包**（provider）：含 `//macro: <syntax-id>` 标注的 `XxxExpand` 函数，以及一个或多个**包级**语法桩（panic 函数）。
- 注册表（全局或按模块构建）：
  - `syntax-id` → `Expand(ctx, in) (out, err)` 实现
  - **桩函数名**（如 `Try`, `Try2`）→ `syntax-id`
- 同一 `syntax-id` MAY 绑定多个桩名，共享一个 `Expand`（`try` 即此模式）。
- **注册范围（已定）**：仅处理**宏主文件所在包**通过 `import`（含 dot-import、别名）**实际引入**的 provider 包。对未被 import 的 provider **不**扫描、**不**注册，即使同一 module 内存在 `//macro:` 定义（含官方宏库 `inline`、`try`）。
- 注册步骤：从待展开包的 macro 主文件收集 import 路径 → 与 **官方宏库目录**（`expander` 内，路径 → `Expander`）及调用方传入的 `extra []Provider` 求交 → 仅对命中路径解析 provider AST/types → 构建「桩名 → syntax-id → Expand」表。
- **官方宏库（Decision 8）**：`inline`、`try` 与框架同模块发布，供示例与可选依赖；**不是** `cmd/macro` 默认注册的宏。用户须在宏主文件中 import；`go tool macro expand` 调用 `ExpandPackages(..., nil)`。
- **自研 provider**：未列入官方目录时，须通过自定义 expand 入口传入 `extra`（或未来插件机制），因 Go 须在进程内持有 `Expander` 函数指针。

**理由**：与 Go 可见性一致，避免误扫未使用宏；加快 expand。  
**备选**：扫描整个 module 下所有 `//macro:` 包——可能注册到未 import 的宏，首版弃用。  
**备选**：每个宏名独立 `Expand` 函数——样板多，仅当语义完全无关时使用。

### 2. AST 抽象层：包装 `go/ast` 而非重新发明 IR

`ast.Node` 作为对外接口，内部持有 `go/ast.Node` 与 `token.FileSet` 位置信息。展开器读写的仍是 Go 标准 AST。

**理由**：降低维护成本，格式化与打印可直接用 `go/format`。  
**备选**：自定义 IR——过度设计，首版弃用。

### 3. 宏调用识别（通用管线）

**识别**与**语义校验**分离：识别只判断「这是否为已注册宏的调用」；实参/返回/语境等由对应 `Expand`（如 `TryExpand`）在展开期用 `go/types` 完成（见 Decision 7 仅约束 `syntax-try`）。

#### 3a. 扫描范围

- 仅处理宏**主文件**（方案 C：`//go:build macro` 的 `foo.go`）。
- **不**扫描 `*_macro_gen.go` 及其它默认 build 文件。
- 使用 `go/packages` / `types.Config` 时加载 build tag `macro`，与主文件一致。

#### 3b. 语法候选（AST）

遍历 `*ast.CallExpr`，`Fun` 剥去括号后须为：

| 形态 | 示例 | 首版 |
|------|------|------|
| `*ast.Ident` | `Try(f())`（dot-import 后） | 支持 |
| `*ast.SelectorExpr` | `try.Try(f())`、`alias.Try(f())` | 支持 |
| 方法调用 / 方法值 | `x.Try(f())`、`f := x.Try; f()` | **不支持**（见下） |
| 泛型实例化调用 | `Try[T](f())` | 首版报错或忽略，待后续 |

**「不支持方法值」含义**：只识别**包级函数桩**上的调用（定义在 provider 包的 `func Try(...)`）。下列形式**不**视为宏调用，即使名字相同：

```go
type S struct{}
func (S) Try(int) int { return 0 }

var s S
s.Try(1)              // 方法调用：Fun = SelectorExpr，X 是变量 s，非包标识符
f := s.Try; f(1)      // 方法值：先取方法再 Call
```

原因：宏桩在 provider 包中是**包函数**；`go/types` 里 `Uses` 会指向 `S.Try` 的方法定义，与 import 的桩 `try.Try` 不同。首版用 `types.Info` 判定定义点所在包路径 + 是否为包级函数即可排除。若未来要支持「方法形态的宏」，需单独语法与注册规则，不在首版范围。

#### 3c. 符号确认（必须 `go/types`）

禁止仅靠「名字 ∈ 注册表」识别（避免同包 shadow `func Try()` 误展开）。

对每个候选 `CallExpr`：

1. `types.Info.Uses` 解析 `Fun` 中的标识符 → `types.Object`
2. 对象须为**包级**函数（非 method、`builtin`）
3. `Obj.Pkg().Path()` 须与**该宏主文件包所 import 的、且已扫描注册的** provider 包路径一致（不硬编码 `try`，不纳入未 import 的 provider）
4. `Obj.Name()` 须在注册表「桩名 → syntax-id」中

通过后得到 `(syntax-id, CallExpr)`，分发到对应 `Expand`。

#### 3d. 分发

引擎调用 `Expander(ctx, call *ast.CallExpr) (ExpandResult, error)`（见 §8）。核心引擎按 `CallSite` 与 `ExpandResult` 字段 splice AST，不假设特定宏语义（`Try` 规则在 `TryExpand` 内）。

#### 3e. 难例与测试矩阵（框架级）

| # | 场景 | 期望 |
|---|------|------|
| 1 | dot-import provider 的 `Macro(f())` | 识别并分发 |
| 2 | `pkg.Macro(f())` 显式 import | 识别 |
| 3 | import 别名 | `Uses` 指向 provider 包 |
| 4 | 调用方同包 shadow `func Macro()` | **不**识别 |
| 5 | 未 import provider | 不识别 |
| 6 | `x.Macro()` 方法调用 | **不**识别 |
| 7 | 嵌套宏调用 | 多个 `CallExpr` 各识别 |
| 8 | gen 文件残留宏调用 | 不扫描 / 报错 |
| 9 | 主文件无 `macro` tag | expand 失败 |
| 10 | 两个 provider 导出同名桩 | 以 `Uses` 的 `Pkg.Path` 区分 |
| 11 | 新 provider 已 import 且含 `//macro:` + 桩 | 注册并分发，无需改 expander 核心；官方库须在官方目录登记 `Expander` |
| 12 | module 内有 provider 但宏主文件未 import | **不**注册；调用该桩则按普通函数/编译错误处理，**不**展开 |
| 13 | 官方宏库 `try` 在 module 内但宏主文件未 import `try` | **不**注册 `syntax-try`；CLI 不得默认传入 try 的 `Expand` |
| 13 | 方法值 / 方法调用 `x.Macro()` | **不**识别（首版不支持） |

**弃用**：特殊注释标记调用点；纯名字匹配无 `go/types`；全 module 扫描注册 provider。

**参考宏**：`syntax-try` / `TryExpand` / `Try` 族桩为上述机制的具体实例（Decision 7）。

### 4. 展开管线

```
扫描包 → 解析 AST → 收集宏调用点 → 按调用点调用 *Expand → 替换 AST 子树 → 格式化写回
```

`Expander` 接收 `*ast.CallExpr`，返回 `ExpandResult`（`Stmts` / `Exprs` / `Expr`，见 §8）。引擎统一贴回 AST。

**理由**：过程宏标准模型；分离「宏语义」与「替换粒度」。  
**备选**：仅 `out ast.Node` 替换 `CallExpr`——无法表达 `Try` 等多语句展开，弃用。

### 5. 代码写回策略：方案 C（主文件 + 生成侧）

采用 **build tag 互斥的孪生文件**：**带宏的源文件保持原有主文件名**（`foo.go`），工具仅新增/更新生成文件；**不**要求改为 `.macro.go` / `foo_macro.go` 等其它主文件名。

| 角色 | 文件名 | build constraint | 谁编辑 | 默认 `go build` |
|------|--------|------------------|--------|-----------------|
| **宏源（主文件）** | `foo.go` | **用户**在源文件自行维护（须含 `macro`） | 开发者 | 否 |
| **展开产物** | `foo_macro_gen.go` | 工具根据宏源 constraint **推导** `!macro` 侧 | 仅工具 | 是 |

**用户工作流**

```
编辑 foo.go（主文件，含 Try；自行写 //go:build macro [&& 平台约束…]）
        │
        ▼
go generate / go tool macro expand
        │
        ├─▶ 不修改 foo.go 的 build tag（用户自行合并 macro 与已有 tag）
        └─▶ 生成/更新 foo_macro_gen.go（//go:build !macro …，DO NOT EDIT）
        │
        ▼
go build（无 -tags macro）──▶ 仅编译 foo_macro_gen.go
```

示例（平台 tag 由用户自行合并到主文件）：

```go
// foo.go — 主文件，用户维护
//go:build macro && linux
//go:generate go tool macro expand

package demo

func ReadFile() ([]byte, error) {
    file := Try(os.Open("hello.txt"))
    ...
}
```

```go
// foo_macro_gen.go — 工具生成
// Code generated by go-macro; DO NOT EDIT.
//go:build !macro && linux

package demo

func ReadFile() ([]byte, error) { /* 展开后标准 Go */ }
```

**理由**：保留 C 的 tag 互斥；主文件仍是团队熟悉的 `foo.go`；已有 `linux` 等与 `macro` 的合并方式交给用户，工具只负责生成互补的展开侧。  
**弃用**：`.macro.go` 主文件命名、工具重写主文件 `go:build` 头、将展开产物写回同名 `foo.go`（覆盖主文件）。

#### 5.1 主文件 build tag：用户负责，工具只读

- 宏源主文件 **MUST** 在默认 build 中排除：constraint **MUST** 包含标识符 `macro`（用户自行写成 `macro`、`macro && linux` 等）。
- 工具 **MUST NOT** 自动改写主文件的 `//go:build` / `// +build` 行（不注入、不剥离、不合并用户区）。
- 若 expand 时主文件 constraint **不含** `macro`：MUST 报错并提示用户在主文件加入 `macro`（及如何与已有平台 tag 合并），避免默认 build 编进含 `Try` 的源码。
- 若主文件含 `ignore` 且无 `macro`：MUST 报错（`ignore` 与「宏源 + 生成侧」模型不一致，见 README 说明与 A 方案差异）。

**生成侧 constraint（工具推导）**

- 解析主文件完整 constraint 表达式 `E_main`（`go/build`）。
- 生成 `E_gen`：将 `E_main` 中的 `macro` 标识符替换为 `!macro`，其它部分（`linux`、`linux || darwin` 等）**原样保留**。
- 示例：`macro && linux` → `!macro && linux`；仅 `macro` → `!macro`。
- `E_gen` MUST 校验合法；主文件若含 `ignore` 且无 `macro` 已在上条拒绝。

实现 helper：`ComplementMacroConstraint(eMain string) (eGen string, error)`，**仅**用于写 `foo_macro_gen.go` 头部。

#### 5.2 文件配对与扫描

- 配对规则：主文件 `foo.go` 含宏调用 → 展开输出 `foo_macro_gen.go`（固定后缀 `_macro_gen.go`）。
- 同一 package 内多组主文件各自生成对应 `_macro_gen.go`。
- 生成文件 **不** 复制主文件的 `//go:generate`（避免重复执行）；generate 指令留在主文件。

#### 5.3 IDE / gopls

主文件含 `//go:build macro` 时，默认 gopls 不分析该文件。推荐：

```json
"buildFlags": ["-tags=macro"]
```

README 说明：日常编辑的是 `foo.go`（主文件），与生成侧 `foo_macro_gen.go` 区分开。

#### 5.4 生成物提交与展开范围

**展开范围（已定）**：`go tool macro expand` **仅**处理**当前主模块**内的包（如 `expand ./...`）。**不**展开 module cache / 依赖库源码；**不提供** `-deps` 或等价「替依赖写 gen」能力。

**生成物提交（已定）**：

| 场景 | 策略 |
|------|------|
| **作为 module 依赖被引用的库** | 库作者 **MUST** 在发布前对本模块执行 `expand`，并将 `foo.go` 与 `foo_macro_gen.go` **一并提交**（或随发布产物提供等价已展开源码）。下游仅 `go build`，无需安装 macro、无需改依赖树。 |
| 应用 / 私有 monorepo（非对外库） | 维护者自选：提交 gen，或 CI/build 前对本模块 `go generate` / `expand` |

库作者 CI 建议：`go tool macro expand ./...` → `git diff --exit-code`（防止 gen 漂移）→ `go test`。

**库作者工作流：**

```
编辑 foo.go（macro 主文件）→ expand ./... → 提交 foo.go + foo_macro_gen.go → tag / go get
```

**下游工作流：**

```
go get lib@version → go build   （依赖库已含 _macro_gen.go）
```

**弃用 / Non-Goal**：对依赖模块执行 macro expand；要求下游在构建链中展开依赖源码。

### 6. CLI 与 go generate 集成

- `go tool macro expand [packages]`：展开指定包；**不**在 `cmd/macro` 中默认链接 `inline`/`try` 的 `Expand`（Decision 8）
- `//go:generate go tool macro expand`：标准 generate 钩子
- 官方宏库：宏主文件 `import github.com/arcane-craft/go-macro/inline` 或 `/try` 后，由 `expander` 官方目录在 expand 时自动衔接；与「仅注册已 import provider」一致

**理由**：与 Go 生态工具链一致，无需新全局命令名；官方库与用户自研宏同为「按需 import」的依赖，避免 CLI 隐含启用 Try/Inline。  
**备选**：独立二进制 `gomacro`——需额外安装步骤，作为别名可后续添加。  
**备选**：`cmd/macro` 内置全量官方列表——已弃用，与 Decision 8 冲突。

### 7. Try 宏展开语义与类型边界

#### 7.0 语法桩族（为何不能只有一个 `Try`）

Go 不支持按「多返回值个数」重载。单一桩：

```go
func Try[T any](v T, err error) T
```

仅能在**类型检查期**适配 callee 为 `(T, error)` 的调用；对 `(A, B, error)` 写 `Try(f())` 时，编译器无法把三个值塞进两个形参，**IDE/gopls 报错**，与展开器实际支持的语义不一致。

因此采用 **「按载荷个数分桩 + 单一展开器」**：

| 桩名 | 形参（最后一项均为 `error`） | 桩返回值 | 适配 callee | 典型调用 |
|------|------------------------------|----------|-------------|----------|
| `Try0` | `(err error)` | （无） | `(error)` | `Try0(expr);` / `return Try0(expr)` |
| `Try` | `(v T, err error)` | `T` | `(T, error)` | `x := Try(f())` / `return Try(f())` |
| `Try2` | `(a A, b B, err error)` | `(A, B)` | `(A, B, error)` | `a, b := Try2(f())` / `return Try2(f())` |
| `Try3` | `(a A, b B, c C, err error)` | `(A, B, C)` | `(A, B, C, error)` | 三载荷 + error |
| `Try4` | 四载荷 + `error` | 四元组 | 四载荷 + error | 少见 API |

约定：

- **`Try` 为 `Try1` 的简名**（与用户示例 `Try(os.Open(...))` 一致），仅覆盖 1 个载荷 + `error`。
- 载荷数 `k > 1` 时，调用点 MUST 使用 `Try{k}`（如 `Try2`），**不得**用 `Try` 硬套。
- 所有桩函数体 MUST `panic`；共享同一 `//macro: syntax-try` 的 `TryExpand`。
- 注册表：`Try`, `Try0`, `Try2`, `Try3`, `Try4`（及可选别名）→ 同一 `syntax-id` → `TryExpand`（仅当宏主文件 import `try` 包时激活，Decision 8）。
- **展开期**以 `go/types` 校验 callee/外层；桩仅辅助 macro 源文件的静态类型检查。若调用名与 callee 载荷数不一致（如对 `(A,B,error)` 写了 `Try`），`TryExpand` MUST 报错并提示改用 `Try2`。

首版 `try` 包导出：`Try0`, `Try`, `Try2`, `Try3`（`Try4` 视需要）。桩用泛型保证各载荷类型可推断。

```go
package try

//macro: syntax-try
func TryExpand(ctx macro.Context, call *ast.CallExpr) (macro.ExpandResult, error) { ... }

func Try0(err error) { panic("Try0 is a macro") }

func Try[T any](v T, err error) T { panic("Try is a macro") }

func Try2[A, B any](a A, b B, err error) (A, B) { panic("Try2 is a macro") }

func Try3[A, B, C any](a A, b B, c C, err error) (A, B, C) { panic("Try3 is a macro") }
```

**备选（弃用）**：`func Try(...any) any` 单桩——类型信息丢失；或仅依赖展开期检查、不提供桩——macro 源文件 gopls 体验差。

#### 7.1 Go 返回习惯（硬约束）

与 Go 社区约定一致，**`error` MUST 为返回值列表中的最后一项**（具名/无名返回均适用）：

- **外层函数**（`Try` 调用所在函数）：签名中 MUST 至少有一个 `error`，且**最后一个**返回参数的类型为 `error`（或底层类型为 `error` 的类型别名）。
- **内层 `expr`**（`Try` 的实参调用）：MUST 为多返回值调用，且**最后一个**返回值为 `error`。

不满足则 `TryExpand` MUST 报错，并指向 `Try(...)` 调用位置，说明「外层函数 / callee 返回值排列不合法」。

不支持：`error` 不在最后一位、外层无 `error`、内层单返回且非 `error`、内层零返回。

#### 7.2 内层（callee）返回值形态

设 `expr` 的类型为 `(R₁, R₂, …, Rₙ₋₁, E)`，其中 `E` 为 `error`，前缀 `R₁…Rₙ₋₁` 可为 0 个或多个（记 `k = n-1` 个「载荷」返回值）。

| k | 桩名 | 含义 | 调用示例 |
|---|------|------|----------|
| 0 | `Try0` | 仅 `(error)` | `Try0(expr);` / `return Try0(expr)` |
| 1 | `Try` | `(T, error)` | `x := Try(expr)` |
| ≥2 | `Try{k}` | `(R₁,…,Rₖ,error)` | `r1, r2 := Try2(expr)`；单变量绑定仅取 `R₁` 时仍可用 `Try` 的 k=1 桩，多载荷须用 `Try{k}` |

内层类型检查 MUST 在展开期完成（`go/types` 或等价），不能仅靠语法桩。

#### 7.3 外层（enclosing）函数校验

展开前 MUST 解析 `Try` 所在函数的 `*ast.FuncDecl` / `FuncLit` 的 Results：

1. 若 Results 为空或最后一个非 `error` → **非法**：报错「当前函数未按 Go 习惯以 error 结尾，不能使用 Try」。
2. 记外层返回为 `(O₁, …, Oₘ₋₁, Oₑ)`，`Oₑ` 为最后一个且为 `error`。
3. 错误路径上的 `return` MUST 生成 **m 个** 返回值：前 `m-1` 个为各自类型的零值，`m` 为 `_err`（或具名 `err` 字段）。

成功路径：按调用语境（赋值 / return / 表达式）使用载荷临时变量。

#### 7.4 调用语境与展开形状

**赋值** `lhs := Try(expr)`（k=1）或 `r1, r2 := Try2(expr)`（k=2）等：

```go
_v1, _v2, …, _vk, _err := expr
if _err != nil {
    return _zero(O₁), …, _zero(Oₘ₋₁), _err
}
lhs = _v1   // 单值绑定取第一个载荷
```

**return** `return Try(expr)` / `return Try2(expr)` 等：

- MUST 展开为**完整错误处理语句序列**（`ExpandResult.Stmts`），**不得**简化为仅替换 `return` 表达式列表（不用 `Exprs` 把 `return Try(e)` 改成 `return _v, nil` 而省略 `if err != nil` 块）。
- k≥1 典型形状：

```go
_v1, …, _vk, _err := expr
if _err != nil {
    return <zeros>, _err   // 未来可在此插入 error 包装（%w、附加上下文）
}
return _v1, …, _vk, nil
```

- k=0（`Try0`）：`Stmts` 为 `_err := expr` + `if _err != nil { return …, _err }`，成功路径无额外 `return`（原 `return Try0` 整句被替换块承接）。

**理由**：保留完整 `if err != nil { return … }` 块，便于后续扩展 error wrapping，与赋值/语句语境一致。

**语句** `Try0(expr);`（k=0，内层仅 error）：

```go
_err := expr
if _err != nil {
    return …, _err
}
```

**非法语境示例**（MUST 报错）：

- `func f() int` 内任意 `Try`
- 内层 `(error, int)`（error 非最后）
- 外层 `(error, int)`（error 非最后）
- k=0 但写法 `x := Try(expr)` 或 `x := Try0(expr)` 赋值（无载荷可绑定）
- k≥2 但调用名仍为 `Try`（应改用 `Try{k}`，如 `Try2`）
- k≥1 但外层无任何 `error` 返回

#### 7.5 具名返回值

若外层使用具名返回（`func f() (data []byte, err error)`），展开时：

- `return` 语句 SHOULD 优先使用具名标识符（`return nil, err` 而非仅位置零值），以利可读性与 `defer` 场景。
- 仍 MUST 满足「最后一个为 `error`」；具名顺序与类型检查以 AST + types 信息为准。

#### 7.6 错误信息

报错 MUST 包含：文件名、行号、类别（外层非法 / 内层非法 / 语境非法）、简要修复建议（例如「请将 error 放在返回列表最后」）。

### 8. Expand 函数契约（已定）

宏作者实现的展开器须符合以下 API；`try.TryExpand` 为参考实现。

#### 8.1 类型定义

```go
type CallSiteKind int

const (
    SiteAssign CallSiteKind = iota // lhs := Macro(...)
    SiteReturn                     // return Macro(...)
    SiteStmt                       // Macro(...); / Try0
    SiteExpr                       // 其它表达式位置（纯表达式宏）
)

type Context interface {
    FileSet() *token.FileSet
    Types() *types.Info
    Package() *types.Package
    Call() *ast.CallExpr
    StubName() string   // 如 Try, Try2
    SyntaxID() string   // 如 syntax-try
    Site() CallSiteKind
    EnclosingFunc() *ast.FuncDecl // 或 *ast.FuncLit
    TempIdent(prefix string) *ast.Ident
    MacroPos() token.Pos          // 宏调用位置，供 //line 与 ErrorAt
}

// 三选一为主；引擎按 Site 与「设置了哪些字段」决定 splice 方式
type ExpandResult struct {
    Stmts []ast.Stmt // 替换整条语句（Assign/Return/ExprStmt）
    Exprs []ast.Expr // 仅替换 return 的 Results 列表（少用；Try 不用）
    Expr  ast.Expr   // 仅替换 CallExpr（表达式宏）
}

type Expander func(ctx Context, call *ast.CallExpr) (ExpandResult, error)
```

Provider 注册：`//macro: <syntax-id>` 标注的函数须为 `macro.Expander` 签名（或包内 `func XxxExpand(...) (macro.ExpandResult, error)`）。

#### 8.2 引擎 splice 规则

| `CallSite` | `ExpandResult` 字段 | 行为 |
|------------|----------------------|------|
| `SiteAssign` | `Stmts` 非空 | 用 `Stmts` 替换整条 `AssignStmt`（可多条，如多值赋值 + if + 最终赋值） |
| `SiteReturn` | `Stmts` 非空 | 用 `Stmts` 替换整条 `ReturnStmt`（**完整错误处理**，见 §7.4） |
| `SiteStmt` | `Stmts` 非空 | 用 `Stmts` 替换 `ExprStmt` |
| `SiteExpr` | `Expr` 非空 | 仅用 `Expr` 替换 `CallExpr`（**表达式宏**） |
| 其它组合 | — | 报错：Expand 与调用语境不匹配 |

`Exprs` 保留供特殊宏（首版保留，见 Decision 0.2）；引擎仅按上表 splice，不解释 error 语义。`syntax-try` 在 `SiteReturn` 禁止 `Exprs` 的规则见 `specs/syntax-try/spec.md`，**非**框架 MUST。

#### 8.3 表达式宏（允许）

宏作者 MAY 仅返回 `ExpandResult{Expr: ...}`，且 `Site()` 为 `SiteExpr`（或其它引擎文档允许的语境）。引擎只替换宏调用子树，不插入额外语句。

首版参考：`syntax-inline` 的 `InlineExpand`（Decision 0.3）；亦可假想 `Must(expr)` 由内联实现。

#### 8.4 Try 专用展开约定（非框架）

以下内容属于 **`syntax-try` provider**，已迁至 Decision 7 与 `specs/syntax-try/spec.md`，**不得**写入引擎或 `macro` 包：

- 在 `SiteAssign` / `SiteReturn` / `SiteStmt` 使用 `Stmts` 展开 error 路径
- `return Try(...)` 须完整 `if err != nil { return … }` 块，禁止 `Exprs` 简化
- error 在内外层返回列表末尾、k 与 `Try{k}` 桩名一致等

#### 8.5 嵌套调用

同一函数内多个宏调用：收集后建议**从后往前**按偏移替换，或每轮替换后重新 typecheck；避免外层 `CallExpr` 在替换内层后失效。

### 9. 宏作者体验（已定）

#### 9.1 AST 辅助库：首版轻薄

- `macro` 包首版仅提供 **§8** 中的 `Context`、`ExpandResult`、`TempIdent`、统一错误辅助（如 `ErrorAt`）等最小能力。
- **不**在首版引入厚重的 `astbuilder`；待第二个及更多 provider 宏落地后，再按真实重复模式抽取封装。
- `TryExpand` 允许直接使用 `go/ast` + `go/format` 作为参考实现样板。

#### 9.2 `//line` 指向宏源码（需要）

写入 `*_macro_gen.go` 时，引擎 MUST 在生成语句块中插入 `//line` 指令，使调试器、覆盖率、`go test -cover`、panic 堆栈将执行位置映射回**宏主文件**（`foo.go`）的行号，而非 gen 文件行号。

约定（示例）：

```go
//line readfile.go:12
_tmp, _err := os.Open("hello.txt")
//line readfile.go:13
if _err != nil { ... }
```

- 行号取自宏调用点或 `Expand` 触发的逻辑起点（由引擎在 `Context` 中提供 `MacroPos()` 或等价 API）。
- 格式化时 MUST 保留合法 `//line` 语义（`go/format` 对 line directive 有规则，实现须遵循）。

#### 9.3 Provider 测试：支持纯 `Expand` 单测

- Provider 包的 **`go test` 默认 MUST NOT 依赖** 使用方式的 `//go:build macro` 或全链路 `macro expand`。
- 推荐测试方式：
  1. **纯 Expand 单测**：构造 `*ast.CallExpr` + 伪造/加载 `types.Info` 的 `macro.Context`（`macro/mactest` 或 `macro_test` 辅助包），直接调用 `FooExpand`，断言 `ExpandResult` 的 AST 或打印快照。
  2. **可选集成测**：`testdata/` + golden 全链路（可放 `*_test.go` 带 `macro` tag 的子包或单独 `expandtest` 目录）。
- `syntax-try` 的 `TryExpand` MUST 具备纯单测，不强制先跑 CLI。

#### 9.4 文档：宏作者指南

- 仓库 MUST 提供 **`docs/author-guide.md`**（或 `docs/macro-author-guide.md`），面向 P0 宏作者，内容包括：
  - provider 包结构、`//macro:`、桩族命名、多桩名共享 `Expand`
  - `Expander` / `ExpandResult` / `CallSite` 契约
  - 纯 Expand 单测写法（`mactest`）
  - 使用方 macro 主文件 / gen 提交策略（简述，链到 README）
  - `go tool macro init provider` 脚手架用法
- 根 README 面向快速上手，详细契约以作者指南为准。

#### 9.5 `go tool macro init provider` 脚手架（已定：最小骨架）

CLI MUST 提供：

```bash
go tool macro init provider <pkgname> [--module <path>]
```

生成（或写入）**最小** provider 骨架（Decision 0.4），例如：

```
<pkgname>/
  expand.go      # //macro: syntax-<pkgname> + XxxExpand 占位实现
  stubs.go       # 单个 panic 桩（非 Try 式多桩族）
  expand_test.go # 纯 Expand 单测模板（使用 mactest）
  README.md      # 指向 docs/author-guide.md；链到 syntax-inline / syntax-try 作进阶示例
```

- `syntax-id` 默认 `syntax-<pkgname>`，可 flag 覆盖（首版可用 flag）。
- **MUST NOT** 默认生成 `Try0`/`Try2` 式多桩模板；多桩共享 `Expand` 的模式在作者指南中以 Try/inline 为例说明。
- 不替代业务项目的 `foo.go` / `foo_macro_gen.go` 创建（使用方模版可后续 `init usage`）。

#### 9.6 `syntax-inline` 表达式宏示例（已定）

- 包路径建议 `inline/`（或 `macrolib/inline`），`syntax-id` 为 `syntax-inline`。
- **官方宏库**：用户 import 后展开；`expander/official_providers.go`（或等价）登记 `ImportPath` 与 `InlineExpand`；`cmd/macro` 不默认注册。
- 单桩 `Inline[T any](v T) T`（panic）+ `InlineExpand`：在 `SiteExpr` 返回 `ExpandResult{Expr: call.Args[0]}`（或等价），演示 **仅替换 CallExpr**、无 error 语义。
- MUST 有 `mactest` 单测；可选极小 macro 主文件 + golden，供 P0 框架验收（与 `syntax-try` 的 P2 分离）。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 展开后源码与手写源码漂移 | golden + 可选 CI `git diff`；库作者可提交 gen 固化 |
| 依赖库未提交 gen 导致下游无法编译 | 文档与 spec 要求库作者发布前提交 gen；expand 仅本模块 |
| 宏桩被误调用导致 panic | 桩函数 panic + 静态分析提示（后续） |
| `go/ast` 打印格式与手写风格不一致 | 使用 `go/format` 统一格式化 |
| 多返回值类型推断复杂 | 展开期用 `go/types`；强制 error 在最后；具名返回单独测试 |
| 跨文件/跨包宏调用 | 首版：调用方包 import provider 即可；跨模块远程宏 Non-Goal |
| 方法值/方法调用形态的宏 | 首版仅包级桩；文档说明（Decision 3b） |
| 未 import 的 provider | 不注册、不展开（Decision 1） |

## Migration Plan

1. 初始化 `go.mod`（`github.com/arcane-craft/go-macro`）
2. **P0**：`macro-core` → `macro-expander` → `macro-codegen`；`syntax-inline` + 框架级识别/splice 测试；`init provider`（最小骨架）
3. **P2**：`syntax-try` 全语义 + `ReadFile` golden
4. 文档：`docs/author-guide.md` + README（区分框架契约 vs Try vs inline 示例）
5. 无既有用户，无需数据迁移

## Open Questions

- 是否在 expand 时检测主文件缺 `macro` tag 并给出可复制推荐的 constraint 模板（不自动写入）
