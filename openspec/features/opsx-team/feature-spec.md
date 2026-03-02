# Feature: opsx:team — 并行执行 Feature Spec Changes

## PRD Source
[prd.md](./prd.md)

## Overview
提供 `opsx:team` skill，利用 Claude Code 原生 Agent Teams 能力，实现 Coordinator + Workers 模式并行执行 Feature Spec 中的 changes。仅支持 Claude Code 环境；弱模型/工具环境继续使用 `opsx:next` 顺序执行。

## Dependency Diagram
```
[A: team-coordinator-skill]
        (独立，无依赖)
```

## Changes

### A: team-coordinator-skill
- **Summary**: 创建 `opsx:team` skill，实现完整的 Coordinator + Workers 并行执行流程。Coordinator 读取 Feature Spec，分析依赖图，按批次调度 Workers 并行 propose；人逐个审阅后并行 apply；利用 Claude Code 的 Agent tool 启动 Workers 在独立 worktree 中工作，支持 plan approval mode 让人介入。
- **Scope**: `.claude/commands/opsx/team.md`, `.claude/skills/openspec-team/SKILL.md`
- **Depends on**: []
- **Status**: completed
- **Key decisions**:
  - 利用 Claude Code 原生 Agent tool（subagent_type + isolation: "worktree"）而非自建 tmux 管理
  - Workers 在 worktree 隔离中工作，避免文件冲突
  - 人必须逐个审阅每个 proposal（风险前置）
  - Workers 遇到歧义时暂停等待人介入，不自行决策
  - Feature Spec 是 single source of truth，Coordinator 负责更新状态
  - 仅 Claude Code 环境可用，不适配弱模型工具
