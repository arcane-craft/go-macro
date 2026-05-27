## Context

go-macro 今日模型：

1. **元数据**：provider 包内至少一处 `//macro: <syntax-id>` + 扫所有「单语句 `panic`」包级函数为桩。
2. **链接**：`expandtool.Register(importPath, Expander)` 在 expand 进程 `init` 中执行（宏作者 `register/` 或 contrib 集中 `register`）。
3. **入口**：使用方自建 `cmd/macroexpand` 或 `go run .../examples/cmd/macroexpand`，blank import register 后 `expandtool.Main()`。

Go 要求 Expander 为**编译期可 link 的函数值**，无法在纯运行时仅靠 AST 反射调用未 import 的 `TryExpand`。因此「去掉 register」在实现上等于 **由框架生成** 含 blank import + `Register` 的 link 源码，而不是取消 link。

用户目标：桩与 Expander 均用 `//macro:` 标注；展开工具解析注释完成关联；使用方不写 expand 入口；宏作者不写 register。

## Goals / Non-Goals

**Goals:**

- 规定并实现 per-function `//macro:` 契约（桩 + Expander）。
- 注册表仅依赖注释解析 +（生成后的）`linked` Expander 函数指针。
- `cmd/macro expand` 为默认展开路径：发现依赖 → 生成/更新 link 文件 → 调用 `expandtool.Run`。
- 更新脚手架、文档、examples；contrib 去掉 hand-written register。

**Non-Goals:**

- 插件/so 动态加载 Expander。
- 去掉 `//go:build macro` 双文件模型。
- 在 expand 进程内完全消除 `expandtool.Register`（生成代码仍调用它）。
- 单仓库内实现 go-macro-contrib 全部代码修改（可列为 tasks 中的 follow-up / 并行 PR）。

## Decisions

### D1：per-function `//macro:` 语法

- 形式：紧邻函数 doc 的第一行或 doc 块内任一行 `//macro: <syntax-id>`（与现有 `parseMacroComment` 兼容）。
- **桩函数**：包级函数 doc 含 `//macro: syntax-X` → 该函数名为宏桩，syntax-id 为 `syntax-X`。
- **Expander**：函数 doc 含 `//macro: syntax-X`、签名为 `Expander` → 该 syntax-id 的展开实现；函数名记入 `syntaxToExpandFunc`。
- 同一 provider 包内，同一 `syntax-id` 可有多个桩、**恰好一个** Expander（多个 Expander 同 id → expand 失败）。

**备选**：保留包级 `//macro:` 作为默认 id → 拒绝，与用户「每桩标注」目标不一致。

### D2：桩识别不再依赖 panic 形态

- `RegisterProvider` 仅登记带 `//macro:` 的包级函数（仍要求无 receiver）。
- **RECOMMENDED** 桩体 `panic(...)`（运行时误用防护）；非 panic 桩允许编译通过，但 author-guide 警告。

**备选**：注释 + panic 双条件 → 更严但重复；首版采用注释唯一 MUST。

### D3：关联算法

对 import path `P` 的 provider 文件集：

```
stubs[syntax-id] += { stub names from stub func docs }
expanders[syntax-id] = exactly one func name from expander docs
registry: stubName -> syntax-id -> linked[P].(expanderFunc)
```

`linked[P]` 仍来自 `expandtool.Register(P, pkg.TryExpand)`，由生成代码提供。

宏主文件识别：`types.Info` 解析调用 → `(pkgPath, funcName)` → registry 查桩 doc 已登记。

### D4：自动生成 link 文件（取代手写 register / macroexpand）

