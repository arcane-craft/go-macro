## Why

当前宏库契约分散在三处：包级或 expand 函数上的单一 `//macro:`、靠 `panic` 函数体启发式识别桩、以及宏作者手写的 `register/` 与使用方自建的 `cmd/macroexpand`。作者与使用方都要理解 `expandtool.Register` 与 blank import，入门成本高，且「桩 ↔ Expander」的关联对读源码的人不直观。将 `//macro:` 下沉到每个桩函数与 Expander 函数，并由框架工具自动生成展开入口与 link 代码，可以把宏契约收敛为「注释 + 实现 Expand」，与用户对 DSL 标注的直觉一致。

## What Changes

- **BREAKING**：`//macro: <syntax-id>` 必须出现在**每个语法桩函数**与**每个 Expander 函数**的 doc 注释中（不再依赖包级单一注释或仅标注 expand）。
- **BREAKING**：注册表通过解析上述注释构建「桩名 → syntax-id → Expander 函数名」；**不再**以「函数体仅含 `panic`」作为桩的识别条件（运行时仍 RECOMMENDED panic，见 macro-core）。
- **BREAKING**：移除宏作者维护的 `register/` 包；`init provider` 脚手架不再生成该目录。
- 新增 **`cmd/macro expand`**（或等价子命令）：作为宏使用方默认展开入口；扫描模块/包依赖，**自动生成并更新** link 文件（blank import provider + `expandtool.Register`），再执行展开。
- **BREAKING**：文档与 `go:generate` RECOMMENDED 改为调用 `go run github.com/arcane-craft/go-macro/cmd/macro@latest expand`（或 `expand .`），不再要求项目内自建 `cmd/macroexpand` 或 blank import `contrib/register`。
- **BREAKING**：`go-macro-contrib` 删除 `register` 包；inline/try 仅在桩与 Expander 上使用 per-function `//macro:`。
- 更新 author-guide、README、examples 以反映新流程。

## Capabilities

### New Capabilities

- `macro-directive`：`//macro:` 在桩函数与 Expander 上的语义、解析规则、冲突与错误处理。

### Modified Capabilities

- `macro-core`：注册与查找模型；expandtool 与 Expander 绑定方式；init provider 脚手架；移除对 panic 启发式桩识别的 MUST。
- `macro-expander`：Provider 激活、宏调用识别、`RegisterLinked` 输入假设。
- `macro-codegen`：`go:generate` 推荐命令；自动生成 expand link 文件；零自建 expand 入口场景。
- `author-guide`：宏作者/使用方步骤、角色分工表。
- `macro-contrib`：移除 `register` 包要求；contrib 宏库注释形态。
- `syntax-inline` / `syntax-try`：桩与 Expander 的 `//macro:` 标注方式（若现有 spec 写死旧形态则更新）。
- `project-readme`：快速上手与 generate 一行命令。

## Impact

- `macro/registry.go`：`RegisterProvider`、`ProviderSyntaxID`、`isPanicStub` 逻辑重写。
- `internal/expander/load.go`、`expand.go`：link 来源改为解析 + 生成文件或内联构建。
- `cmd/macro`：新 `expand` 子命令、link 代码生成；`init provider` 模板变更。
- `macro/expandtool`：可能新增 `RunWithGeneratedLink` 或接受生成路径；`Register` 仍保留供生成代码使用。
- `docs/author-guide.md`、`README.md`、`examples/`、`openspec/specs/*`。
- **外部**：`go-macro-contrib` 删除 `register/`、各 provider 注释与文档（本变更 tasks 含协调项或 follow-up）。
