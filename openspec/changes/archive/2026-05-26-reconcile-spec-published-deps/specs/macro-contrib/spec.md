## MODIFIED Requirements

### Requirement: contrib 依赖 go-macro 核心版本

`go-macro-contrib` 的**已提交** `go.mod` **MUST** `require` 已发布的 `github.com/arcane-craft/go-macro` 版本（semver tag，非 `v0.0.0` 占位）；**MUST NOT** 在已提交 `go.mod` 中包含指向 sibling 目录的 `replace` 指令。

README **MUST** 注明最低兼容核心版本，且所述版本 **MUST** 与 `go.mod` 中 pin 的 `require` 一致。

本地联调时，contrib 仓库 **SHOULD** 与 `go-macro` 位于同级目录（`go-macro-contrib` 与 `go-macro` 并列）。开发者 **MAY** 在本地向 `go-macro-contrib/go.mod` 添加 `replace github.com/arcane-craft/go-macro => ../go-macro` 以联调未发布的核心变更；该 `replace` **MUST NOT** 作为发布/tag 前提交态的硬性要求。

#### Scenario: 独立仓可解析核心依赖

- **WHEN** 在仅 clone `go-macro-contrib`、已提交 `go.mod` 无 `replace`，且模块代理可解析所 pin 的 `go-macro` tag 时执行 `go test ./...`
- **THEN** MUST 成功解析 `github.com/arcane-craft/go-macro` 模块依赖并通过测试

#### Scenario: 本地 replace 联调核心

- **WHEN** 开发者在 `go-macro-contrib` 本地添加 `replace github.com/arcane-craft/go-macro => ../go-macro` 后执行 `go test ./...`
- **THEN** MUST 使用 sibling 核心源码解析依赖（联调行为；不要求写入已提交 `go.mod`）
