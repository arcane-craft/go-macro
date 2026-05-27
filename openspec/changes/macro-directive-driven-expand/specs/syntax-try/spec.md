## MODIFIED Requirements

### Requirement: Try 语法桩族

`try` 包 MUST 提供按「error 前载荷个数」划分的多个语法桩，均 SHOULD panic，且**每个桩函数**的 doc MUST 含 `//macro: syntax-try`。`TryExpand` 的 doc MUST 含 `//macro: syntax-try`。不得仅提供单一 `func Try[T any](T, error) T` 作为唯一桩。

| 桩名 | 签名（概念） | callee 载荷数 k |
|------|--------------|-----------------|
| `Try0` | `(error)` → 无值 | k=0 |
| `Try` | `(T, error) → T` | k=1 |
| `Try2` | `(A, B, error) → (A, B)` | k=2 |
| `Try3` | `(A, B, C, error) → (A, B, C)` | k=3 |
| `Try4` | 四载荷 + error → 四元组 | k=4（可选） |

#### Scenario: Try 适配一元载荷

- **WHEN** 用户编写 `Try(os.Open(...))` 且 `os.Open` 类型为 `( *os.File, error)`
- **THEN** `Try` 桩 MUST 使 macro 源文件通过类型检查，且展开器 MUST 接受该调用

#### Scenario: 多载荷须用 Try2

- **WHEN** callee 类型为 `(A, B, error)` 且用户编写 `Try2(f())`
- **THEN** 桩 MUST 为 `(A, B, error) → (A, B)` 形式，且 macro 源文件 MUST 通过类型检查

#### Scenario: 错用 Try 处理多载荷

- **WHEN** callee 类型为 `(A, B, error)` 但用户编写 `Try(f())`（两载荷却用 `Try` 桩）
- **THEN** 展开器 MUST 在展开阶段返回错误（类型检查可能已通过）
