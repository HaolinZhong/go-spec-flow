## Context

`extractFuncDiff(fileDiff, lineStart, lineEnd)` 从文件级 diff 中提取与函数行范围重叠的 hunk。当前实现在 `flushHunk` 中判断 hunk 是否与函数范围重叠后，把整个 hunk 的所有行都加入结果。

unified diff 中每一行的含义：
- `@@` 行：hunk header，标记 hunk 的起始行号
- ` ` 开头：上下文行（两边都有），推进 new 和 old 行号
- `+` 开头：新增行，只推进 new 行号
- `-` 开头：删除行，只推进 old 行号

函数的 `lineStart`/`lineEnd` 对应的是 **new 文件**的行号（来自 AST 解析当前代码）。

## Goals / Non-Goals

**Goals:**
- 函数级 diff 只包含函数行范围内（lineStart ≤ newLine ≤ lineEnd）的 diff 行
- 裁剪后的 diff 仍然是合法的 unified diff 片段（带正确的 `@@` header）
- `-` 行（删除行）如果对应的 old 行位置在函数范围内也应保留

**Non-Goals:**
- 不改变 `extractFuncDiff` 的函数签名
- 不改变文件级 diff（FlowNode.Diff 在文件级 root 节点仍是完整文件 diff）
- 不处理函数被整体移动/重命名的场景（那种情况下 AST 行号已经对不上）

## Decisions

### Decision 1: 逐行裁剪策略

**选择**：在遍历 hunk 内部行时，追踪当前 new-file 行号（newLine），只保留 `newLine` 落在 `[lineStart, lineEnd]` 范围内的行。

对于每种行类型：
- `+` 行：推进 newLine，当 newLine 在范围内时保留
- ` ` 上下文行：推进 newLine（和 oldLine），当 newLine 在范围内时保留
- `-` 行：不推进 newLine（删除行在 new file 中不存在），但如果紧邻范围内的 `+` 或 ` ` 行则保留

**理由**：逐行追踪 newLine 是唯一精确的方法。`-` 行的处理需要特殊考虑，因为它们没有 new-file 行号。

### Decision 2: `-` 行保留策略

**选择**：`-` 行保留的条件是：它紧邻（前面或后面）有落在函数范围内的 `+` 行或 ` ` 行。

实现方式：先收集 hunk 内所有行并标记 newLine，然后做两遍扫描：
1. 第一遍：标记所有 newLine 在范围内的 `+` 和 ` ` 行
2. 第二遍：`-` 行如果与已标记的行相邻则也标记

**简化替代**：直接保留所有 `-` 行只要它们在一个被裁剪过的 hunk 内。这更简单且在大多数情况下足够精确。

**选择简化方案**：保留 hunk 内、且位于第一个保留行和最后一个保留行之间的所有行（包括 `-` 行）。

### Decision 3: `@@` header 重写

**选择**：裁剪后的 hunk 需要更新 `@@` header 中的行号和计数。但由于这个 diff 仅用于 HTML 展示而不是 patch apply，可以保留原始 `@@` header 不修改。

**理由**：HTML 渲染器（review.html）解析 `@@` 行只是为了显示行号和着色，不做 patch 校验。保留原始 header 避免了复杂的行号重算逻辑。

## Risks / Trade-offs

- **`-` 行可能被误裁或误保留** → 使用"保留范围内连续块"的策略，大多数场景准确
- **跨函数边界的变更可能被截断** → 可接受，reviewer 可以看文件级 diff 获取完整上下文
