## Why

当前 `gsf diff` 通过 AST 将 git diff 映射到函数级别，输出变更后的完整函数代码。这个设计有两个根本问题：

1. **有损**：`MapDiffToFunctions` 只遍历 `*ast.FuncDecl`，漏掉 FuncLit（cobra RunE 等）、包级变量、init 函数以外的修改。测试文件的变更也可能遗漏。自举验证已证实此问题。
2. **多余**：`git diff` 本身就是完备的变更信息源，支持灵活的范围控制（staged、commit、branch diff），不会遗漏任何改动。`gsf diff` 做了一个不需要解决的问题 —— reviewer 需要看到的是真实的 diff（什么改了），而不是变更后的完整函数。

gsf 的真正独特价值在于 `trace` 和 `callers` —— 提供 AI 无法从 git diff 获取的调用链上下文。Review skill 应该直接读 git diff 获取完整变更，按需调用 gsf trace/callers 补充结构上下文。

## What Changes

- **删除 `gsf diff` 命令**：移除 `internal/cmd/diff.go`
- **删除 `gsf diff` 依赖的 review 模块代码**：移除 `internal/review/` 下的 diff 解析、函数映射、代码提取等功能（`diff.go`、`mapper.go`、`extract.go` 及相关测试）
- **删除 `gsf:diff` skill**：移除 `skills/gsf-diff.md` 和对应 embed 文件
- **更新 `gsf:review` skill**：改为基于 `git diff` 获取变更，`gsf trace`/`gsf callers` 补充上下文，AI 按动线组织 review 并展示真实 diff 代码

## Capabilities

### New Capabilities

- `review-skill-v2`: 重写 `/gsf:review` skill，基于 git diff + gsf trace/callers 的 review 流程

### Modified Capabilities

（无既有 spec 需要修改）

## Impact

- **删除的代码**：`internal/cmd/diff.go`、`internal/review/diff.go`、`internal/review/mapper.go`、`internal/review/extract.go`、`internal/review/extract_test.go`、`skills/gsf-diff.md`、`internal/cmd/embed_data/skills/gsf-diff.md`
- **保留的代码**：`internal/cmd/callers.go`、`internal/cmd/trace.go`、`internal/ast/` 全部保留
- **Skill 变更**：`skills/gsf-review.md` 和 `internal/cmd/embed_data/skills/gsf-review.md` 需重写
- **无外部 API 或依赖变更**
