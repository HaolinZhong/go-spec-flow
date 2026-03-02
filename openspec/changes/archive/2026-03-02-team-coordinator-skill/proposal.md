## Why

当前 `opsx:decompose` 可以将大需求拆为多个 changes，`opsx:next` 按依赖顺序逐个推进。但对于大 feature（5+ changes），顺序执行太慢。当多个 changes 之间没有依赖关系时，应该可以并行执行。

`opsx:team` 利用 Claude Code 原生 Agent tool（worktree 隔离 + subagent），实现 Coordinator + Workers 模式的并行执行。这是 `opsx:next` 的并行增强版，仅限 Claude Code 环境使用。

Parent feature: [opsx-team](../../features/opsx-team/feature-spec.md)

## What Changes

- 新增 `opsx:team` skill，实现完整的 Coordinator + Workers 并行执行流程：
  - Coordinator 读取 Feature Spec，构建依赖图，按批次识别可并行的 changes
  - 每批次：并行启动 Workers（每个在独立 worktree 中 propose）
  - Proposals 就绪后，人逐个审阅（风险前置，不可跳过）
  - 审阅通过的 changes → Workers 并行 apply
  - 更新 Feature Spec 状态，识别下一批 unblocked changes，继续
- 利用 Claude Code 原生 Agent tool 的 `isolation: "worktree"` 实现 Worker 隔离
- Workers 遇到歧义时暂停等待人介入，不自行决策
- Feature Spec 是 single source of truth，Coordinator 负责更新状态

## Capabilities

### New Capabilities
- `team-execution`: Coordinator + Workers 并行执行 Feature Spec changes 的完整工作流，包括批次调度、worktree 隔离、人工审阅、状态管理

### Modified Capabilities
(none)

## Impact

- 新增文件: `.claude/commands/opsx/team.md`, `.claude/skills/openspec-team/SKILL.md`
- 依赖 Claude Code 原生 Agent tool（subagent_type, isolation: "worktree"）
- 与 `opsx:next` 互补，不替代；Feature Spec 格式完全兼容
- 仅 Claude Code 环境可用，弱模型/工具（`.coco/`）继续用 `opsx:next`
