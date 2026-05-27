## REMOVED Requirements

### Requirement: Try 语法桩族

**Reason**: `syntax-try` 规范随官方 `try` 宏库迁至 `go-macro-contrib` 仓库。

**Migration**: 见 contrib 仓 `openspec/specs/syntax-try/spec.md`。

### Requirement: 可选官方宏库与引入方式

**Reason**: 同上。

**Migration**: 见 contrib 仓 `openspec/specs/syntax-try/spec.md`。

### Requirement: 多桩名注册到同一展开器

**Reason**: 同上。

**Migration**: 见 contrib 仓 `openspec/specs/syntax-try/spec.md`。

### Requirement: error 必须在返回列表最后

**Reason**: 同上。

**Migration**: 见 contrib 仓 `openspec/specs/syntax-try/spec.md`。

### Requirement: 外层函数必须含 error 返回

**Reason**: 同上。

**Migration**: 见 contrib 仓 `openspec/specs/syntax-try/spec.md`。

### Requirement: 内层 callee 与桩名一致性

**Reason**: 同上。

**Migration**: 见 contrib 仓 `openspec/specs/syntax-try/spec.md`。

### Requirement: Try 必须使用 Stmts（禁止 Exprs 简化 return）

**Reason**: 同上。

**Migration**: 见 contrib 仓 `openspec/specs/syntax-try/spec.md`。

### Requirement: Try 宏展开语义

**Reason**: 同上。

**Migration**: 见 contrib 仓 `openspec/specs/syntax-try/spec.md`。

### Requirement: 具名返回支持

**Reason**: 同上。

**Migration**: 见 contrib 仓 `openspec/specs/syntax-try/spec.md`。

### Requirement: 非法调用错误信息

**Reason**: 同上。

**Migration**: 见 contrib 仓 `openspec/specs/syntax-try/spec.md`。

### Requirement: Try 端到端示例

**Reason**: 该要求描述 `go-macro/examples` 集成，归属 core 仓 `macro-repo-layout`，不再留在 `syntax-try`。

**Migration**: 见 `go-macro` `openspec/specs/macro-repo-layout/spec.md` 中「readfile Try 端到端示例」要求（本 change 新增）。
