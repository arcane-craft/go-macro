## 1. 大纲与内容迁移

- [x] 1.1 按 `design.md` 在 `README.md` 建立 L0–L3 章节空壳（阅读指引、快速上手、命令、文档、编辑器/gopls、参考）
- [x] 1.2 将现有「快速上手」五步与代码块迁入新结构，保持 MUST/RECOMMENDED 语义不变
- [x] 1.3 将命令表、gopls、文档链接迁入对应 L2 章节

## 2. 参考层与清理

- [x] 2.1 确认 README 已移除 `contrib/` BREAKING 迁移表；将官方宏库、模块路径、本地 `replace` / `GOWORK=off` 说明合并到 `## 参考`（可用 `###` 子标题）
- [x] 2.2 删除「术语澄清（无行为变更）」节；确认无 spec 过程类尾注残留
- [x] 2.3 补充「阅读指引」短节，指向快速上手、author-guide；历史路径迁移指向 CHANGELOG

## 3. 校验

- [x] 3.1 对照 `openspec/specs/macro-codegen/spec.md` 与 `macro-repo-layout/spec.md` 核对 README 必备句（gen 提交、expand 入口、generate 一行）
- [x] 3.2 通读全文：首屏至快速上手结束可独立完成首次 expand + build；全文无 BREAKING / 旧路径对照表
- [x] 3.3 确认 `docs/author-guide.md` 链接有效；章节 `##` 顺序符合 `project-readme` spec

## 4. Spec delta（本 change 内）

- [x] 4.1 更新 `specs/project-readme/spec.md`（快速上手、cmd/macro、gopls、叙述风格）
- [x] 4.2 新增 `specs/macro-codegen/spec.md`（MODIFIED：examples 入口定位、init provider、README 快速上手）
- [x] 4.3 新增 `specs/macro-core/spec.md`（MODIFIED：`go run cmd/macro@latest`）
- [x] 4.4 归档前运行 `openspec validate`（或等价校验）确认 delta 与主 spec 可合并
