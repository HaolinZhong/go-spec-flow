# PRD: opsx:team — 并行执行 Feature Spec Changes

## 背景

当前 `opsx:decompose` 可以将大需求拆为多个 OpenSpec changes（Feature Spec），`opsx:next` 可以按依赖顺序逐个推进。但对于大 feature（5+ changes），顺序执行太慢。当多个 changes 之间没有依赖关系时，应该可以并行执行。

## 目标

提供 `opsx:team` skill，实现 Coordinator + Workers 模式的并行执行：
- Coordinator 读取 Feature Spec，分析依赖图，识别可并行的 changes
- Workers 在独立 git worktree 中执行各自的 change（propose + apply）
- 人在每一批 propose 完成后逐个审阅，审阅通过后并行 apply
- 按批次推进，直到所有 changes 完成

## 核心流程

```
用户: /opsx:team <feature-name>

Coordinator:
  1. 读 Feature Spec，构建依赖图
  2. 识别当前批次可并行的 changes（status=pending, 依赖已满足）
  3. 展示批次计划，用户确认

Batch Loop:
  4. 并行启动 Workers（每个 worker 在独立 worktree）
  5. Workers 并行执行 propose
  6. 所有 proposals 就绪后，Coordinator 汇报
  7. 用户逐个审阅每个 proposal（approve / request changes）
  8. 审阅通过的 changes → Workers 并行 apply
  9. Apply 完成后更新 Feature Spec 状态
  10. 识别下一批 unblocked changes，继续

结束条件: 所有 changes completed
```

## 关键设计约束

### Worktree 隔离
- 每个 Worker 在独立 git worktree 中工作，避免冲突
- Worktree 基于当前 HEAD 创建
- 新批次开始前，Workers 需要 rebase 上一批的成果

### 人工审阅前置（风险前置）
- 必须逐个审阅每个 proposal，不能跳过
- 审阅阶段是顺序的（人是瓶颈，但这是有意的）
- Apply 阶段是并行的（这是真正省时间的地方）

### Proposal 否决处理
- 小调：修改当前 proposal，不影响 Feature Spec
- 大改：暂停 team，回到 Feature Spec 调整拆解，重新启动

### 状态管理
- Feature Spec 是 single source of truth
- 每个 change 状态: pending → proposed → completed
- Coordinator 负责更新状态

### 并行是可选的
- opsx:team 是增强选项，不替代 opsx:next
- 不想并行的用户继续用 opsx:next 顺序执行
- Feature Spec 格式完全兼容，两种模式自由切换

## 参考

- Gastown (steveyegge/gastown): Mayor/Polecat 模式
- 我们不需要搬其基础设施，Feature Spec 本身就是持久化工作清单
- 与 Gastown 的区别：我们有结构化的 Feature Spec + OpenSpec 生命周期管理

## 非目标

- 不做自动分配 changes 给 agent（Coordinator 按依赖图决定）
- 不做自动 merge 到 main（人最终决定）
- 不做自动重试失败的 changes
- 不做跨机器的分布式执行（单机多 worktree 即可）
