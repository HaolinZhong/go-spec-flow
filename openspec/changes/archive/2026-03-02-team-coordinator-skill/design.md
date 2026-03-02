## Context

`opsx:next` 实现了顺序执行 Feature Spec changes 的能力。对于有多个无依赖关系的 changes 的大 feature，顺序执行效率低。Claude Code 原生 Agent tool 提供了 worktree 隔离的 subagent 能力，可以利用这一机制实现并行执行。

现有相关 skills:
- `opsx:decompose` — 将 PRD 拆解为 Feature Spec
- `opsx:next` — 按依赖顺序找下一个 unblocked change，启动 propose
- `opsx:propose` — 创建 OpenSpec change + 生成所有 artifacts
- `opsx:apply` — 按 tasks.md 实施

## Goals / Non-Goals

**Goals:**
- Coordinator 能够读取 Feature Spec，构建依赖图，识别每批可并行的 changes
- 并行启动 Workers（每个在独立 worktree），执行 propose
- 人逐个审阅 proposals 后，并行执行 apply
- 按批次推进，直到所有 changes 完成
- Workers 遇到歧义时暂停等待人介入

**Non-Goals:**
- 不自动 merge worktree 成果到 main（人最终决定）
- 不自动重试失败的 changes
- 不做跨机器分布式执行
- 不适配弱模型/工具环境（那些场景用 opsx:next）
- 不需要外部依赖或新的 CLI 工具

## Decisions

### 1. 使用 Claude Code Agent tool 实现 Workers

**选择**: 使用 Agent tool 的 `isolation: "worktree"` 模式启动 subagent

**理由**: Claude Code 原生支持 worktree 隔离的 subagent，自动创建和清理 worktree。不需要自建 tmux 管理或进程管理。每个 Worker 在独立文件系统空间工作，避免冲突。

**替代方案**:
- tmux 多 session 管理 — 需要自建进程管理，复杂度高
- 顺序执行 — 太慢，不满足需求

### 2. Skill 而非代码

**选择**: `opsx:team` 是纯 Markdown skill（`.claude/commands/opsx/team.md`），不涉及 Go 代码

**理由**: 与 opsx:decompose、opsx:next 一致，所有 opsx skills 都是 Markdown 指令。Coordinator 逻辑由 AI agent 执行，Worker 逻辑通过 Agent tool 的 prompt 传递。

### 3. 批次执行模型

**选择**: 严格按批次（batch）执行，一批全部完成后才开始下一批

**理由**:
- 依赖关系要求：后续批次可能依赖前序批次的产出
- 人工审阅的自然边界：每批 propose 完成后统一审阅
- 简化状态管理：不需要跟踪部分完成的混合状态

### 4. 人工审阅是顺序的、强制的

**选择**: Proposals 完成后，人必须逐个审阅，不可跳过

**理由**: 风险前置原则。AI 编码就像盖楼，一开始歪了还能纠正，走远了就是灾难。审阅阶段是顺序的（人是瓶颈但这是有意的），apply 阶段才并行（这是真正省时间的地方）。

### 5. Worker Prompt 设计

**选择**: Coordinator 在启动 Worker 时，将完整的 propose/apply 指令嵌入 prompt

**理由**: Worker（subagent）没有访问 slash commands 的能力，所以需要把 opsx:propose 和 opsx:apply 的核心步骤直接写进 prompt。Worker prompt 包含：
- Feature Spec 上下文（feature 目标、当前 change 的 summary/scope）
- OpenSpec CLI 命令序列（new change → instructions → create artifacts）
- 已完成依赖 changes 的信息
- 歧义处理指令：遇到不确定的问题时返回而非自行决策

### 6. Worktree 成果合并

**选择**: Apply 完成后，Worker 的 worktree 自动返回 branch 信息，Coordinator 引导合并

**理由**: Agent tool 的 worktree 模式会返回 worktree path 和 branch。Coordinator 可以用 `git merge` 或 `git cherry-pick` 将成果合入主分支。合并冲突时暂停让人处理。

## Risks / Trade-offs

- **Worker 上下文有限** → Worker 的 prompt 需要足够详细，包含所有必要指令和上下文。Mitigation: 复用 opsx:propose 的核心步骤模板。
- **Worktree 合并冲突** → 虽然按批次隔离降低了冲突概率（同批次的 changes 应该 touch 不同文件），但仍可能发生。Mitigation: Coordinator 检测冲突并暂停让人介入。
- **仅 Claude Code 可用** → 限制了使用范围。Mitigation: opsx:next 仍可用于所有环境，两种模式自由切换。
- **Skill 较长** → team.md 内容会比其他 skill 长（包含 Worker prompt 模板）。Mitigation: 结构清晰，分为 Coordinator 流程和 Worker prompt 模板两大块。
