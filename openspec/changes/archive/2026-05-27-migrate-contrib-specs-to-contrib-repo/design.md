## Context

- **代码**：`go-macro-contrib` 已作为独立 module 发布（`inline`、`try`），`go-macro` 根 module 不再包含 `contrib/`。
- **规范**：`openspec/specs/macro-contrib`、`syntax-inline`、`syntax-try` 仍位于 `go-macro`，内容与 contrib 实现强绑定。
- **工具**：两仓均可使用同一 `openspec` CLI（spec-driven schema）；contrib 仓当前无 `openspec/` 目录。

## Goals / Non-Goals

**Goals:**

- 规范所有权与 Git 仓库边界一致：contrib 行为以 `go-macro-contrib/openspec/specs/` 为权威。
- `go-macro` 主规范集仅保留框架、仓库布局、examples 集成及对外引用 contrib 规范的边界条款。
- 迁移过程可归档：通过本 change 的 delta spec 明确 REMOVED/MODIFIED，避免静默丢失要求。

**Non-Goals:**

- 不改变 inline/try 的运行时语义或 import 路径（已在先前 change 完成）。
- 不在 `go-macro` 内用 `git submodule` 嵌入 contrib 的 spec 全文（仅允许链接/README 引用）。
- 不统一两仓为 monorepo OpenSpec workspace（首版保持两独立 `openspec` 根）。

## Decisions

### 1. 三份 spec 整包复制至 contrib，core 侧 REMOVED

**选择**：将 `macro-contrib`、`syntax-inline`、`syntax-try` 的 `spec.md` 原样（辅以 Purpose 补全）复制到 `go-macro-contrib/openspec/specs/`，在 `go-macro` 归档时删除对应主 spec。

**理由**：最小语义漂移；contrib 贡献者单仓即可查全要求。

**备选**：在 core 保留 stub spec 仅含链接 —— 拒绝，易造成双份维护与 archive 合并冲突。

### 2. readfile 端到端要求留在 core 的 `macro-repo-layout`

**选择**：将 `syntax-try` 中「Try 端到端示例 / ReadFile 黄金测试」迁至 `macro-repo-layout` 的 ADDED requirement。

**理由**：示例包路径属于 `go-macro/examples`；contrib 规范不应强制 core 仓目录结构。

### 3. 框架 spec 保留「边界场景」，删除「provider 语义」重复

**选择**：`macro-expander`、`macro-core` 保留 link/import、Site 矩阵、禁止根 module import contrib 等框架条款；删除或改写为「见 contrib 仓 `syntax-*`」的 Try/Inline 载荷与展开语义段落。

**理由**：框架 spec 描述引擎；provider spec 描述 TryExpand/InlineExpand。

### 4. contrib 仓 OpenSpec 初始化方式

**选择**：在 `go-macro-contrib` 根执行 `openspec init`（或与 `go-macro` 相同的 schema 配置），建立 `openspec/specs/<capability>/spec.md`；不复制 `go-macro` 的 `changes/archive` 历史。

**理由**：contrib 规范从零开始独立演进；历史留在 core archive。

### 5. 文档交叉引用格式

**选择**：`author-guide`、`project-readme` 的参考节使用稳定 URL：`https://github.com/arcane-craft/go-macro-contrib/tree/master/openspec/specs`（及 README），不在 core spec 内嵌 contrib requirement 全文。

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 双仓 PR 不同步导致短暂「core 已删、contrib 未增」 | 任务顺序：先 contrib 增 spec 并合并/打 tag，再 core 删 spec；或同一 PR 批次在本地 worktree 验证 |
| 外部链接断裂 | README 与 author-guide 同时更新；CI 可选 link check（非本 change 必须） |
| archive 后搜索旧 capability 名 | 在 `go-macro` README OpenSpec 节注明「官方宏库规范见 contrib 仓」 |
| `syntax-try` 移除后 examples 要求遗漏 | 显式 ADDED 至 `macro-repo-layout` delta |

## Migration Plan

1. **contrib 仓**：`openspec init` → 复制三份 spec → 补全 Purpose → README 增加「规范」小节 → `go test` 无变更。
2. **core 仓**：应用本 change（delta specs + 删主 spec 文件）→ `openspec validate` / archive change。
3. **验证**：在 core 执行 `openspec list` 确认无 `macro-contrib` 等主 spec；在 contrib 确认三 spec 存在且与代码一致。
4. **回滚**：恢复 core 三份 `spec.md` 并 revert contrib `openspec/` 目录（无运行时影响）。

## Open Questions

- contrib 仓是否引入与 core 相同的 `.cursor/skills/openspec-*`（可选，不阻塞迁移）。
- 是否在 `go-macro-contrib` CI 增加 `openspec validate`（建议 follow-up）。