- 路径：**`<module-root>/.gomacro/expand_link.go`**（`package gomacroexpand`，build tag `//go:build ignore` 或工具专用 tag，仅由 `go run cmd/macro expand` 编译进临时 main）。
- 内容模板：blank import 每个待 link 的 provider path；`init()` 内 `expandtool.Register(path, pkg.ExpanderName)`。
- **发现 provider**：对展开 patterns 加载的包图，取宏主文件 import 集合；对每个 path 用 `go/packages` 加载语法树，若存在带 `//macro:` 的 Expander 函数则纳入 link。
- `cmd/macro expand [patterns...]`：  
  1. 解析 module root（`go.mod`）  
  2. 计算需 link 的 paths（可先 dry-run 扫 import，或两遍：先扫 macro 主文件 import）  
  3. 若 `.gomacro/expand_link.go` 与所需集合不一致则重写  
  4. `go run` 合成 main：`.gomacro/expand_link.go` + `macro/expandtool` 的 thin main，或直接 `expandtool.Run` 在同进程（同进程需 link 文件在同一 module——**Decision D4b**）

**D4b（推荐实现）**：`cmd/macro expand` 自身 module 为 `github.com/arcane-craft/go-macro/cmd/macro`；生成 link 写入**用户 module** 的 `.gomacro/expand_link.go`，然后：

```text
go run -tags=gmacro_link github.com/arcane-craft/go-macro/cmd/macro expand
```

或在用户目录执行：

```text
go run ./.gomacro/expand_runner.go   // 生成 runner：import expandtool + link
```

更简单路径：**expand 子命令在用户 module 下生成 `expand_runner/main.go` 并 `go run` 该目录**（一次性目录，可 gitignore RECOMMENDED）。

**文档 RECOMMENDED**：将 `.gomacro/` 加入 `.gitignore`；CI 在 expand 前跑 `cmd/macro expand` 再生 gen 文件。可选 flag `--write-link` 提交 link 文件供离线 CI（默认生成不提交）。

**备选**：继续要求项目内 `cmd/macroexpand` 但由 `macro init expand` 生成 → 仍有一个文件要写 generate；用户明确要不写入口，故采用 `cmd/macro expand` 统一入口。

### D5：deprecate 手写 register

- `init provider` 不再生成 `register/`。
- `expandtool.Register` 保留，供生成代码与测试显式 linked map 使用。
- contrib `register` 包 **REMOVED**；官方库仅保留 per-function 注释。

### D6：examples 与 generate 一行

```go
//go:generate go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .
```

移除 `examples/cmd/macroexpand`，README 指向 `cmd/macro expand`。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 生成 link 与用户 go.mod 模块边界不一致 | 以 `go.mod` module path 为根；document monorepo 每 module 各跑 expand |
| 未提交 `.gomacro/` 时 CI 需网络拉 cmd/macro | generate 步骤固定；或 `--write-link` 提交 link |
| 去掉 panic 启发式后误标普通函数为桩 | 文档 + 可选 vet；桩 MUST 包级且具 `//macro:` |
| 同 stub 名跨 provider 冲突（既有问题） | 识别用 `(importPath, stubName)`；Lookup 改为二元键（本变更一并修） |
| go-macro-contrib 外仓同步 | tasks 单列；本仓 spec 先定契约 |

## Migration Plan

1. 实现新 registry 解析 + expand 子命令 + link 生成。
2. 更新 go-macro 内测试与 examples；author-guide / README。
3. 发布 go-macro 新版本；contrib PR：per-function 注释、删 register。
4. **BREAKING** 使用者：generate 改一行；删自建 `cmd/macroexpand` 与对 `contrib/register` 的 blank import；宏作者删 `register/`、给每个桩/Expander 加 `//macro:`。

回滚：保留旧 tag 文档；link 生成可关闭（不推荐长期）。

## Open Questions

- `.gomacro/` 默认 gitignore 还是提交 link 以利 CI 无 codegen？→ 首版 **gitignore + expand 在 CI generate**，与 `*_macro_gen.go` 提交策略对称讨论。
- `mactest.Expand` 是否改为从 provider 源码自动读 syntax-id？→ 可 follow-up，tasks 含可选项。
- expand 子命令是否合并 `expand` + 写 `*_macro_gen.go` 为单步（是，与现 Main 行为一致）。
