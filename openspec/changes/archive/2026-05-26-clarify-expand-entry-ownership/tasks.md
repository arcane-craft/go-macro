## 1. 更新规范文本

- [x] 1.1 在 `macro-repo-layout` delta 中将 `examples/cmd/macroexpand` 的角色改为“参考实现”，并补充“调用方项目承载入口”的 MUST 约束。
- [x] 1.2 在 `macro-core` delta 中更新 `expandtool` requirement，明确框架提供能力、调用方承载可执行入口、examples 非唯一路径。
- [x] 1.3 校验两个 delta spec 的场景均使用 `#### Scenario` 且满足 WHEN/THEN 可测试性。

## 2. 对齐文档措辞

- [x] 2.1 更新 README 中“官方/通用/唯一 macroexpand”相关措辞，区分 MUST 与 RECOMMENDED 语义。
- [x] 2.2 更新 `docs/author-guide.md` 中入口职责描述，明确 examples 为示例调用方项目。
- [x] 2.3 添加或更新迁移说明，标注本次为“语义澄清，无行为变更”。

## 3. 验证与收尾

- [x] 3.1 运行 `openspec status --change clarify-expand-entry-ownership`，确认 proposal/design/specs/tasks 全部可用。
- [x] 3.2 在评审备注中记录“为何不再将 examples 路径当作唯一入口”的设计决策与边界。
- [x] 3.3 准备进入实现阶段（`/opsx:apply`），按任务清单逐项落地文档改动。
