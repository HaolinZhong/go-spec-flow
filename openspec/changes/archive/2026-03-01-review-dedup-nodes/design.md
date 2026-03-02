## Context

FlowTree 的 builder 在构建调用链时，对每个函数独立执行 Tracer.Trace()，产生的 CallNode 树通过 callNodeToFlowNode 递归转为 FlowNode。当多个函数调用同一个下游函数时，该函数在 JSON 中作为多个父节点的 child 各出现一次。同时，同文件中声明的函数既是顶层 FlowNode 又是彼此的 trace child，造成进一步重复。

当前实测：143 个节点中，27 个函数存在重复（`contains` 10x, `ResolveCallTarget` 9x, `readFileLines` 8x）。

## Goals / Non-Goals

**Goals:**
- 消除同文件兄弟函数在 trace children 中的冗余出现
- 减少跨 trace 分支中相同函数的重复完整展示
- 保持 FlowTree 的调用关系语义完整性

**Non-Goals:**
- 不改变 HTML 渲染逻辑（消费相同 FlowNode 结构）
- 不改变 goast.Tracer 本身的行为
- 不做跨文件级别的全局去重（那属于 AI 编排的职责）

## Decisions

### Decision 1: 两层去重策略

**选择**：在 builder 层实施两层去重，不修改 Tracer。

- **Layer 1 — 同文件兄弟过滤**：构建完一个文件的所有函数节点后，收集所有已声明函数名为 `siblingNames`。递归过滤每个函数的 trace children，移除 label 匹配兄弟的节点。
- **Layer 2 — per-file seen 集合**：为每个文件创建 `seen map[string]bool`，在 callNodeToFlowNode 递归中传入。首次遇到的函数完整展示（带 code + children），后续遇到的返回精简引用节点（仅 label + package + nodeType，无 code/children）。

**理由**：在 Tracer 层去重会破坏其通用性；在 HTML 层去重太晚（JSON 已膨胀）；在 builder 层去重刚好，既精确又不影响上下游。

### Decision 2: 精简引用节点的设计

**选择**：已 seen 的函数返回 FlowNode，保留 ID/Label/Package/NodeType/Description，清空 Code/Diff/Children。

**理由**：保留 label 使得 HTML 的 nodeIndex 能识别和链接到首次出现的完整节点。不携带 code 和 children 避免 JSON 膨胀。

### Decision 3: callNodeToFlowNode 签名变更

**选择**：增加 `seen map[string]bool` 参数。key 格式为 `pkg.FuncName`。

**替代方案**：使用全局变量或在 struct 中维护 seen — 拒绝，因为 builder 函数是无状态的，显式参数更清晰。

## Risks / Trade-offs

- **调用关系信息丢失**：过滤兄弟 trace child 后，用户无法从代码中直接看到"函数 A 调用了同文件的函数 B"。→ 缓解：函数 B 已作为顶层兄弟节点存在，用户可通过 nodeIndex 的跳转链接导航。
- **精简引用节点在 AI 编排时可能缺少上下文** → 缓解：引用节点保留 description 字段，AI 仍有足够信息进行分组。
