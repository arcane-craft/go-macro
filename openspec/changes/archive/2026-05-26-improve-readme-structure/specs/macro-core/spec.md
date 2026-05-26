## MODIFIED Requirements

### Requirement: init provider 脚手架

`github.com/arcane-craft/go-macro/cmd/macro` MUST 提供 `init provider` 子命令，生成**最小** provider 目录：含 `//macro:`、`Expand` 占位、**单个** panic 语法桩及 `expand_test.go`（mactest 模板）；MUST NOT 默认生成 Try 式多桩族模板。

用户文档 RECOMMENDED 通过 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>` 调用该子命令（`go tool macro` MAY 在已安装 tool 的环境下使用，但 MUST NOT 作为唯一文档入口）。

#### Scenario: 初始化新 provider

- **WHEN** 用户执行 `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider mymac`
- **THEN** MUST 创建可编译的 provider 骨架且文档指向作者指南
