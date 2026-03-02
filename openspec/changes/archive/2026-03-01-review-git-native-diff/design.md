## Context

当前 `/gsf:review` skill 依赖 `gsf diff` 命令获取变更信息。`gsf diff` 的工作流是：git diff → 解析 hunk → AST 映射到函数 → 输出完整函数代码。这个流程存在 AST 覆盖盲区（FuncLit、包级变量等），且丢失了原始 diff 信息（reviewer 看不到具体改了哪些行）。

AI agent（Claude Code 或内部工具）本身就能直接运行 `git diff` 并阅读输出。gsf 不需要包装 git diff，而应该专注于 AI 做不到的事：Go AST 分析（调用链追踪、调用者查找）。

## Goals / Non-Goals

**Goals:**
- Review skill 基于原生 git diff，获取完整、无遗漏的变更信息
- Review 展示真实 diff（+/- 行），而非只展示变更后的完整函数
- 保留 `gsf trace` / `gsf callers` 作为补充上下文工具
- 删除 `gsf diff` 命令及其全部依赖代码，减少维护负担

**Non-Goals:**
- 不改动 `gsf trace`、`gsf callers`、`gsf analyze` 等命令
- 不改动 `internal/ast/` 包的任何代码
- 不引入新的 gsf 命令

## Decisions

### 1. 删除 `gsf diff` 命令及全部 review 模块代码

**选择**：完全删除，不保留

**理由**：
- `internal/review/diff.go`（git diff 解析）：AI 直接读 git diff 输出即可
- `internal/review/mapper.go`（hunk → 函数映射）：有 FuncLit 盲区，且映射本身不是 reviewer 需要的
- `internal/review/extract.go`（函数代码提取）：reviewer 需要看 diff，不是完整函数
- 这些代码没有其他调用者，删除无副作用

**替代方案考虑**：保留 `gsf diff` 作为可选工具 → 拒绝，维护成本高且功能与 git diff 重复

### 2. Review skill 直接使用 git diff

**选择**：skill 指导 AI 运行 `git diff` 命令

**理由**：
- `git diff` 完备、灵活、零盲区
- AI agent 本身就能运行 shell 命令，不需要 gsf 包装
- 支持所有 diff 范围：`--staged`、`HEAD~N`、`base...HEAD` 等

### 3. Review 内容展示真实 diff + 按动线组织

**选择**：review 文档按逻辑动线组织，每个节点展示相关的 git diff 片段，需要时用 gsf trace/callers 补充上下文

**review 流程**：
1. `git diff` 获取完整变更（AI 直接运行）
2. AI 分析变更意图，确定 review 动线
3. 按动线组织，每个节点展示对应的 diff 片段
4. 按需调用 `gsf callers` / `gsf trace` 补充调用关系
5. 给出 review 结论

## Risks / Trade-offs

- **[Risk] git diff 输出量大时 AI context window 压力** → AI 可以分文件读取，或用 `--stat` 先概览再选择性深入
- **[Risk] 删除后无法用 gsf 做函数级变更统计** → 当前无此需求，未来需要时可重新实现
- **[Trade-off] 不再有结构化的函数级变更输出（JSON/YAML）** → 实际使用中 AI 直接读 git diff 更高效，结构化输出未被消费
