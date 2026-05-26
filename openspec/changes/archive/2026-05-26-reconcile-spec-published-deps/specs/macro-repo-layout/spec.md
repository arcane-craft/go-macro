## MODIFIED Requirements

### Requirement: 本地开发 workspace

`go-macro` 仓库根 **MAY** 提供 `go.work`；若提供，其 `use` **MUST** 仅包含根 module（`.`）与 `./examples`，**MUST NOT** `use` 已迁出的 `contrib` 或 sibling `go-macro-contrib` 路径。

仓库 **MUST NOT** 要求根目录必须存在 `go.work` 才能满足本规范。根 module 与 `examples` module 的测试 **MUST** 可按 module 边界分别执行（见下方场景）。

`examples/go.mod` 的**已提交**依赖 **MUST** 通过版本化 `require` 引用已发布的 `github.com/arcane-craft/go-macro` 与 `github.com/arcane-craft/go-macro-contrib` tag（与当前仓库 tag 策略兼容的版本号）。**MAY** 在已提交文件中保留 `replace github.com/arcane-craft/go-macro => ../`，以便在本仓开发核心时联调 examples。

对 `go-macro-contrib` 的本地并行开发，开发者 **MAY** 在本地（含未提交变更）向 `examples/go.mod` 添加 `replace github.com/arcane-craft/go-macro-contrib => ../go-macro-contrib`（contrib checkout 位于 `go-macro` 仓库根同级目录 `../go-macro-contrib`，路径相对 `examples/go.mod` 所在目录）。**MUST NOT** 要求该 contrib `replace` 必须出现在已提交的 `examples/go.mod` 中。

#### Scenario: 根 module 独立测试

- **WHEN** 于 `go-macro` 仓库根执行 `GOWORK=off go test ./...`
- **THEN** MUST 通过且仅覆盖根 module 包

#### Scenario: examples module 独立测试

- **WHEN** 于 `go-macro/examples` 目录执行 `go test ./...`，且 `examples/go.mod` 已 `require` 兼容的已发布 `go-macro-contrib` 版本（或开发者本地已添加 contrib `replace`）
- **THEN** MUST 通过（含 `readfile` golden 等）

#### Scenario: 可选 workspace 联调本仓两 module

- **WHEN** 开发者于 `go-macro` 根提供 `go.work`（`use` 为 `.` 与 `./examples`）并在根执行 `go test ./...`
- **THEN** MUST 能同时测试根 module 与 examples module
