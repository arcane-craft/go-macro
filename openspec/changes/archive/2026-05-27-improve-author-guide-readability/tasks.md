## 1. 大纲与结构

- [x] 1.1 按 `author-guide` spec 在 `docs/author-guide.md` 建立新 `##` 骨架：阅读指引 → 角色分工 → 编写宏库 → 宏使用方 → 参考
- [x] 1.2 在 `编写宏库` 下添加 `###` 子节：框架契约、调用语境（Site）、init provider、纯 Expand 单测
- [x] 1.3 在 `宏使用方` 下添加 `###` 子节：方案 C（build tag / gen）、expand 入口、发布 checklist
- [x] 1.4 在 `参考` 下添加 `###` 子节：官方宏库、本地联调、消费第三方宏库、Try 桩族附录

## 2. 内容迁移与改写

- [x] 2.1 迁移「角色分工」表与 examples 说明；删除「术语澄清（无行为变更）」节，必要语义并入角色分工或 expand 入口
- [x] 2.2 改写「框架契约」与 Site 表为自然中文，保留 provider 签名、ExpandResult、EnclosingFunc 等 normative 事实
- [x] 2.3 迁移 `init provider` 命令与 mactest 代码示例至 `编写宏库` 对应子节
- [x] 2.4 迁移方案 C、expand 入口、发布 checklist 至 `宏使用方`；使用方 expand 长说明精简并链至 README 快速上手
- [x] 2.5 拆分官方 contrib 长段：主节保留模块路径与 syntax-id；`replace` / `go.work` / 测试命令移入 `### 本地联调`
- [x] 2.6 迁移第三方 register 附录与 Try 桩族表至 `参考` 层

## 3. 阅读指引与互链

- [x] 3.1 新增「阅读指引」表格：宏库作者 → 编写宏库；宏使用方 → README 或宏使用方节
- [x] 3.2 确认 README `文档` 节链至 author-guide 有效；author-guide 链回 README 快速上手

## 4. 校验

- [x] 4.1 对照 `macro-core`、`macro-codegen`、`macro-repo-layout` 核对 normative 要点无遗漏
- [x] 4.2 确认 `##` 顺序符合 `author-guide` spec（编写宏库在参考节之前；无 meta 尾注）
- [x] 4.3 通读全文：语气为使用指引而非 spec/plan 文体
