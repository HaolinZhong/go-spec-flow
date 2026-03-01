## 1. 统一 resolveCallTarget

- [x] 1.1 将 `callers.go` 的 `resolveCallTargetStatic` 提升为 `internal/ast` 包公共函数 `ResolveCallTarget(pkg, call) (pkgPath, funcName string)`
- [x] 1.2 修改 `callgraph.go` 的 `Tracer.resolveCallTarget` 为委托调用 `ResolveCallTarget`
- [x] 1.3 验证 `go test ./internal/ast/...` 全部通过（现有 trace 测试不回归）

## 2. 修复 FindCallers — 覆盖 FuncLit

- [x] 2.1 在 `FindCallers` 中增加对 `*ast.GenDecl` 的遍历：遍历 file.Decls 中的 GenDecl，用 `ast.Inspect` 递归找 `*ast.FuncLit`，对每个 FuncLit.Body 执行 call expr 匹配
- [x] 2.2 FuncLit caller naming：从 GenDecl 的 ValueSpec 中提取变量名作为 caller name（如 `diffCmd`），package 为当前包
- [x] 2.3 更新 `callers_test.go`：新增测试用例验证 cobra RunE 模式中的调用能被发现（用 go-spec-flow 自身作为 testdata，或在 sample-app 中增加类似模式）

## 3. 修复 findFuncDeclByName — receiver 匹配

- [x] 3.1 给 `findFuncDeclByName` 增加 `receiver string` 参数，匹配时同时校验 receiver type
- [x] 3.2 更新 `ExtractDiffEntries` 调用处，传入 `cf.Receiver`
- [x] 3.3 更新 `extract_test.go`：新增同名函数+方法场景的测试用例

## 4. gsf diff 支持 staged + untracked

- [x] 4.1 修改 `GetDiff`：无 `--commit`/`--base` 时，先尝试 `git diff --staged --unified=0`，有内容则使用；否则 fallback 到 `git diff --unified=0`
- [x] 4.2 新增 `GetUntrackedFiles(dir)` 函数：运行 `git ls-files --others --exclude-standard`，过滤 `.go` 文件，为每个文件生成 `FileDiff{Path, IsNew: true, Hunks: 全文件}`
- [x] 4.3 在 `internal/cmd/diff.go` 新增 `--include-untracked` flag，将 untracked 文件合并到 diff 结果中
- [x] 4.4 text 输出时显示 diff 模式提示（"Showing staged changes" / "Showing unstaged changes" / "Including N untracked files"）

## 5. 自举验证

- [ ] 5.1 确保 `go test ./...` 全部通过
- [ ] 5.2 自举 review 流程：commit 后运行 `gsf diff --commit HEAD`，验证所有修改函数正确列出
- [ ] 5.3 自举 callers 验证：运行 `gsf callers --pkg .../internal/review --func ExtractDiffEntries`，确认能找到 `internal/cmd` 中的调用者
- [ ] 5.4 自举 staged diff 验证：stage 一个文件变更，运行 `gsf diff`，确认能输出 staged changes
