# go-macro

Go 过程宏（procedural macro）框架：宏作者在 provider 包中定义语法桩与 `Expand`，工具链在构建前将宏主文件中的调用改写为合法 Go，并生成 `*_macro_gen.go` 供默认 `go build` 使用。

## 快速上手

1. 在宏主文件 `foo.go` 顶部添加 build tag 与 generate：

```go
//go:build macro

//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .
```

2. 在宏主文件中 **import 你要用的宏库**（如官方库 `contrib/inline`、`contrib/try`，或自研 provider）并编写宏调用。未 import 的宏库不会参与展开。

3. 展开本模块（与 generate 等价）：

```bash
go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand ./...
```

无需在项目内维护 `tools/macroexpand`。

4. 提交 `foo.go` 与 `foo_macro_gen.go`（**对外库 MUST 提交 gen**）。

5. 日常构建无需 `-tags macro`：

```bash
go build ./...
go test ./...
```

## 命令

| 命令 | 说明 |
|------|------|
| `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand [packages]` | 展开当前模块内宏主文件（默认 `./...`）；示例见 `examples/` module |
| `go tool macro init provider <name>` | 创建最小 provider 骨架（含 `register/`） |

## gopls

编辑宏主文件时建议：

```json
"gopls": { "buildFlags": ["-tags=macro"] }
```

## 文档

- [宏作者指南](docs/author-guide.md)
- 示例：`examples/readfile/`

## 官方宏库（可选依赖）

`contrib` 子 module 维护，**由你在宏主文件中 import 后才会展开**；框架 expand 二进制通过 `contrib/register` 在编译期 link Expander。

- `contrib/inline/` — 表达式宏
- `contrib/try/` — `Try` 族错误处理宏

## 模块路径

```
github.com/arcane-craft/go-macro
github.com/arcane-craft/go-macro/contrib
github.com/arcane-craft/go-macro/examples
```

官方 expand 入口位于 `examples/cmd/macroexpand`（`examples` 独立 module，含 contrib link）。
