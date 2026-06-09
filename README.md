# go-macro

Go 宏框架。

## 阅读指引


| 你是谁         | 从这里开始                         |
| ----------- | ----------------------------- |
| 在项目里**使用宏** | [快速上手](#快速上手)                 |
| **编写**宏库    | [宏作者指南](docs/author-guide.md) |
| 查官方宏库、本地联调  | [参考](#参考)                     |


## 项目说明

你在源码里写宏调用；编译前运行展开工具，它会把宏改写成普通 Go，并生成 `foo_macro_gen.go`。之后日常 `go build`、`go test` 走这份生成代码，通常不需要长期加 `-tags macro`。

```
宏主文件 foo.go          expand 工具              生成文件 foo_macro_gen.go
(//go:build macro)   →   cmd/macro expand    →   (//go:build !macro)
     你手写宏调用              扫描 import、展开              日常构建用这份
```

- 若你**编写宏库**，你需要定义「能展开什么」——详见 [宏作者指南](docs/author-guide.md)
- 若你**使用宏**，你在宏主文件 import 宏库，用 `go:generate` 触发展开——见下方 [快速上手](#快速上手)

## 快速上手

下面按第一次接入的顺序说明。完整对照示例见 [examples/readfile/readfile.go](examples/readfile/readfile.go)。

### 1. 在宏主文件接上 generate

在要使用宏的源文件顶部加上 build tag 和展开命令：

```go
//go:build macro

//go:generate go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .
```

运行 `expand` 时，工具会扫描该文件所在包的 import，自动链接已 import 的宏库，并写回 `*_macro_gen.go`。它还会在模块根目录生成 `.gomacro/expand_runner/`（建议加入 `.gitignore`）。

### 2. import 宏库并编写调用

在同一文件中 import 你要用的宏库，然后像普通 Go 一样写宏调用：

```go
import "github.com/arcane-craft/go-macro-contrib/try"

func Example() error {
    f := try.Try(os.Open("hello.txt"))
    _ = f
    return nil
}
```

### 3. 运行展开

在**宏主文件所在包**执行 generate（`go generate` 的工作目录即该包）：

```bash
go generate .
```

你也可以直接调用 expand（整模块用 `./...`）：

```bash
go run github.com/arcane-craft/go-macro/cmd/macro@latest expand .
```

### 4. 提交并日常构建

若项目会被他人 `require`，建议把 `foo.go` 与 `foo_macro_gen.go` 一并提交。之后日常开发：

```bash
go build ./...
go test ./...
```

## 命令


| 命令                                                                              | 说明                          |
| ------------------------------------------------------------------------------- | --------------------------- |
| `go run github.com/arcane-craft/go-macro/cmd/macro@latest expand [patterns]`    | 展开宏；省略 patterns 时默认 `./...` |
| `go run github.com/arcane-craft/go-macro/cmd/macro@latest init provider <name>` | 生成宏库 provider 骨架（写宏库时用）     |


## 文档

- [宏作者指南](docs/author-guide.md) — provider 契约、单测、使用方细节
- [examples/readfile](examples/readfile/) — 可运行的接入示例

## gopls

带宏的源文件通常有 `//go:build macro`。若 IDE 同时分析宏主文件和 `*_macro_gen.go`，gopls 可能提示重复定义。

你可以任选一种方式处理：

1. 在工作区设置中排除 `*_macro_gen.go`，让 IDE 只看生成侧
2. 为 gopls 配置 `-tags=macro`，让 IDE 按宏版本源码做类型检查与补全：

```json
"gopls": {
  "buildFlags": ["-tags=macro"]
}
```

更多文件布局说明见 [宏作者指南 · 宏使用方](docs/author-guide.md#宏使用方)。

## 参考

### 官方宏库

[go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) 提供官方宏库。在宏主文件中 import 对应包，执行 `cmd/macro expand` 即可展开。


| syntax-id       | 模块路径                                              | 说明  |
| --------------- | ------------------------------------------------- | --- |
| `inline` | `github.com/arcane-craft/go-macro-contrib/inline` | 过程宏 |
| `try`    | `github.com/arcane-craft/go-macro-contrib/try`    | 过程宏 |


更多说明（含最低兼容核心版本）见 [go-macro-contrib](https://github.com/arcane-craft/go-macro-contrib) 仓库 README。

### 本地联调

将 `go-macro-contrib` clone 到与 `go-macro` 同级目录后，可在你的 `go.mod` 中加：

```go
replace github.com/arcane-craft/go-macro-contrib => ../go-macro-contrib
```

本仓库根目录可选用 `go.work`（`use` 为 `.` 与 `./examples`）。测试框架核心：

```bash
GOWORK=off go test ./...
```

在 `examples/` 目录下：

```bash
go test ./...
```

