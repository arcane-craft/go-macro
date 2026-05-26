## 1. 归档 spec delta

- [x] 1.1 将 `openspec/changes/reconcile-spec-published-deps/specs/macro-repo-layout/spec.md` 合并入 `openspec/specs/macro-repo-layout/spec.md`（替换「本地开发 workspace」整条 requirement）
- [x] 1.2 将 `openspec/changes/reconcile-spec-published-deps/specs/macro-contrib/spec.md` 合并入 `openspec/specs/macro-contrib/spec.md`（替换「contrib 依赖 go-macro 核心版本」整条 requirement）
- [x] 1.3 运行 `openspec validate`（或项目惯用校验）确认 delta 格式与场景标题层级正确

## 2. 对齐文档（无 Go 实现变更）

- [x] 2.1 更新 `go-macro/README.md`：根目录不强制 `go.work`；examples 以 `require` 消费已发布 contrib；本地 contrib `replace` 为可选说明
- [x] 2.2 更新 `go-macro/CHANGELOG.md`：修正「examples 已含 contrib replace」「根 go.work 仅根+examples」等与现仓不符的表述
- [x] 2.3 更新 `go-macro/docs/author-guide.md` 中本地联调段落，与 `macro-repo-layout` / `macro-contrib` 一致
- [x] 2.4 （可选）更新 `go-macro-contrib/README.md` 最低 `go-macro` 版本说明，与 `go.mod` pin（如 `v0.1.0`）一致

## 3. 验证

- [x] 3.1 对照现仓：`go-macro` 无根 `go.work`、`examples/go.mod` 为版本化 `require` + 可选 `replace go-macro => ../`，确认无新增 spec 冲突
- [x] 3.2 对照现仓：`go-macro-contrib/go.mod` 无已提交 `replace`，确认 `macro-contrib` 场景可描述当前状态
