## Context

当前 `go-macro` 为 monorepo：`contrib/`（module `github.com/arcane-craft/go-macro/contrib`）、`examples/`、根 module 通过根 `go.work` 联调。contrib 已满足「根 module 不 require contrib、expand 经 register link」等边界，但物理上仍在同一 Git 仓库。本次变更仅做**仓库与 module 路径**外迁，不改变 `expandtool.Register` 模型、`examples/cmd/macroexpand` 参考入口职责，也不改变 inline/try 宏语义。

## Goals / Non-Goals

**Goals:**

- 将 `contrib/` 完整迁入新仓库 `github.com/arcane-craft/go-macro-contrib`，作为该仓库唯一（根）Go module。
- 统一更新规范与代码中的 import path、`expandtool.Register` 键、文档与测试 fixture。
- `go-macro` 仓库仅保留根 module + `examples/`；`go.work` 与 CI 反映两 module 布局。
- 明确本地双仓联调方式（`replace`）与对外 `go get` 版本约束。

**Non-Goals:**

- 不修改 `macro/expandtool`、`internal/expander` 的 API 或展开算法。
- 不将 `examples/` 迁出 `go-macro`（仍作本仓示例与参考 macroexpand）。
- 不在本变更内实现 contrib 新仓的完整 CI/发布流水线细节（可留 follow-up），但须在 tasks 中列出最小验证步骤。
- 不提供旧路径的永久兼容 shim（无 `go-macro/contrib` 转发 module）。

## Decisions

### D1: 新 module 根路径为 `github.com/arcane-craft/go-macro-contrib`

**选择**：独立仓库根即 module 根；子包为 `.../inline`、`.../try`、`.../register`（不再嵌套 `contrib/` 目录段）。

**理由**：与 Go 社区常见「框架 + 官方扩展」分仓命名一致；缩短 import 路径；避免继续占用 `go-macro/contrib` 这一易与「子目录」混淆的路径。

**备选**：保留 `github.com/arcane-craft/go-macro/contrib` 作为 module 路径但仓库独立——Go module 路径不必与仓库 URL 一一对应，但会增加「路径含 contrib、仓库名不含」的认知成本，故不采用。

### D2: contrib 继续 `require` 已发布的 `go-macro` 核心

新仓 `go.mod`：

```go
module github.com/arcane-craft/go-macro-contrib

require github.com/arcane-craft/go-macro vX.Y.Z
```

本地 checkout 约定：与 `go-macro` 同级目录 **`../go-macro-contrib`**（即 `go-macro` 仓库根的上级目录下的 `go-macro-contrib`）。

在 `go-macro` 的 `examples/go.mod` 联调时使用：

```go
replace github.com/arcane-craft/go-macro-contrib => ../go-macro-contrib
```

在 `go-macro-contrib` 仓联调核心时使用：

```go
replace github.com/arcane-craft/go-macro => ../go-macro
```

**理由**：宏实现仍依赖 `macro` 包与 `expandtool`；依赖方向保持 contrib → core，core 不反向 require contrib；固定相邻目录便于文档与 CI 脚本一致。

### D3: `go-macro` 根 `go.work` 仅 `use` 根与 `./examples`

删除 `./contrib`。双仓联调在 `examples/go.mod` 中对 `go-macro-contrib` 使用 `replace => ../go-macro-contrib`（见 D2 目录约定）。

**理由**：contrib 已不在本仓；避免 `go.work` 指向不存在路径；`go.work` 不纳入 sibling 仓，由 `examples/go.mod` replace 覆盖本地开发。

### D4: `examples` 通过版本化 `require` 引用 contrib

`examples/go.mod` 改为：

```go
require github.com/arcane-craft/go-macro-contrib v0.x.x
```

并移除对 `../contrib` 的 replace；本地开发在 `examples/go.mod` 增加 `replace github.com/arcane-craft/go-macro-contrib => ../go-macro-contrib`（README 说明须将新仓 clone 至该相对路径）。

**理由**：`require` 展示对外消费方式；`replace` 固定相邻目录以支持本仓 `go test ./...` 联调。

### D5: 迁移方式：目录复制 + 路径批量替换，Git 历史可选保留

实现阶段：用 `git subtree split` 或在新仓 `git filter-repo` 保留 `contrib/` 历史均可；若工期紧，首版可采用「新仓初始提交 + CHANGELOG 迁移说明」，不阻塞路径切换。

**理由**：规范与代码正确性优先于历史连续性。

### D6: 同步更新 `expandtool.Register` 的 map 键

`register/register.go` 中注册的 import path 字符串 MUST 与新 module 路径一致，否则 expand 时 `linked` 与宏主文件 import 不匹配。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| **BREAKING** 所有使用旧路径的项目无法编译 | 在 `go-macro` 与 `go-macro-contrib` README/CHANGELOG 提供路径对照表与 `go get` 指引 |
| core/contrib 版本漂移导致 API 不兼容 | contrib `go.mod` 注明最低 `go-macro` 版本；发 tag 时在 release note 对齐 |
| 本仓 CI 不再一键 `go test ./...` 覆盖 contrib | `go-macro` CI 测根+examples；contrib CI 在新仓 `go test ./...`；文档说明双仓 clone |
| 开发者本地仅 clone `go-macro` 时 examples 缺 contrib | README 说明须将 `go-macro-contrib` clone 至 `../go-macro-contrib` 并启用 `examples/go.mod` 中的 replace，或 `go get` 已发布版本 |

## Migration Plan

1. 在 `../go-macro-contrib` 创建/初始化仓库，将 `contrib/*` 移至该仓根（调整 `go.mod` module 行与全部 import）。
2. 在新仓运行 `go test ./...`，确认通过。
3. 在 `go-macro` 删除 `contrib/`，更新 `examples`、`go.work`、文档、测试 fixture、`register` 引用处。
4. 在 `go-macro` 运行 `go test ./...`（workspace 下根+examples）。
5. 合并 OpenSpec delta 至 `openspec/specs/*`。
6. 对两仓分别打 BREAKING minor/major tag（按当前 semver 策略），发布迁移说明。

**回滚**：恢复 `contrib/` 目录与旧 import 路径；撤销 spec delta。已发布 tag 不回撤，以新 patch 修复为宜。

## Open Questions

- 首个 contrib 独立 tag 与 `go-macro` 最低兼容版本的数值（实现 tasks 时根据当前 API 敲定）。

## 已确认

- 独立仓库 Go module 路径：**`github.com/arcane-craft/go-macro-contrib`**（子包：`.../inline`、`.../try`、`.../register`）。
- 本地 checkout 路径（相对 `go-macro` 仓库根）：**`../go-macro-contrib`**。
