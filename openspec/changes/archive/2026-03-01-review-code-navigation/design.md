## Context

当前 HTML review 的代码面板是只读展示——函数名只是文本，无法交互。主动线步骤节点 `code` 为空，用户看到的只有抽象描述。

树数据已包含所有节点的 `id` 和 `label`（函数名），为代码跳转提供了天然索引。

## Goals / Non-Goals

**Goals:**
- 步骤节点携带入口函数代码，让主动线有"代码可读"
- 代码中出现的函数名若存在于树中，渲染为可点击链接
- 点击链接自动导航：展开目标节点的父节点链、高亮、显示代码+描述
- 仅链接树中已有的函数，不链接外部/不存在的函数

**Non-Goals:**
- 不做全文搜索或"Go to Definition"
- 不改 FlowTree/FlowNode 数据结构
- 不做描述文本中的链接（仅代码区域）

## Decisions

### 1. 节点索引构建

在 JS 初始化时递归遍历 flowData，构建 `Map<funcName, {node, labelEl}>` 索引。同时在 `createTreeNode` 中存储 label DOM 元素引用，用于跳转时高亮和展开。

### 2. 代码后处理时机

在 highlight.js 渲染完成后，对 `#code-content` 做一次 DOM 后处理：遍历文本节点，对匹配索引中函数名的文本包裹 `<span class="code-link" data-target="funcName">` 元素。

匹配策略：在已 escape 的 HTML 文本中用正则 `\b(FuncName)\b` 匹配。只匹配索引中存在的函数名，避免误匹配关键字。

### 3. 跳转行为

点击 code-link 时：
1. 在索引中查找目标节点
2. 展开从 root 到目标节点的所有父节点（递归打开 `.tree-children`）
3. 移除旧 `.active`，添加新 `.active` 到目标 label
4. 调用 `showCode(targetNode)` 显示代码+描述
5. 滚动左侧树面板使目标 label 可见

### 4. Skill 指令 — 步骤节点带代码

在 Step 3 指令中明确：每个步骤节点应设置 `code` 为该步骤入口函数的完整代码。入口函数同时也作为子节点存在（带详细描述）。步骤节点的 code 让用户在主动线层级就能看到代码。

## Risks / Trade-offs

- **函数名冲突** — 不同包可能有同名函数（如 `New`）。索引用 label 做 key，同名后者覆盖前者。可接受，因为同一个 review scope 内同名函数很少。
- **highlight.js 后处理** — 对已高亮的 DOM 做文本替换可能破坏 highlight.js 的 span 结构。解决方案：只在**文本节点**中做替换，不触碰已有的 HTML 标签。
