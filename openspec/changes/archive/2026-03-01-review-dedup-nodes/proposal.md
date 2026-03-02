## Why

FlowTree 渲染的 HTML 中，同一个函数/方法大量重复出现（如 `contains` 出现 10 次，`ResolveCallTarget` 9 次）。根因：trace 调用链中，同一函数作为多个父函数的 callee 被独立创建为节点，且同文件兄弟函数也作为 trace child 重复出现。用户浏览时看到大量重复节点，难以分辨和导航。

## What Changes

- Builder 在构建 FlowTree 时增加去重逻辑：
  - 同文件兄弟过滤：函数 A 调用同文件的函数 B，B 已是顶层节点，从 A 的 trace children 中移除
  - 跨分支 seen 集合：同一文件内，已在某个 trace 分支完整展示过的函数，后续分支只保留精简引用（无 code、无 children）
- 影响 `buildFuncNodesFromDiff` 和 `BuildCodebaseTree` 两个入口
- `callNodeToFlowNode` 增加 `seen` 参数支持去重

## Capabilities

### New Capabilities
- `trace-dedup`: FlowTree 构建时的调用链节点去重机制

### Modified Capabilities

## Impact

- `internal/review/builder.go` — 修改 buildFuncNodesFromDiff、BuildCodebaseTree、callNodeToFlowNode
- FlowTree JSON 输出结构变化：重复节点被精简为引用，节点总数显著减少
- HTML 渲染不受影响（消费相同的 FlowNode 结构）
