## MODIFIED Requirements

### Requirement: 与作者指南的职责边界

README MUST 链至 `docs/author-guide.md` 作为详细说明入口。README MUST NOT 复制 author-guide 级别的长契约正文（如完整 provider 实现说明、宏位置与返回字段表、主文件/生成文件 build tag 细节）。

`docs/author-guide.md` MUST 遵循 `author-guide` spec 的信息架构（含 `阅读指引`、角色分工、编写/使用方分节），并在 `阅读指引` 或首段链回 README 供宏使用方跳转。

#### Scenario: 深度文档跳转

- **WHEN** 读者需要 ignore/tag 或 provider 实现细节
- **THEN** README `文档` 节 MUST 提供指向 `docs/author-guide.md` 的链接

#### Scenario: 使用方从 author-guide 回到 README

- **WHEN** 仅想在使用方项目接 expand 的读者打开 author-guide
- **THEN** MUST 能在 `阅读指引` 或文档首段找到指向 README 快速上手的链接，且 MUST NOT 必须先阅读完整 provider 契约
