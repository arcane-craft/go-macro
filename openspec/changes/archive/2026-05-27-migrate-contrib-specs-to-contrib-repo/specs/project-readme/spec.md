## MODIFIED Requirements

### Requirement: 参考内容与主路径分离

contrib 本地 `replace` 说明、双 module 测试命令（如 `GOWORK=off`）MUST 位于 `快速上手` 之后的参考类章节（`## 参考` 或其子标题 `###`），MUST NOT 插入快速上手编号步骤中间。

官方宏库（`go-macro-contrib`）的详细规范与版本兼容说明 MUST 链至 contrib 仓库 README 与 `openspec/specs/`，README MUST NOT 暗示 contrib 的 OpenSpec 仍位于本仓库。

#### Scenario: 主路径无参考信息打断

- **WHEN** 读者执行 `快速上手` 中的步骤
- **THEN** MUST 能在连续阅读主路径步骤的过程中完成「准备 expand → 接 generate → 运行展开」，中间 MUST NOT 插入非操作性的长篇参考段落

#### Scenario: 官方宏库文档外链

- **WHEN** 读者在 README `参考` 节查找 inline/try 行为说明
- **THEN** MUST 能找到 `go-macro-contrib` 仓库文档链接，且 MUST NOT 仅引用本仓已移除的 `openspec/specs/syntax-try` 路径
