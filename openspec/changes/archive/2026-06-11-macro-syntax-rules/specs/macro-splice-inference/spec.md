## REMOVED Requirements

本 capability 已合并至 **`macro-splice-apply`**（`SplicePlan` / `SpliceStep` 模型，取代 `InferTarget` + `SpliceTarget` 推断）。

### Requirement: InferTarget 推断贴回目标

**Reason**: 贴回计划在 `site.Match` 时确定；见 macro-splice-apply「Match 产出 SplicePlan」。

**Migration**: 删除 `InferTarget`；使用 `ValidateSplice(out, meta)` + `Apply(file, meta, out)`。

### Requirement: Apply 仅替换 MatchedSpan（InferTarget 版）

**Reason**: Apply 签名与执行模型迁至 macro-splice-apply（按 `[]SpliceStep` 执行）。

**Migration**: 同上。

### Requirement: Validate 载荷与 Target

**Reason**: 合并为 `ValidateSplice(out, meta)`，校验 `out` 与 `meta.Plan` 而非 `SpliceTarget`。

**Migration**: 同上。
