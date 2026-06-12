## MODIFIED Requirements

### Requirement: 生成代码行号

写入 `*_macro_gen.go` 时，引擎 MUST 在生成语句块中使用 `//line` 指向宏主文件。Call 宏行号 MUST 取自 **`site.MacroPos()`**（不再取自 `CallContext.MacroPos()`）。Decl 宏行号 MUST 取自 embed 处 **`site.MacroPos()`**（或 ResolveSite 记录的 embed 位置）。

#### Scenario: Call 宏 line 指令

- **WHEN** expand 替换 `Try(...)` 为多条 stmt
- **THEN** 生成 stmt MUST 带 `//line` 指向原宏主文件 macro 调用行

#### Scenario: 与 StampStmtPos 一致

- **WHEN** Apply 完成后 StampStmtPos
- **THEN** MUST 使用 `site.MacroPos()` 作为 stamp 输入
