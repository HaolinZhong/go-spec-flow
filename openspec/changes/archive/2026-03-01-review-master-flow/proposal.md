## Why

当前 AI 编排的动线是并列的子模块（如 "Git Diff 层"、"构建器"、"渲染"），缺少一条**贯穿全局的主动线**来串联它们。开发者看到 4 条并列动线，仍需自行理解它们之间的关系和执行顺序，心智负担没有实质降低。

需要让 AI 先建立一条"请求/指令的完整生命周期"主动线，告诉开发者整体流程是 1→2→3→4，然后每一步可以展开查看子动线细节。

## What Changes

- 更新 `skills/gsf-review.md` Step 3 编排指令：要求 AI 先构建主动线（master flow），再将子动线挂载为主动线各步骤的 children
- 更新 JSON 输出格式示例：从多个并列 roots 改为单个主动线 root + 多层嵌套 children
- 更新 HTML 模板：为主动线节点（无 code 的 group 节点）增加描述展示，点击时显示该步骤的说明而非空白

## Capabilities

### New Capabilities

- `master-flow-orchestration`: AI 编排时构建单一主动线 root，子动线作为嵌套 children，形成"概览→展开"的层级阅读体验

### Modified Capabilities

## Impact

- `skills/gsf-review.md` — Step 3 编排指令重写
- `internal/cmd/embed_data/skills/gsf-review.md` — 同步 embed 副本
- `internal/review/templates/review.html` — group 节点点击时展示 description
- 需重新构建 gsf 二进制
