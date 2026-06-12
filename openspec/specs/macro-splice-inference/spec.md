# macro-splice-inference Specification

## Purpose
本 capability 已合并至 `macro-splice-apply`（`SplicePlan` / `SpliceStep` 模型，取代 `InferTarget` + `SpliceTarget` 推断）。
## Requirements
### Requirement: 已合并至 macro-splice-apply

`InferTarget`、`Validate` 载荷与 `SpliceTarget` 推断 MUST NOT 再作为 normative API。贴回 MUST 由 `site.Match` 产出 `SplicePlan`，经 `ValidateSplice` + `Apply` 执行。

#### Scenario: 无 InferTarget

- **WHEN** 阅读 expand 引擎公开贴回路径
- **THEN** MUST NOT 出现 `InferTarget` 或 normative `SpliceTarget` 作者 API
