## 1. go-macro-contrib 仓 OpenSpec 初始化

- [x] 1.1 在 `go-macro-contrib` 根目录执行 `openspec init`（或与 `go-macro` 一致的 spec-driven 配置）
- [x] 1.2 创建 `openspec/specs/macro-contrib/spec.md`，自 `go-macro/openspec/specs/macro-contrib/spec.md` 复制并补全 Purpose
- [x] 1.3 创建 `openspec/specs/syntax-inline/spec.md`，自 `go-macro/openspec/specs/syntax-inline/spec.md` 复制并补全 Purpose
- [x] 1.4 创建 `openspec/specs/syntax-try/spec.md`，自 `go-macro/openspec/specs/syntax-try/spec.md` 复制（不含 readfile 端到端要求）并补全 Purpose
- [x] 1.5 更新 `go-macro-contrib/README.md`：增加「规范 / OpenSpec」小节，链至 `openspec/specs/`

## 2. go-macro 主规范迁移与归档

- [x] 2.1 应用本 change：合并 delta 至 `openspec/specs/`（`macro-repo-layout`、`macro-expander`、`macro-core`、`macro-codegen`、`author-guide`、`project-readme`）
- [x] 2.2 删除 `openspec/specs/macro-contrib/`、`syntax-inline/`、`syntax-try/` 主 spec 目录
- [x] 2.3 运行 `openspec validate`（或项目等价检查）确认 core 主 spec 无悬空引用
- [x] 2.4 归档 change：`openspec archive migrate-contrib-specs-to-contrib-repo`（或项目既定 archive 流程）

## 3. 文档与交叉引用

- [x] 3.1 更新 `docs/author-guide.md`：`参考` 节链至 contrib 仓 README / OpenSpec，移除对已迁出 core spec 路径的暗示
- [x] 3.2 更新根 `README.md`：`参考` 节注明官方宏库规范在 `go-macro-contrib`
- [x] 3.3 若 `README` / author-guide 提及 OpenSpec 目录结构，对齐为「core 框架 spec + contrib 官方宏库 spec」双仓说明

## 4. 验证

- [x] 4.1 在 `go-macro`：`openspec list` 确认无 `macro-contrib`、`syntax-inline`、`syntax-try`
- [x] 4.2 在 `go-macro-contrib`：确认三份 spec 存在且与 `inline/`、`try/` 实现一致
- [x] 4.3 在 `go-macro`：`GOWORK=off go test ./...` 与 `cd examples && go test ./...` 仍通过（无运行时变更）
