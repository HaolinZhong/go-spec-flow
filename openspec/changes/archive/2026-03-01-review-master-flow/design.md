## Context

当前 `/gsf:review` 的 AI 编排（Step 3）将代码组织成多个并列的子动线（roots），每个子动线覆盖一个模块。这种扁平结构缺少全局叙事 — 用户看到 4 个并列动线，不知道它们之间的执行顺序和因果关系。

FlowTree 数据结构已支持任意深度嵌套（FlowNode.Children），HTML 树形组件也支持多级展开/折叠。

## Goals / Non-Goals

**Goals:**
- AI 编排时构建单一主动线 root，讲述完整的请求/指令生命周期
- 现有子动线嵌套为主动线各步骤的 children，保留全部细节
- 用户"先看全局 → 按需展开"的层级阅读体验
- 无 code 的 group 节点（主动线步骤）点击时展示 description

**Non-Goals:**
- 不改 FlowTree/FlowNode 数据结构（已经支持 description + 嵌套 children）
- 不改 `gsf review` CLI 的 flags 或行为
- 不做跨动线交叉引用/链接

## Decisions

### 1. 单 root 主动线 vs 首位 root + 并列子动线

选择**单 root 主动线**，所有子动线作为其 children。

理由：单 root 形成真正的层级树，用户展开主动线后自然看到步骤，展开步骤后看到函数。如果用"首位 root 概览 + 并列子动线"，树形面板仍然是扁平的，没有层级感。

### 2. Skill 指令更新策略

只修改 Step 3 的编排指令和 JSON 示例。Step 1/2/4 保持不变。

### 3. HTML group 节点展示

当前点击无 code 的节点显示 "No code available"。改为：如果节点有 description 但无 code，展示 description 卡片作为该步骤的说明。

## Risks / Trade-offs

- **AI 编排质量依赖 prompt** — 主动线的叙事质量完全取决于 AI 理解代码的能力。通过在 skill 中提供明确的示例和规则来引导。
- **三层嵌套在小包上可能过度** — 对于只有 3-4 个函数的小包，强制三层（主动线→步骤→函数）可能增加点击负担。在 skill 指令中说明：当代码量很少时可以简化为两层。
