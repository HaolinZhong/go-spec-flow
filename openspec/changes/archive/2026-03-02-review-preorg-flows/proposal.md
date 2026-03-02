## Why

当前 `gsf review --json` 在 diff 模式下的输出结构是按文件组织的：每个变更文件是一个 root，其下包含该文件的**全部**函数（不管是否有实际变更），每个函数再附带 trace 子节点。这导致两个严重问题：

1. **输出膨胀**：一个文件改了 1 个函数，但输出包含该文件全部 20 个函数 + 它们的 trace 子节点，上下文极度冗长
2. **AI 编排门槛高**：`/gsf:review` 的 Step 3 需要 AI 从平铺的文件列表中理解调用关系、构建动线（master flow）。弱模型（团队内部 AI 工具）根本做不到这一步

需要让 gsf 在输出阶段就完成调用链组织，降低 AI 编排门槛，同时大幅减少输出体积。

## What Changes

- `BuildDiffTree` 重构：不再输出文件下的全部函数，而是只输出有实际 diff 的函数（`funcDiff != ""`）
- 新增调用链组织逻辑：在变更函数之间构建调用图，识别入口函数（不被其他变更函数调用的变更函数），从入口向下构建 flow 树
- 引入桥接函数（bridge）：当两个变更函数之间存在未变更的中间函数时，作为桥接节点保留，维持调用链连通性
- 独立变更函数（不与其他变更函数有调用关系的）作为独立 flow 输出
- 非 Go 文件（.yaml, .md 等）归组到一个 "Non-code Files" 节点下
- 更新 `/gsf:review` skill，简化 Step 3 的 AI 编排工作

## Capabilities

### New Capabilities
- `review-preorg-flows`: diff 模式下 JSON 输出按调用链预组织为 flow 树，包括入口检测、桥接函数、独立 flow、非代码文件归组

### Modified Capabilities
- `review-comments` 的 `/gsf:review` skill 需要更新 Step 3 以适配新的预组织 JSON 结构

## Impact

- `internal/review/builder.go` — 重构 `BuildDiffTree` 和 `buildFuncNodesFromDiff`，新增调用链组织逻辑
- `internal/review/diff.go` — 可能新增变更函数提取辅助函数
- `internal/review/tree.go` — FlowNode 可能新增 `IsBridge` 字段标记桥接函数
- `skills/gsf-review.md` + `internal/cmd/embed_data/skills/gsf-review.md` — 更新 Step 3 AI 编排指令
