# go-macro

Go 过程宏（procedural macro）框架。

## 阅读指引


| 你是谁         | 从这里开始                         |
| ----------- | ----------------------------- |
| 在项目里**使用宏** | [快速上手](#快速上手) → [命令](#命令)     |
| **编写**宏库    | [宏作者指南](docs/author-guide.md) |


## 项目说明

你需要先在源码里写好宏调用，再在编译前运行展开工具；展开工具会把宏改写成普通 Go，并生成 `*_macro_gen.go`。之后你执行 `go build`、`go test` 时，编译器使用的就是这份生成代码。

宏库负责定义「能展开什么」，你的项目负责在构建流程里触发展开。使用宏的步骤见下方 [快速上手](#快速上手)；编写宏库请看 [宏作者指南](docs/author-guide.md)。

## 快速上手

### 1. 准备展开工具

你需要在自己项目的 `cmd/macroexpand` 里放置展开工具：它扫描源码，把宏调用改写成普通 Go。你能展开哪些宏，取决于你在编译展开工具时是否 blank import 了对应宏库的 `register` 包。

请在你的项目里新建 `cmd/macroexpand/main.go`。例如只使用官方宏库时，可以这样写：

```go
package main

import (
	_ "github.com/arcane-craft/go-macro-contrib/register" // 登记官方 inline、try 等
	"github.com/arcane-craft/go-macro/macro/expandtool"
)

func main() {
	expandtool.Main()
}
```

如果你还要用别的宏库，请在该文件中再增加一行 blank import，引入那个宏库自带的 `register` 包（自研宏库可用下方 `macro init provider` 命令生成骨架，其中会包含 `register`）。

**以 `examples` 项目为例**：本仓库的 `examples` 演示了一个宏使用方工程如何接入。

- 展开工具：[examples/cmd/macroexpand/main.go](examples/cmd/macroexpand/main.go)
- 使用宏的文件与 generate：[examples/readfile/readfile.go](examples/readfile/readfile.go)

你可以对照上述示例，在自己的项目里创建 `cmd/macroexpand`，并在使用宏的文件里写 `go run ./cmd/macroexpand`（或你模块下的等价路径）。

### 2. 在使用宏的文件里接上 generate

请在你打算写宏调用的源文件（例如 `foo.go`）顶部加上 build tag，并写上展开命令，例如：

```go
//go:build macro

//go:generate go run ./cmd/macroexpand .
```

### 3. 编写宏调用

请你在同一文件中 **import 你要用的宏库**（如 `go-macro-contrib/inline`、`go-macro-contrib/try`，或自研宏库），然后像普通 Go 一样写宏调用。

### 4. 运行展开

你可以用 `go generate` 触发展开，也可以直接运行展开工具，例如：

```bash
go generate ./...
# 或直接运行展开工具（路径与第 1 步一致）
go run ./cmd/macroexpand ./...
```

### 5. 提交与日常构建

请提交 `foo.go` 与生成的 `foo_macro_gen.go`（如果项目会被别人依赖，建议把生成文件一并提交）。

日常编译、测试时，你可以直接执行：

```bash
go build ./...
go test ./...
```

## 命令


| 命令                                    | 说明                                  |
| ------------------------------------- | ----------------------------------- |
| `go run ./cmd/macroexpand [packages]` | 在你自己的项目里运行展开工具（需先按快速上手第 1 步创建）      |
| `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>` | 编写宏库时：生成 provider 骨架（含 `register/`） |


对照示例：[examples/cmd/macroexpand](examples/cmd/macroexpand/main.go)、[examples/readfile](examples/readfile/readfile.go)。

## 文档

如需更详细的说明（ignore/tag、provider 实现等），请参阅 [宏作者指南](docs/author-guide.md)。示例代码见 `examples/readfile/`。

## gopls

使用宏的源文件通常带有 `//go:build macro` 标记；默认情况下 gopls 不会启用这个 tag，编辑器里可能看不到这些文件、补全和类型检查也会不对。你在编辑这类文件时，可以在设置里加上 `-tags=macro`，让 gopls 按「宏版本」的源码来分析：

```json
"gopls": { "buildFlags": ["-tags=macro"] }
```

## 参考

### 官方宏库（可选）

官方宏库由独立仓库 [go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) 维护。你需要在源码里 import 对应包，并在展开工具里登记其 `register`，然后才能展开这些宏。

- `github.com/arcane-craft/go-macro-contrib/inline` — 表达式宏
- `github.com/arcane-craft/go-macro-contrib/try` — `Try` 族错误处理宏

### 模块路径

```
github.com/arcane-craft/go-macro          # 本仓库：框架核心
github.com/arcane-craft/go-macro/examples # 示例调用方（含参考 macroexpand）
github.com/arcane-craft/go-macro-contrib  # 独立仓库：官方宏库
```

你需要把宏展开入口放在**使用宏的项目**里。本仓库的 `examples` 是示范用的调用方工程，你可以对照其中的 `examples/cmd/macroexpand` 复制到自己的项目。

### 本地联调

- **contrib**：你可以将 `go-macro-contrib` clone 到与 `go-macro` 同级目录 `../go-macro-contrib`。`examples` 模块通过 `examples/go.mod` 里的 `require` 引用已发布的 contrib；你若要同时改 contrib，可在本地添加 `replace github.com/arcane-craft/go-macro-contrib => ../go-macro-contrib`（一般只留在本机，不必提交）。
- **测试**：你可以在仓库根目录执行 `GOWORK=off go test ./...` 测试核心模块；在 `examples/` 目录下执行 `go test ./...` 测试示例模块。

