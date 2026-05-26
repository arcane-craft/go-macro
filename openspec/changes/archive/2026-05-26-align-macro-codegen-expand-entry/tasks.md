## 1. 更新规范文本

- [x] 1.1 将 `macro-codegen` delta 合并至 `openspec/specs/macro-codegen/spec.md`（REMOVED/ADDED/MODIFIED）。
- [x] 1.2 将 `macro-contrib` delta 合并至 `openspec/specs/macro-contrib/spec.md`。
- [x] 1.3 校验所有新增/修改 requirement 含 `#### Scenario` 与 WHEN/THEN。

## 2. 文档抽查对齐

- [x] 2.1 检查 README / `docs/author-guide.md` 是否仍有「框架 macroexpand」「唯一官方入口」等残留措辞。
- [x] 2.2 检查 `cmd/macro` 脚手架生成的 README 模板是否与 RECOMMENDED generate 表述一致。

## 3. 验证与归档

- [x] 3.1 运行 `openspec status --change align-macro-codegen-expand-entry` 确认产物完整。
- [x] 3.2 确认无代码/API 变更需求后归档 change。
