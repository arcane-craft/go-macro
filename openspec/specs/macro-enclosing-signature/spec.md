# macro-enclosing-signature Specification

## Purpose
TBD - created by archiving change macro-syntax-rules.
## Requirements
### Requirement: EnclosingSignature 通用 API

系统 MUST 提供 `EnclosingSignature(ctx Context, site Syntax) (*types.Signature, error)` 与 `EnclosingResults(ctx Context, site Syntax) (*types.Tuple, error)`。实现 MUST 使用 `ctx.Types()` 与 `site.MacroPos()`，通过 **internal** 在 `*ast.File` 上定位 enclosing `*types.Func`。公开 API MUST NOT 返回 `*ast.FuncDecl` 或 `EnclosingFunc()` 方法。

#### Scenario: return 路径获取 Results

- **WHEN** site 位于某函数体内且该函数 `return (int, error)`
- **THEN** `EnclosingResults` MUST 返回长度为 2 的 Results，末位为 error 类型（若源码如此）

#### Scenario: 不在函数体内

- **WHEN** site 无 enclosing function
- **THEN** `EnclosingSignature` MUST 返回 error

### Requirement: ZeroSyntax

系统 MUST 提供 `ZeroSyntax(ctx Context, typ types.Type) (Syntax, error)`，为任意 `types.Type` 生成零值 `Syntax`（供 Quote 组合）。此 API MUST NOT 为单一 syntax-id 特化。

#### Scenario: int 零值

- **WHEN** 调用 `ZeroSyntax(ctx, types.Typ[types.Int])`
- **THEN** MUST 返回等价于 literal `0` 的 Syntax

### Requirement: provider 自行组合 error return

框架 MUST NOT 提供 Try 专用的 `ErrorReturn(site)` 或自动 `#zeros` 注入。Try 等宏 MUST 在 Transform 内使用 `EnclosingResults` + `ZeroSyntax` 生成 `if err != nil { return ... }` 内容。

#### Scenario: assign 路径仍需完整 Results

- **WHEN** 函数签名为 `(int, string, error)` 且 expand assign 形态 Try
- **THEN** provider MUST 使用 `EnclosingResults` 生成 error 分支 return 三值，不得仅依据 assign lhs 两个名字
