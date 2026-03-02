## Problem

FlowTree 中同一个函数在 trace 调用链中重复出现多次。根因：
1. **同文件兄弟重复**：函数 A 和函数 B 同属一个文件，A 调用 B，B 既是文件的顶层函数节点，又作为 A 的 trace child 出现
2. **跨分支 trace 重复**：工具函数（如 contains、readFileLines）被多个函数调用，每个 trace 分支都创建完整副本

## Tasks

### 1. Builder 去重 — 同文件兄弟过滤

- [x] 1.1 `builder.go` buildFuncNodesFromDiff: 收集当前文件所有已声明函数名为 `siblingNames` set，在创建 trace children 时传入
- [x] 1.2 `builder.go` callNodeToFlowNode 内置兄弟过滤（siblings 参数），nil 返回值在调用处跳过
- [x] 1.3 `builder.go` BuildCodebaseTree: 同样对 codebase 模式应用兄弟去重

### 2. Builder 去重 — 跨分支 seen 集合

- [x] 2.1 `builder.go` buildFuncNodesFromDiff: 创建 per-file `seen` map，在 callNodeToFlowNode 递归中传入，已 seen 的函数只保留最小引用节点（无 code、无 children）
- [x] 2.2 `builder.go` 修改 `callNodeToFlowNode` 签名，增加 `seen map[string]bool` 参数，已 seen 的函数返回精简节点
- [x] 2.3 `builder.go` BuildCodebaseTree: 同样使用 per-file seen 去重

### 3. 验证

- [x] 3.1 `go build` + `go test ./...`
- [x] 3.2 自举验证：143→32 节点，27→3 重复（均为精简引用 ref- 节点，符合预期）
