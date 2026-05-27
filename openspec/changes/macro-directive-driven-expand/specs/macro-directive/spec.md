## ADDED Requirements

### Requirement: macro 注释指令格式

`//macro: <syntax-id>` MUST 为单行注释，`<syntax-id>` MUST 为非空、由字母数字与连字符组成的标识符（与现有 `syntax-inline`、`syntax-try` 形态一致）。解析器 MUST 忽略 `//macro:` 后的多余空白。

#### Scenario: 解析标准指令

- **WHEN** 函数 doc 含 `//macro: syntax-mine`
- **THEN** 解析得到的 syntax-id MUST 为 `syntax-mine`

### Requirement: 桩函数上的 macro 指令

语法桩 MUST 为包级函数（无 receiver）。每个语法桩函数的 doc 注释 MUST 包含恰好一个 `//macro: <syntax-id>`，声明该函数属于该 syntax-id 的宏桩。

#### Scenario: 带指令的桩被登记

- **WHEN** provider 包中 `func Try[T any](v T) T` 的 doc 含 `//macro: syntax-try`
- **THEN** 注册表 MUST 将 `Try` 登记为 `syntax-try` 的桩

#### Scenario: 无指令的包级函数不是桩

- **WHEN** provider 包中存在包级函数 `Helper`，其 doc 不含 `//macro:`
- **THEN** 注册表 MUST NOT 将 `Helper` 登记为宏桩

### Requirement: Expander 函数上的 macro 指令

每个 syntax-id 在单个 provider 包内 MUST 有且仅有**一个** Expander 函数：包级函数、签名为 `func(Context, *ast.CallExpr) (ExpandResult, error)`，且 doc 含 `//macro: <syntax-id>`（与对应桩相同的 syntax-id）。

#### Scenario: 单 Expander 绑定 syntax-id

- **WHEN** `TryExpand` 的 doc 含 `//macro: syntax-try`，且包内无其它函数 doc 含 `//macro: syntax-try` 且符合 Expander 签名
- **THEN** 注册表 MUST 将 `syntax-try` 的展开实现解析为 `TryExpand`

#### Scenario: 重复 syntax-id 的 Expander

- **WHEN** 同一 provider 包内两个不同函数 doc 均含 `//macro: syntax-try` 且均符合 Expander 签名
- **THEN** expand MUST 失败并说明 syntax-id 冲突

### Requirement: 桩与 Expander 通过 syntax-id 关联

对给定 import path，注册表 MUST 将所有带 `//macro: syntax-X` 的桩名映射到同一 syntax-id，并将该 syntax-id 映射到该包内标注 `//macro: syntax-X` 的 Expander 函数（经 `linked` 提供的函数指针）。

#### Scenario: 多桩共享 Expander

- **WHEN** `Try` 与 `Try2` 的 doc 均为 `//macro: syntax-try`，且 `TryExpand` 的 doc 为 `//macro: syntax-try`
- **THEN** `Try` 与 `Try2` 的调用 MUST 均分发到 `TryExpand`

#### Scenario: 桩与 Expander syntax-id 不一致

- **WHEN** 桩 `Foo` 标注 `//macro: syntax-a`，唯一 Expander 标注 `//macro: syntax-b`
- **THEN** expand MUST 失败并说明 `syntax-a` 缺少 Expander 或 id 不一致

### Requirement: 禁止包级兜底指令替代 per-function 指令

系统 MUST NOT 将「文件中任意位置的单一 `//macro:`」作为该包所有桩的隐式 syntax-id。包级或文件级注释中的 `//macro:` MUST NOT 单独满足桩或 Expander 的登记要求。

#### Scenario: 仅包注释不登记桩

- **WHEN** provider 文件仅有包注释 `//macro: syntax-x`，函数 `Macro` 无 doc 指令
- **THEN** `Macro` MUST NOT 被登记为宏桩
