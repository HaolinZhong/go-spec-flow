## 1. 数据结构 — FlowNode 增加 Diff 字段

- [x] 1.1 `tree.go`: FlowNode 添加 `Diff string json:"diff,omitempty"` 字段，与 Code 分离
- [x] 1.2 `diff.go`: 添加 `extractFuncDiff(fileDiff string, lineStart, lineEnd int) string` 函数 — 从文件 diff 中提取与函数行号范围重叠的 hunk

## 2. Builder — 同时填充 Diff 和 Code

- [x] 2.1 `builder.go` BuildDiffTree: 文件节点 Diff=df.Content, Code=readEntireFile(实际文件)
- [x] 2.2 `builder.go` buildFuncNodesFromDiff: 函数节点 Diff=extractFuncDiff(df.Content, start, end)

## 3. HTML 模板 — Toggle + Diff 渲染修复

- [x] 3.1 `review.html` header: 添加 toggle 开关按钮（"Show Diff" / "Show Source"），codebase 模式下禁用
- [x] 3.2 `review.html` showCode: 根据 toggle 状态选择渲染 diff 或 source，node 无 diff 时 toggle 置灰
- [x] 3.3 `review.html` renderDiff: 修复行号显示 — 从 hunk header 解析新文件行号并显示
- [x] 3.4 `review.html`: 修复 highlight.js 调用 — 非 diff 模式或 toggle=source 时才 highlight

## 4. Skill 更新

- [x] 4.1 `skills/gsf-review.md`: 更新说明 — diff 模式下节点同时携带 diff 和 source，HTML 支持 toggle

## 5. 同步与构建

- [x] 5.1 同步 `internal/cmd/embed_data/skills/gsf-review.md`
- [x] 5.2 `go build` 构建 gsf
- [x] 5.3 `gsf init`

## 6. 验证

- [x] 6.1 `go test ./...`
- [x] 6.2 自举测试：diff 模式查看 review 包变更，验证 toggle、行号、diff 高亮
