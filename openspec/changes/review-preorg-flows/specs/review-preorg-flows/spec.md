# review-preorg-flows

Diff 模式下 `gsf review --json` 输出按调用链预组织为 flow 树。

## Requirements

### R1: 变更函数过滤

- 只有 `funcDiff != ""` 的函数才被识别为变更函数
- 变更函数需记录：package path, function name, file path, line range, source code, funcDiff
- 非变更函数不作为顶级节点输出

### R2: 调用图构建

- 对每个变更函数调用 `Tracer.Trace()`，搜索其 trace 树中的其他变更函数
- 记录变更函数之间的有向边（A 调用 B）
- 如果路径上存在未变更的中间函数，记录为桥接节点
- 使用 visited set 处理循环调用

### R3: 入口函数识别

- 在变更函数调用 DAG 中，入度为 0 的变更函数是入口
- 每个入口及其可达的变更函数和桥接函数构成一个 flow
- 如果变更函数不与其他变更函数有调用关系，作为独立 flow

### R4: Flow 树构建

- 每个 flow 是 FlowTree.Roots 中的一个 root
- Flow root 的 children 包含：入口函数节点 + 桥接函数节点 + 被调用的变更函数节点
- 节点按调用链顺序排列
- 桥接函数节点标记 `IsBridge: true`，有 code 但无 diff、无 children

### R5: 独立变更函数

- 不与任何其他变更函数有调用关系的函数作为独立 flow
- 独立 flow 只包含该函数自身，不保留 trace 子节点

### R6: 非 Go 文件归组

- 非 `.go` 后缀的变更文件归组到一个 "Non-code Files" root 节点
- 每个文件作为子节点，保留 file path 和 diff

### R7: 输出顺序

- Roots 按以下顺序排列：调用链 flow → 独立变更 flow → Non-code Files
- 调用链 flow 按入口函数名排序
- 独立变更 flow 按函数名排序

### R8: 向后兼容

- `--json` 输出的 FlowTree/FlowNode JSON 结构不变（只新增 `isBridge` 可选字段）
- `--codebase` 模式完全不受影响
- HTML 渲染逻辑不需要修改（现有树渲染和 source/diff 切换正常工作）

## Acceptance Criteria

- 对包含 3 个变更函数且其中 2 个有调用关系的 diff，输出 2 个 roots（1 个调用链 flow + 1 个独立 flow），而非按文件平铺全部函数
- 桥接函数出现在调用链 flow 中，标记 `isBridge: true`，有源码但无 diff
- 未变更函数不出现在输出中（除非是桥接函数）
- 非 Go 文件归组到 Non-code Files
- `gsf review --commit HEAD --json | jq '.roots | length'` 返回的 root 数量明显少于变更前
- `gsf review --render` 正确渲染新结构（树、代码面板、source/diff 切换）
