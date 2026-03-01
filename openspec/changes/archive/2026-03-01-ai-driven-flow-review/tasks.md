## 1. gsf diff 命令

- [x] 1.1 在 `internal/review/` 中新增函数体提取功能：给定 `ChangedFunc` 和 AST 信息，从源文件读取完整函数代码（利用 `FuncDecl.Pos()/End()` 定位行号范围）
- [x] 1.2 定义 `DiffResult` 输出结构体（package, name, receiver, file, line_start, line_end, is_new, code），实现 `String()` 方法支持 text 格式输出
- [x] 1.3 新增 `internal/cmd/diff.go`，注册 `gsf diff` 子命令，支持 `--commit`、`--base`、`--format` 标志，串联 GetDiff → FilterGoFiles → MapDiffToFunctions → 提取函数体 → 格式化输出
- [x] 1.4 为 gsf diff 编写单元测试：使用 testdata 目录验证函数级 diff 映射和代码提取

## 2. gsf callers 命令

- [x] 2.1 新增 `internal/ast/callers.go`：实现 `FindCallers(project, pkg, funcName)` 函数 — 遍历项目所有函数体 AST，收集 `*ast.CallExpr`，构建反向索引，返回 `[]CallerInfo`（package, name, file, line）
- [x] 2.2 新增 `internal/cmd/callers.go`，注册 `gsf callers` 子命令，支持 `--pkg`、`--func`、`--format` 标志
- [x] 2.3 为 callers 编写单元测试：在 testdata 中构造跨包调用场景，验证反向索引正确性

## 3. 删除硬编码 review

- [x] 3.1 删除 `internal/review/flow.go`（`BuildFlowReview`、`FlowNode`、`FlowEntry` 等类型和函数）
- [x] 3.2 删除 `internal/cmd/review.go`（`gsf review` 命令注册）
- [x] 3.3 清理相关测试：移除 `flow.go` 和 `review` 命令相关的测试用例，确保剩余测试通过
- [x] 3.4 更新 `skills/gsf-trace.md` 的描述（移除对 `gsf review` 的引用，如有）

## 4. /gsf:review Skill

- [x] 4.1 新增 `skills/gsf-review.md`：编写 skill 文件，包含工具介绍（gsf diff/callers/trace）、review 流程编排指引、策略示例（new feature/bugfix/refactor）、输出格式要求
- [x] 4.2 新增 `skills/gsf-diff.md` 和 `skills/gsf-callers.md`：为新命令编写 skill 使用说明
- [x] 4.3 更新 `internal/cmd/init.go`：在 `gsf init` 安装列表中加入新的 skill 文件（gsf-review, gsf-diff, gsf-callers）

## 5. 集成验证

- [x] 5.1 自举测试：在 go-spec-flow 项目自身上运行 `gsf diff` 和 `gsf callers`，验证输出正确性
- [x] 5.2 确保所有现有单元测试通过（`go test ./...`）
- [x] 5.3 更新 resume-state.md 中的命令列表，反映 review → diff/callers 的变化
