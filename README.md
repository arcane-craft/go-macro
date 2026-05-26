# go-macro

Go 过程宏（procedural macro）框架：宏作者在 provider 包中定义语法桩与 `Expand`，工具链在构建前将宏主文件中的调用改写为合法 Go，并生成 `*_macro_gen.go` 供默认 `go build` 使用。

## 快速上手

1. 在宏主文件 `foo.go` 顶部添加 build tag 与 generate：

```go
//go:build macro

//go:generate go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand .
```

2. 在宏主文件中 **import 你要用的宏库**（如官方库 `go-macro-contrib/inline`、`go-macro-contrib/try`，或自研 provider）并编写宏调用。未 import 的宏库不会参与展开。

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
| `go run github.com/arcane-craft/go-macro/examples/cmd/macroexpand [packages]` | **推荐**：使用 examples 参考入口展开当前模块（默认 `./...`）；亦可在项目内自建等价 `cmd/macroexpand` |
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

独立仓库 [go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) 维护，**由你在宏主文件中 import 后才会展开**；框架 expand 二进制通过 `go-macro-contrib/register` 在编译期 link Expander。

- `github.com/arcane-craft/go-macro-contrib/inline` — 表达式宏
- `github.com/arcane-craft/go-macro-contrib/try` — `Try` 族错误处理宏

本地开发：将 `go-macro-contrib` clone 到与 `go-macro` 同级目录 `../go-macro-contrib`；`examples/go.mod` 已含 `replace` 指向该路径。

## 模块路径

```
github.com/arcane-craft/go-macro
github.com/arcane-craft/go-macro/examples
github.com/arcane-craft/go-macro-contrib   # 独立仓库
```

### BREAKING：自 `go-macro/contrib` 迁移

| 旧路径 | 新路径 |
|--------|--------|
| `github.com/arcane-craft/go-macro/contrib/inline` | `github.com/arcane-craft/go-macro-contrib/inline` |
| `github.com/arcane-craft/go-macro/contrib/try` | `github.com/arcane-craft/go-macro-contrib/try` |
| `github.com/arcane-craft/go-macro/contrib/register` | `github.com/arcane-craft/go-macro-contrib/register` |

```bash
go get github.com/arcane-craft/go-macro-contrib@v0.1.0
```

宏展开可执行入口由**宏调用方项目**承载（`register` + `expandtool.Main()`）。本仓库 `examples` 为示例调用方工程；`examples/cmd/macroexpand` 为**推荐参考实现**（RECOMMENDED），非唯一路径。

## 术语澄清（无行为变更）

规范要求调用方提供等价 expand 接线能力；`go run .../examples/cmd/macroexpand` 仍为文档推荐默认命令，API 与目录结构不变。
