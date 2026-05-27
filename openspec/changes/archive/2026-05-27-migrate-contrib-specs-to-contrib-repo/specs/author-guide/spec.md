## MODIFIED Requirements

### Requirement: 参考内容与主路径分离

contrib 模块路径、本地 `replace` / `go.work`、双 module 测试命令（如 `GOWORK=off`）MUST 位于 `编写宏库` 与 `宏使用方` 之后的参考类章节（`## 参考` 或其 `###` 子标题），MUST NOT 插入 `编写宏库` 或 `宏使用方` 的编号/步骤中间。

`参考` 节中关于官方宏库（inline/try）的 normative 行为 MUST 链至 `go-macro-contrib` 仓库 README 与 `openspec/specs/`，MUST NOT 假定读者在 `go-macro` 仓内可找到 `macro-contrib` / `syntax-*` 主 spec。

#### Scenario: 主路径无参考信息打断

- **WHEN** 读者连续阅读 `编写宏库` 各子节（含 `init provider` 与 mactest）
- **THEN** 中间 MUST NOT 插入非操作性的长篇本地联调段落

#### Scenario: 官方宏库规范外链

- **WHEN** 读者在 `参考` 节查找 Try/Inline 展开语义或 contrib 发布约定
- **THEN** MUST 能找到指向 `github.com/arcane-craft/go-macro-contrib` 文档或 OpenSpec 的链接，且 MUST NOT 仅指向已删除的 `go-macro/openspec/specs/syntax-try` 路径
