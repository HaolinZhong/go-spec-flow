## Context

自举 review 暴露了 `ai-driven-flow-review` 变更中引入的三个 bug 和一个体验问题。这些问题直接影响工具的实用性 — `gsf callers` 在 CLI 项目上几乎无法工作，`gsf diff` 在常见开发流程中需要先 commit 才能使用。

当前代码状态：
- `internal/ast/callers.go`: `FindCallers` 只遍历 `FuncDecl`，漏掉 `GenDecl` 中的 `FuncLit`
- `internal/ast/callgraph.go`: `Tracer.resolveCallTarget` 方法（私有）
- `internal/ast/callers.go`: `resolveCallTargetStatic` 独立函数，与上者逻辑完全重复
- `internal/review/extract.go`: `findFuncDeclByName` 只匹配 Name，不匹配 Receiver
- `internal/review/diff.go`: `GetDiff` 只调 `git diff --unified=0`，不支持 staged

## Goals / Non-Goals

**Goals:**
- 修复 FindCallers 遗漏 package-level FuncLit 调用（cobra 模式）
- 修复 findFuncDeclByName 同名函数/方法歧义
- 统一 resolveCallTarget 逻辑，消除重复
- gsf diff 支持 staged changes 和 untracked 新文件
- 自举验证通过：在 gsf 自身上运行完整 review 流程无问题

**Non-Goals:**
- 不做 FindCallers 对接口动态分派的支持（已知局限，不在此次范围）
- 不改变 gsf diff 的 JSON/YAML 输出结构（向后兼容）
- 不做 callers 多层递归（设计上是 AI 控制递归）

## Decisions

### D1: FindCallers 增加 GenDecl + FuncLit 遍历

**问题**: 在 Go 中，cobra 命令的 `RunE` 字段是 package-level 变量声明里的 function literal：
```go
var diffCmd = &cobra.Command{
    RunE: func(cmd *cobra.Command, args []string) error {
        review.ExtractDiffEntries(...)  // ← FindCallers 找不到这个调用
    },
}
```
这是 `*ast.GenDecl` → `*ast.ValueSpec` → `*ast.CompositeLit` → `*ast.FuncLit`，不是 `FuncDecl`。

**方案**: 在 `FindCallers` 中，除了遍历 `FuncDecl`，增加一个 pass 遍历所有 `GenDecl`，用 `ast.Inspect` 递归查找其中的 `FuncLit` 节点，对每个 `FuncLit.Body` 执行同样的 call expr 匹配逻辑。

caller 命名策略：对 `FuncLit`，使用其所在的变量名 + "(init)" 作为 caller name（例如 `diffCmd(init)`），因为 FuncLit 没有自己的函数名。

**替代方案**: 对整个 file 做 `ast.Inspect` 而不区分 FuncDecl/GenDecl。放弃是因为这样会丢失 caller function name 信息 — 我们需要知道"谁"在调用目标函数。

### D2: 统一 resolveCallTarget

**方案**: 将 `resolveCallTargetStatic` 提升为 `internal/ast` 包的公共函数 `ResolveCallTarget(pkg, call)`。同时修改 `Tracer.resolveCallTarget` 为调用此公共函数的委托。

这样 trace 和 callers 保证使用完全相同的调用解析逻辑。

### D3: findFuncDeclByName 匹配 receiver

**方案**: 给 `findFuncDeclByName` 增加 `receiver string` 参数。匹配时：
- 如果 receiver 非空，要求 `FuncDecl.Recv` 的类型名一致
- 如果 receiver 为空，要求 `FuncDecl.Recv` 为 nil（standalone function）

调用方 `ExtractDiffEntries` 已经有 `cf.Receiver` 信息，直接传入。

### D4: gsf diff 支持 staged + untracked

**方案**: 修改 `GetDiff` 的行为：
1. 无参数（无 `--commit`/`--base`）时：先尝试 `git diff --staged --unified=0`，如果有内容则使用 staged diff；否则 fallback 到 `git diff --unified=0`（unstaged）
2. 新增 `--include-untracked` flag：运行 `git ls-files --others --exclude-standard` 获取未跟踪文件列表，对每个 `.go` 文件生成一个 `IsNew: true` 的 `FileDiff`（整个文件作为一个 hunk）

这样常见流程不需要先 commit：`git add` 后直接 `gsf diff` 即可。

## Risks / Trade-offs

- **[FuncLit caller naming]** `diffCmd(init)` 这种名字不够精确（可能一个变量声明里有多个 FuncLit）→ 可接受，一个变量通常只有一个 RunE/Run
- **[staged vs unstaged 歧义]** 用户可能困惑为什么有时候看 staged 有时候看 unstaged → 缓解：text 输出显示当前使用的 diff 模式（"Showing staged changes" / "Showing unstaged changes"）
- **[性能]** FindCallers 额外遍历 GenDecl → 影响可忽略，文件数量级不变
