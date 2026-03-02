## 1. 变更函数提取

- [x] 1.1 新增 `ChangedFunc` 结构体（pkg, name, file, lineStart, lineEnd, code, funcDiff），用于收集变更函数
- [x] 1.2 新增 `collectChangedFuncs(project, dir, diffs) []ChangedFunc`：遍历 diff 中的 Go 文件，找到所有 `funcDiff != ""` 的函数
- [x] 1.3 单元测试：验证只有实际有 diff 的函数被收集，未变更函数被过滤

## 2. 调用图构建与入口检测

- [x] 2.1 新增 `buildCallGraph(project, changedFuncs) CallGraph`：对每个变更函数调用 Tracer.Trace()，在 trace 树中搜索其他变更函数，记录有向边和桥接函数
- [x] 2.2 新增 `findEntries(graph) []ChangedFunc`：在 DAG 中找入度为 0 的变更函数作为入口
- [x] 2.3 新增 `findBridges(graph) []BridgeFunc`：提取调用链中连接变更函数的未变更中间函数
- [x] 2.4 单元测试：验证 A→B 调用关系正确识别、入口检测正确、桥接函数正确标记、循环调用正确处理

## 3. Flow 树构建

- [x] 3.1 新增 `buildFlowRoots(project, graph, changedFuncs) []*FlowNode`：从入口开始构建调用链 flow 树，包含入口函数 + 桥接函数 + 被调用的变更函数
- [x] 3.2 独立变更函数（不在任何调用链中的）作为独立 flow root 输出，不附带 trace 子节点
- [x] 3.3 非 Go 文件归组到 "Non-code Files" root 节点，每个文件作为子节点
- [x] 3.4 FlowNode 新增 `IsBridge bool` 字段（`json:"isBridge,omitempty"`）

## 4. BuildDiffTree 重构

- [x] 4.1 重构 `BuildDiffTree`：替换现有的按文件遍历逻辑，改为：收集变更函数 → 构建调用图 → 构建 flow roots → 归组非 Go 文件
- [x] 4.2 保持 `--codebase` 模式（`BuildCodebaseTree`）完全不变
- [x] 4.3 端到端验证：`gsf review --commit HEAD --json` 输出按 flow 组织，root 数量显著减少

## 5. Skill 更新

- [x] 5.1 更新 `skills/gsf-review.md` Step 3：适配预组织 JSON，AI 只需添加 description 和可选调整
- [x] 5.2 同步更新 `internal/cmd/embed_data/skills/gsf-review.md`
- [x] 5.3 运行 `gsf init` 重新安装 skills

## 6. 验证

- [x] 6.1 `go build ./...` + `go vet ./...`
- [x] 6.2 `go test ./...` 确保现有测试通过
- [x] 6.3 端到端测试：对本项目执行 `gsf review --commit HEAD --json`，验证输出结构正确
- [x] 6.4 端到端测试：`gsf review --render` + `--serve` 验证 HTML 正确渲染新结构
