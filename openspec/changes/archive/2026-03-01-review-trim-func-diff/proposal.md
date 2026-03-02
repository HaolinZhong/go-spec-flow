## Why

`extractFuncDiff` 当前逻辑是：只要 hunk 与函数行范围有重叠，就把整个 hunk 加入结果。这导致函数级 diff 远大于实际变更——一个 300B 代码的函数可能附带 8000B 的 diff，因为同一个 hunk 跨越了多个函数。在 review HTML 中，Source 视图精确展示函数代码，但切换到 Diff 视图后变成一大片，体验割裂。

## What Changes

- 修改 `extractFuncDiff`：在 hunk 与函数行范围重叠时，只保留函数行范围内的 diff 行（加上必要的 `@@` header），裁剪掉函数范围外的上下文行
- 保持 `extractFuncDiff` 的接口和调用方式不变

## Capabilities

### New Capabilities

### Modified Capabilities
- `review-comments` 中 diff 显示的精确度提升

## Impact

- `internal/review/diff.go` — 修改 `extractFuncDiff` 的 hunk 裁剪逻辑
