## Why

自举 review（用 gsf 工具 review gsf 自身代码）暴露了三个影响工具可靠性的问题：`FindCallers` 无法发现 cobra RunE 模式中的调用、`findFuncDeclByName` 同名函数/方法歧义、`resolveCallTarget` 逻辑重复可能导致 trace 与 callers 结果矛盾。同时 `gsf diff` 不支持 staged/untracked 文件，导致 review 流程必须先 commit 才能工作。这些问题必须在推广使用前修复。

## What Changes

- **修复 `FindCallers` 遗漏 package-level FuncLit 调用**：除了遍历 `FuncDecl`，增加对 `GenDecl`（变量声明）中嵌套 `FuncLit` 的遍历，覆盖 cobra `RunE`/`Run` 等模式
- **修复 `findFuncDeclByName` 同名歧义**：匹配时同时校验 receiver 类型，避免同 package 内同名函数和方法返回错误结果
- **消除 `resolveCallTarget` 重复**：将 `Tracer.resolveCallTarget` 和 `resolveCallTargetStatic` 统一为一个共享函数，确保 trace 和 callers 行为一致
- **`gsf diff` 支持 staged 和 untracked 文件**：无参数时检测 staged changes（`git diff --staged`），新增 `--include-untracked` 支持未跟踪的新文件
- **自举验证作为回归测试**：在 go-spec-flow 项目自身上运行完整 review 流程（gsf diff → gsf callers → gsf trace），验证所有修复

## Capabilities

### New Capabilities
（无新能力，全部为现有能力的修复和增强）

### Modified Capabilities
- `callers-command`: FindCallers 须覆盖 package-level FuncLit 中的调用
- `diff-command`: gsf diff 须支持 staged changes 和 untracked 文件；findFuncDeclByName 须匹配 receiver

## Impact

- **修改文件**: `internal/ast/callers.go`（FuncLit 遍历）、`internal/ast/callgraph.go`（提取共享 resolveCallTarget）、`internal/review/extract.go`（receiver 匹配）、`internal/review/diff.go`（staged + untracked 支持）、`internal/cmd/diff.go`（新 flag）
- **测试更新**: `internal/ast/callers_test.go`、`internal/review/extract_test.go` 增加回归用例
- **自举验证**: 完整 review 流程在自身项目上运行通过
