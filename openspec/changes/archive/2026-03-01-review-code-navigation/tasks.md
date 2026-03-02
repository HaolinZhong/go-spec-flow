## 1. HTML 模板 — 节点索引与跳转

- [x] 1.1 `review.html` JS：初始化时递归遍历 flowData 构建 `nodeIndex` (Map<label, {node, labelEl}>)，`createTreeNode` 中存储 labelEl 引用
- [x] 1.2 `review.html` JS：`navigateToNode(funcName)` 函数 — 查索引、展开父节点链、高亮、showCode、滚动可见
- [x] 1.3 `review.html` JS：`showCode` 中 highlight.js 渲染后，对代码 DOM 做后处理 — 遍历文本节点，将匹配 nodeIndex 的函数名包裹为 `<span class="code-link">`
- [x] 1.4 `review.html` JS：为 `code-link` 元素添加 click 事件，调用 `navigateToNode`
- [x] 1.5 `review.html` CSS：`.code-link` 样式（下划线、颜色、cursor pointer）

## 2. Skill 指令 — 步骤节点带代码

- [x] 2.1 `skills/gsf-review.md` Step 3：更新指令 — 步骤节点 `code` 设置为入口函数代码（非空）
- [x] 2.2 `skills/gsf-review.md` Step 3：更新 JSON 示例 — 步骤节点 `code` 字段非空

## 3. 同步与构建

- [x] 3.1 同步 `internal/cmd/embed_data/skills/gsf-review.md`
- [x] 3.2 `go build` 重新构建 gsf 二进制
- [x] 3.3 `gsf init` 安装更新后的 skill

## 4. 验证

- [x] 4.1 `go test ./...` 全部通过
- [x] 4.2 `/gsf:review` 自举验证：步骤节点显示代码、函数名可点击跳转到子动线节点
