## Why

自举测试暴露了 M5 Flow Review 的核心局限：`gsf review` 硬编码 Hertz 路由作为入口发现策略，导致 CLI 项目（包括 gsf 自身）的所有变更都被归为 "Standalone Changes"，无法生成有意义的 flow review。根本原因是把"智能决策"（选入口、定动线）硬编码在了 Go 代码里，而这恰恰是 AI 擅长、静态分析不擅长的事。

## What Changes

- **新增 `gsf diff` 命令**：解析 git diff 并映射到函数/方法级别，输出每个变更函数的完整代码片段（package, name, file, line range, code body），复用现有 `internal/review/diff.go` + `mapper.go`
- **新增 `gsf callers` 命令**：给定一个函数，查找它的直接调用者（一层），基于新的 AST 反向索引能力
- **改造 `gsf trace` 命令**：支持 `--pkg X --func Y` 参数直接指定函数入口（不只是 `--route`），使 AI 可以对任意函数追踪调用链
- **新增 `/gsf:review` skill**：AI 编排的 flow review，调用 gsf diff/trace/callers 工具，由 AI 判断变更性质、决定 review 动线、组织输出文档
- **删除 `gsf review` 命令**：移除硬编码 Hertz 的 flow review 实现
- **删除 `internal/review/flow.go`**：移除 `BuildFlowReview()` 及其 Hertz 路由依赖逻辑

## Capabilities

### New Capabilities
- `diff-command`: `gsf diff` 命令 — 函数级变更分析，输出变更函数列表及完整代码片段
- `callers-command`: `gsf callers` 命令 — 函数直接调用者查找（一层），基于 AST 反向索引
- `ai-review-skill`: `/gsf:review` skill — AI 编排的 flow-based code review

### Modified Capabilities
- `cli-framework`: 新增 `gsf diff`、`gsf callers` 子命令，删除 `gsf review` 子命令，`gsf trace` 新增 `--pkg/--func` 参数

## Impact

- **删除代码**: `internal/review/flow.go`, `cmd/gsf/` 中的 review 命令注册
- **复用代码**: `internal/review/diff.go`, `internal/review/mapper.go` 被 diff 命令复用
- **新增代码**: `internal/review/callers.go`（AST 反向索引），diff/callers/trace 的 CLI 注册
- **Skill 文件**: `.claude/skills/` 下新增 review skill markdown
- **对外接口变更**: `gsf review` 被移除（**BREAKING**），替换为 `/gsf:review` skill + `gsf diff`/`gsf callers` 工具命令
