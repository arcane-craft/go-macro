## Context

当前 `go-macro` 通过 `official_providers` 在 expander 内硬编码官方宏库。用户已定案：删除 `go tool macro expand`、breaking 迁 contrib、**expand 易用性由 macro 库与 examples 提供的 expand 入口承担，不转移给宏作者**。

Go 限制：`Expander` 须编译期 link；框架通过 **`expandtool.Register` + examples 内带 blank import 的 `cmd/macroexpand`** 解决，宏作者仅在被消费时提供 `register` 子包（`init` 内 Register）。

## Goals / Non-Goals

**Goals:**

- `macro/expandtool`：`Register`、`Run`、`Main`、`Registered`
- `examples/cmd/macroexpand`：官方 expand 二进制；用户 generate 一行、零项目 expand 文件
- `contrib/register`：注册 inline/try
- 根 module：`expander` 库与测试零依赖 contrib
- 三 module 布局：根 / contrib / examples；根 `go.work` 联调
- 宏作者脚手架生成 `register/register.go`，**不**生成 `tools/macroexpand`

**Non-Goals:**

- 不要求宏作者或宏使用方维护 `tools/macroexpand`
- 不把 `Main`/`Run` 放在 contrib 包（contrib 仅 `register` + 宏实现）
- 根 module `go.mod` require contrib（contrib 仅由 examples/contrib module 引用）
- 根路径兼容别名、plugin、全 module 自动 link

## Decisions

### D1：删除 `go tool macro expand`

`cmd/macro` 仅保留 `init provider`。expand 由 `go run .../examples/cmd/macroexpand` 承担。

### D2：contrib 子 module

```
contrib/
  go.mod
  inline/
  try/
  register/     # init → expandtool.Register(官方路径, Expand)
```

根 module 的 `expander`、`macro/expandtool` 及根测试 **不** import contrib。

### D3：`macro/expandtool`（核心库）

```go
func Register(importPath string, expand macro.Expander)
func Registered() map[string]macro.Expander
func Run(args []string, linked map[string]macro.Expander) error  // linked==nil → Registered()
func Main()  // Run(os.Args[1:], nil); 失败 os.Exit(1)
```

- `Run`：`args` 空则 `[]string{"./..."}`，调用 `expander.ExpandPackages`
- 宏作者 **不** 调用 `Main`；由 examples/cmd/macroexpand 或用户自建 cmd 使用

### D4：官方 expand 位于 examples module

```
examples/
  go.mod          # module github.com/arcane-craft/go-macro/examples
  cmd/macroexpand/main.go
  readfile/       # 示例；包路径 .../examples/readfile
```

```go
package main

import (
    _ "github.com/arcane-craft/go-macro/contrib/register"
    "github.com/arcane-craft/go-macro/macro/expandtool"
)

func main() { expandtool.Main() }
```

用户 generate（**canonical，零项目文件**）：

```go
//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .
```

**根 `go.mod` MUST NOT require contrib**；`contrib` 依赖仅出现在 `examples/go.mod`（及 `contrib/go.mod`）。集成测试（readfile golden、ExpandPackages 等）放在 **examples** module，根 module 测试用 stub，不 import contrib。

### D5：宏作者职责边界

| 角色 | 负责 | 不负责 |
|------|------|--------|
| 框架 | expandtool、examples/cmd/macroexpand、contrib/register | — |
| 宏作者（provider） | stubs、`Expand`、`//macro:`、`register/register.go`（脚手架） | expand main、tools/macroexpand、linked map |
| 宏使用方 | import 宏库、一行 generate | 默认无需 expand 代码 |

**消费自研/第三方宏库时**（附录）：使用方 MAY 复制 `examples/cmd/macroexpand` 为项目内 `cmd/macroexpand`，仅增加 blank import 宏库的 `register` 包并调用 `expandtool.Main()`；仍 **不** 手写 linked map。

### D6：`ExpandPackages` 与 expander 去耦

`ExpandPackages(patterns, linked map[string]macro.Expander)`；删除 `Provider`、official_providers。

### D7：Breaking 路径与 generate

| 旧 | 新 |
|----|-----|
| `.../go-macro/try` | `.../go-macro/contrib/try` |
| `go tool macro expand` | `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .` |
| 根 `cmd/macroexpand` | `examples/cmd/macroexpand` |

### D8：本地 `go.work`

仓库根提供 `go.work`，`use` 根、`./contrib`、`./examples`，便于同仓开发与 `go test ./...` 覆盖三 module。

### D9：其他

- 通用化 `imports.go`；修复 `load.go` append bug

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| generate 路径变长（含 `examples/`） | 文档与示例固定一行；可 `go install .../examples/cmd/macroexpand@version` |
| 第三方宏库需额外 blank import | 作者指南附录；作者提供 `register` 子包 |
| 多 module 克隆开发 | 根 `go.work`；各 module 独立 `go test` |

## Migration Plan

1. 实现 `macro/expandtool`、`examples/cmd/macroexpand`
2. 迁 contrib + `contrib/register`；examples 独立 module
3. 改 expander API，删 official_providers；根测试迁 examples
4. 删 `cmd/macro expand`；`init provider` 生成 `register/register.go`
5. 更新示例、文档与 specs（`macro-repo-layout`）

## Open Questions

- （已关闭）expand 入口放 contrib 还是 macro → **macro/expandtool + examples/cmd/macroexpand**
- （已关闭）用户 tools/macroexpand → **不需要**
- （已关闭）根 go.mod 是否 require contrib → **否；仅 examples module**
