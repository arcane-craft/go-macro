## 1. 引擎：值用法校验

- [x] 1.1 抽取或共享 `objectStub` / `isPackageSelector` / `unwrapParen`，供识别与校验共用（`recognize.go` 或 `stub_resolve.go`）
- [x] 1.2 实现 `isDirectMacroCallee`（父节点表 + `CallExpr.Fun` 与 `unwrapParen` 比较）
- [x] 1.3 实现 `ValidateStubValueUsage(fset, file, info, reg) error`：遍历 `Ident`/`SelectorExpr`，对已注册桩且非 callee 返回 `macro.ErrorAt`
- [x] 1.4 在 `ExpandFile` 开头调用 `ValidateStubValueUsage`（先于 `RecognizeMacroCalls`）

## 2. 测试（对齐 macro-expander spec scenarios）

- [x] 2.1 桩作实参：`apply(tr.Try)` → 期望失败
- [x] 2.2 赋值：`fn := tr.Try` → 期望失败
- [x] 2.3 return 值：`return tr.Try` → 期望失败
- [x] 2.4 reflect：`reflect.ValueOf(tr.Try)` → 期望失败
- [x] 2.5 死代码：`if false { _ = tr.Try }` → 期望失败
- [x] 2.6 直调合法：`return tr.Try(1)`、`(tr.Try)(1)` → 校验通过；可与现有 recognize 测共用
- [x] 2.7 未 link：import provider 但未 `RegisterProvider` → 值用法校验不失败
- [x] 2.8 shadow：包内 `func Try` + `Try(1)` / `_ = Try` → 不失败
- [x] 2.9 方法：`s.Try(1)` → 不失败
- [x] 2.10 嵌套：`outer(tr.Try(1))` → 校验通过
- [x] 2.11（可选）`ExpandFile` 在值用法失败时不 mutate 文件 / 不展开

## 3. 文档

- [x] 3.1 更新 `docs/author-guide.md`：编写宏库 / 宏使用方补充「桩须直调、不可作函数值、expand 报错」
- [x] 3.2 通读与 `specs/author-guide/spec.md` delta 一致，链接行为与 README 不冲突

## 4. 收尾

- [x] 4.1 `go test ./internal/expander/...` 通过
- [x] 4.2 `GOWORK=off go test ./...`（及 `cd examples && go test ./...` 若适用）通过
- [x] 4.3 `openspec validate reject-stub-value-usage` 通过
- [x] 4.4 归档 change：`openspec archive reject-stub-value-usage`（合并 spec 至 `openspec/specs/`）
