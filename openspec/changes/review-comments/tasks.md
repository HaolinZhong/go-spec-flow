## 1. HTML 评论 UI

- [x] 1.1 添加评论相关 CSS 样式（评论指示器、评论输入框、评论气泡、行号悬停效果）
- [x] 1.2 实现 JavaScript 评论状态管理（comments Map，key: `file:line`，存储评论文本）
- [x] 1.3 实现行号点击事件：source 模式下点击 `.line-num` 弹出内联评论输入框
- [x] 1.4 实现行号点击事件：diff 模式下点击 `.diff-line-num`（仅 `+` 和上下文行可评论，`-` 行不可评论）
- [x] 1.5 实现评论的编辑和删除（点击已有评论行打开编辑，清空即删除）
- [x] 1.6 已有评论的行号添加视觉指示器（高亮或图标）
- [x] 1.7 Header 区域显示评论总数计数器

## 2. 本地 HTTP 服务器

- [x] 2.1 新建 `internal/review/server.go`：实现 `StartServer(tree *FlowTree, dir string, port int) error`，启动 HTTP 服务器
- [x] 2.2 实现 `GET /` 端点：内存渲染 review HTML 返回
- [x] 2.3 实现 `POST /comments` 端点：接收评论 JSON，写入项目根目录 `review-comments.json`
- [x] 2.4 `internal/cmd/review.go`：新增 `--serve` flag，调用 StartServer 替代生成静态文件

## 3. 评论自动保存与导出

- [x] 3.1 HTML 中实现 `saveComments()` 函数：检测是否 serve 模式，是则 POST 到 `/comments`
- [x] 3.2 每次评论增删改自动调用 `saveComments()`
- [x] 3.3 静态文件模式下的 fallback：提供下载按钮，生成 JSON 文件（Blob URL + `<a download>`）
- [x] 3.4 评论文件格式：包含 reviewTitle、mode、exportedAt、comments 数组（file、line、codeContext、comment）

## 4. `/gsf:fix` Skill

- [x] 4.1 新建 `skills/gsf-fix.md`：定义 skill 流程（读取评论文件 → 按文件分组 → 逐条执行修改）
- [x] 4.2 复制到 `internal/cmd/embed_data/skills/gsf-fix.md`
- [x] 4.3 `internal/cmd/init.go`：在 skillFiles 中注册 gsf-fix.md

## 5. 验证

- [x] 5.1 `go build` + `go test ./...`
- [x] 5.2 端到端测试：`gsf review --render /tmp/gsf-flow.json --serve`，在浏览器中添加评论，验证 `review-comments.json` 被正确生成
