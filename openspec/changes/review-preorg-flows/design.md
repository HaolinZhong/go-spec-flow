## Context

当前 `BuildDiffTree` 的 diff 模式输出结构：

```
FlowTree
├── Root: file-0 (internal/review/builder.go)  ← 变更文件
│   ├── Func: BuildDiffTree          ← 实际有 diff
│   │   └── trace children...
│   ├── Func: BuildCodebaseTree      ← 无 diff，但被包含
│   │   └── trace children...
│   ├── Func: callNodeToFlowNode     ← 无 diff，但被包含
│   │   └── trace children...
│   └── ... (文件中所有函数)
├── Root: file-1 (internal/review/diff.go)
│   └── ... (同上，全部函数)
└── ...
```

问题：`buildFuncNodesFromDiff` 遍历变更文件的全部 `ast.FuncDecl`，不判断函数是否有实际变更（`funcDiff`），导致输出包含大量无关函数。AI 在 Step 3 需要从这些平铺节点中理解调用关系并构建动线，弱模型无法胜任。

目标输出结构：

```
FlowTree
├── Root: flow-0 "处理 diff 输出"     ← 调用链 flow
│   ├── BuildDiffTree                 ← 入口（变更函数，不被其他变更函数调用）
│   │   └── trace children...
│   ├── buildFuncNodesFromDiff        ← 桥接函数（未变更，连接两个变更函数）
│   └── extractFuncDiff              ← 变更函数（被 buildFuncNodesFromDiff 调用）
├── Root: flow-1 "新增评论保存"        ← 另一条调用链
│   └── SaveComment                   ← 独立变更函数
├── Root: non-code-files "Non-code Files"
│   ├── config.yaml
│   └── README.md
└── ...
```

## Goals / Non-Goals

**Goals:**
- 只输出有实际 diff 的 Go 函数（`funcDiff != ""`），大幅减少输出体积
- 在变更函数之间建立调用关系，自动识别入口函数并构建 flow 树
- 引入桥接函数连接不相邻的变更函数，保持调用链连通
- 独立变更函数作为独立 flow 输出
- 非 Go 文件归组到 "Non-code Files" 节点
- AI（包括弱模型）可直接使用 JSON 输出做简单编排，无需理解复杂调用关系

**Non-Goals:**
- 不改变 codebase 模式的输出结构（只影响 diff 模式）
- 不改变 HTML 渲染逻辑（已有的树渲染和 source/diff 切换不受影响）
- 不做跨文件函数级 diff 合并（每个函数节点仍然关联自己的 funcDiff）
- 不引入外部依赖

## Decisions

### Decision 1: 变更函数识别

**选择**：用 `extractFuncDiff(fileDiff, startLine, endLine)` 的返回值判断。`funcDiff != ""` 即为变更函数。

**理由**：这是现有逻辑中已有的函数，精确可靠。无需引入新的分析手段。

### Decision 2: 调用图构建策略

**选择**：对所有变更函数两两检查是否存在调用关系，通过现有 `Tracer.Trace()` 实现。

算法：
1. 收集所有变更函数 `changedFuncs []ChangedFunc`（包含 pkg, name, diff, code 等）
2. 构建 `changedSet map[string]bool`（key: `pkg.Name`）
3. 对每个变更函数调用 `Tracer.Trace(pkg, name)`，遍历 trace 树
4. 在 trace 树中查找其他变更函数：
   - 直接子节点是变更函数 → 记录 `A → B` 边
   - 子节点不在变更集中但其后代在 → 该子节点是桥接函数，记录 `A → bridge → B`
5. 构建 DAG，拓扑排序找入口（入度为 0 的变更函数）

**理由**：复用现有 Tracer 能力，不需要新建 callgraph 分析器。MaxDepth 可控制搜索范围。

**替代方案**：全项目 callgraph 分析（`golang.org/x/tools/go/callgraph`）— 过重，且当前 trace 已足够。

### Decision 3: 桥接函数处理

**选择**：桥接函数是 trace 路径上连接两个变更函数的未变更函数。在 FlowNode 上新增 `IsBridge bool` 字段标记。

桥接条件：
- 函数 X 不在 changedSet 中
- X 位于变更函数 A 的 trace 路径上
- X 的 trace 后代中包含另一个变更函数 B

桥接函数的 FlowNode：
- 有 code（读取源码）、有 file/line 信息
- `IsBridge: true`
- 没有 diff（因为未变更）
- 没有 trace children（只作为连接节点）

**理由**：桥接函数让 reviewer 理解变更函数之间的调用路径，是 "tour guide" 体验的关键。但不需要深入展开它的子节点。

### Decision 4: 入口函数识别

**选择**：在变更函数的调用 DAG 中，入度为 0 的变更函数即为入口。

入口函数的特征：
- 有实际 diff（在 changedSet 中）
- 不被任何其他变更函数直接或间接调用
- 每个入口函数及其可达的变更函数+桥接函数构成一个 flow

**边界情况**：
- 循环调用：用 visited set 打断循环，保留首次遇到的方向
- 所有变更函数互相独立：每个都是独立 flow
- 单个变更函数：单节点 flow

### Decision 5: 独立变更函数

**选择**：不与任何其他变更函数有调用关系的变更函数，作为独立 flow 输出。不保留 trace 子节点。

**理由**：explore 阶段确认 — 独立变更函数的 trace 子节点（都是未变更函数）提供的上下文价值有限，去掉可显著减少输出体积。AI 如果需要了解独立函数的调用链，可以在 review 编排时按需调用 `gsf trace`。

### Decision 6: 非 Go 文件处理

**选择**：所有非 `.go` 后缀的变更文件归组到一个 root 节点 "Non-code Files"，每个文件作为子节点，保留 diff。

```json
{
  "id": "non-code-files",
  "label": "Non-code Files",
  "nodeType": "file",
  "children": [
    { "id": "file-config.yaml", "label": "config.yaml", "diff": "...", "nodeType": "file" }
  ]
}
```

**理由**：非 Go 文件无法做 AST 分析和 trace，但仍然是变更的一部分，需要展示。归组避免污染 flow 结构。

### Decision 7: Flow 树输出结构

**选择**：每个 flow 是一个 root，roots 按以下顺序排列：

1. 调用链 flow（按入口函数名排序）
2. 独立变更 flow（按函数名排序）
3. Non-code Files（如果有）

每个 flow root 的 label 使用入口函数名或描述，`nodeType: "file"`（让 HTML 渲染正确处理层级）。

### Decision 8: 对 `/gsf:review` skill 的影响

**选择**：更新 Step 3，利用预组织的 flow 结构简化 AI 编排。

变更前：AI 需要从平铺文件列表构建完整动线
变更后：AI 只需为每个 flow 添加 description，并可选地调整 flow 顺序或合并 flow

**理由**：这是本次改造的核心价值 — 降低 AI 编排门槛，让弱模型也能生成有意义的 review。

## Risks / Trade-offs

- **Trace 深度限制** → 两个变更函数之间的调用路径可能超过 maxDepth 导致找不到桥接函数。缓解：对入口检测使用更大的 depth（如 6），因为这是一次性分析不影响输出体积
- **调用图构建耗时** → 每个变更函数都要 trace，如果变更函数多（>50）可能较慢。缓解：变更函数多的情况下可以跳过调用图分析，退化为每函数独立 flow
- **桥接函数误判** → 两个变更函数恰好经过同一个桥接函数但逻辑无关。可接受：reviewer 可以看到桥接函数的代码，自行判断关联性
- **Codebase 模式不受影响** → 本次只改 diff 模式。codebase 模式的全函数输出在架构探索场景下是合理的
