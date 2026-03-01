## ADDED Requirements

### Requirement: Review skill 使用 git diff 获取变更

`/gsf:review` skill SHALL 指导 AI 使用原生 `git diff` 命令获取代码变更，而非依赖 `gsf diff`。

skill SHALL 支持以下 review 范围：
- 默认：staged 变更（有的话），否则 unstaged 变更
- 指定 commit：`git diff <commit>~1 <commit>`
- 指定 base branch：`git diff <base>...HEAD`
- 指定 commit 范围：`git diff <from> <to>`

#### Scenario: Review 最近一次 commit
- **WHEN** 用户运行 `/gsf:review` 并指定 review 最近 commit
- **THEN** AI 运行 `git diff HEAD~1 HEAD` 获取完整变更，输出包含所有被修改文件的真实 diff 内容

#### Scenario: Review 分支相对于 main 的变更
- **WHEN** 用户运行 `/gsf:review` 并指定 base branch 为 main
- **THEN** AI 运行 `git diff main...HEAD` 获取分支全部变更

#### Scenario: Review staged 变更
- **WHEN** 用户运行 `/gsf:review` 未指定范围，且有 staged 变更
- **THEN** AI 运行 `git diff --staged` 获取 staged 变更

### Requirement: Review 展示真实 diff 代码

Review 文档的每个节点 SHALL 展示对应的 git diff 片段（包含 +/- 行），而非仅展示变更后的完整代码。

#### Scenario: 函数修改的 diff 展示
- **WHEN** review 文档展示一个被修改的函数
- **THEN** 展示的代码包含 git diff 的 +/- 行，让 reviewer 能看到具体改了什么

#### Scenario: 新文件的 diff 展示
- **WHEN** review 文档展示一个新增的文件
- **THEN** 展示完整的新增代码（全部为 + 行）

### Requirement: Review 按逻辑动线组织

Review 文档 SHALL 按逻辑动线（而非文件顺序）组织变更，动线由 AI 根据变更性质判断。

AI SHALL 按需调用 `gsf callers` 和 `gsf trace` 补充调用链上下文。

#### Scenario: Bugfix 的 review 动线
- **WHEN** 变更性质为 bugfix
- **THEN** review 按影响面组织：修复点 → callers（影响面） → 相关测试

#### Scenario: 新 feature 的 review 动线
- **WHEN** 变更性质为新功能
- **THEN** review 按请求流组织：入口 → trace 调用链 → 各实现节点

### Requirement: 删除 gsf diff 命令

`gsf diff` 命令 SHALL 被完全移除，包括：
- `internal/cmd/diff.go`
- `internal/review/diff.go`、`mapper.go`、`extract.go` 及其测试
- `skills/gsf-diff.md` 和 `internal/cmd/embed_data/skills/gsf-diff.md`

`gsf trace`、`gsf callers`、`gsf analyze` 等命令 SHALL 不受影响。

#### Scenario: gsf diff 命令不再可用
- **WHEN** 用户运行 `gsf diff`
- **THEN** 输出 unknown command 错误

#### Scenario: gsf callers 和 trace 不受影响
- **WHEN** 用户运行 `gsf callers` 或 `gsf trace`
- **THEN** 命令正常工作，行为不变
