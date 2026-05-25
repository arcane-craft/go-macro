## Why

Go 语言缺少编译期元编程能力，样板代码和 DSL 往往需要手写重复逻辑或专用代码生成器。`go-macro` 旨在为 Go 提供**过程宏（procedural macro）**框架：让**宏作者**能在普通包中定义语法桩与 `Expand`，由工具链在构建前将调用点改写为合法 Go AST。`Try` 仅是演示该能力的**参考宏**（错误处理场景），不是项目唯一目标。

## What Changes

- 引入 `github.com/arcane-craft/go-macro` 核心库：宏注册、`Context`、AST 节点抽象、展开器调用约定
- 定义宏作者契约：语法桩 + `//macro:` + `Expander(Context, *CallExpr) (ExpandResult, error)`（支持语句宏与表达式宏）
- 实现宏展开管线：解析调用点 AST → 调用对应 `*Expand` → 写回展开后的 AST
- 提供代码生成入口：`go generate` 指令集成与 `go tool macro` 命令行工具
- 提供宏作者契约、`docs/author-guide.md`、`go tool macro init provider` 脚手架；`macro/mactest` 支持纯 Expand 单测；生成代码含 `//line` 指向宏主文件
- 附带**官方宏库** `syntax-try`（`Try` 族）与 `syntax-inline`（表达式宏）：验证端到端流程，由用户在宏主文件中 import 后启用，CLI 不默认注册
- 建立测试与示例：展开前后 golden；文档说明带宏库作依赖时须提交生成物

## Capabilities

### New Capabilities

- `macro-core`：核心库 API——`Context`、`ast.Node`、按**已 import 的 provider** 构建注册表与查找
- `macro-expander`：通用展开引擎——`go/types` 识别 provider 包级宏桩调用、按 syntax-id 分发 `Expand`（与具体宏无关）
- `macro-codegen`：代码生成集成——`go generate` 指令、`go tool macro` CLI、展开结果写回策略（同文件 / 生成文件）
- `syntax-inline`：官方宏库，极简表达式宏（单桩 + `InlineExpand`）；import 后展开；供 P0 验收
- `syntax-try`：官方宏库，参考宏族 `Try0`/`Try`/`Try2`/… + 统一 `TryExpand`；import 后展开；ReadFile 端到端验证（P2）

### Modified Capabilities

（无——仓库尚无既有 spec）

## Impact

- **新包结构**：`macro/`（核心）、`cmd/macro/`（CLI，不默认链官方宏库）、`expander/`（含官方宏库目录）、`inline/`、`try/`（官方宏库）、`internal/`（AST 解析与写回）
- **构建流程**：编辑宏主文件 `foo.go`；`go tool macro expand ./...` 仅展开**本模块**；**对外发布的库 MUST 提交** `*_macro_gen.go`，下游直接 `go build`，不展开依赖源码
- **依赖**：可能依赖 `go/ast`、`go/parser`、`go/token`、`go/format` 及用于写回的 `golang.org/x/tools` 等
- **兼容性**：展开产物必须是标准 Go 源码，保证 `go test` / `go vet` 无额外运行时依赖；宏桩函数不得在生产路径被调用
