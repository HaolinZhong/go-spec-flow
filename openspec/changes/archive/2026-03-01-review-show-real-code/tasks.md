## 1. Flow Tree 数据结构

- [x] 1.1 创建 `internal/review/` 包，定义 `FlowTree`、`FlowNode` 数据结构
- [x] 1.2 实现 `BuildDiffTree(dir, gitRange string) (*FlowTree, error)` — diff 模式的 tree 构建
- [x] 1.3 实现 `BuildCodebaseTree(project, entryPkg string) (*FlowTree, error)` — codebase 模式的 tree 构建

## 2. 代码收集

- [x] 2.1 实现 diff 代码收集：运行 git diff，按文件拆分，关联到 FlowNode
- [x] 2.2 实现源码收集：从 AST 定位函数声明，读取对应行范围的源码，关联到 FlowNode

## 3. HTML 渲染

- [x] 3.1 创建 HTML 模板（Go embed）：左侧 tree + 右侧 code 布局
- [x] 3.2 实现 tree 交互 JS：折叠/展开、点击切换代码面板
- [x] 3.3 实现 diff 着色：+ 行绿底、- 行红底
- [x] 3.4 集成 highlight.js CDN 做 Go 语法高亮（带离线 fallback）
- [x] 3.5 实现 `RenderHTML(tree *FlowTree, w io.Writer) error`

## 4. CLI 命令

- [x] 4.1 创建 `internal/cmd/review.go` — `gsf review` 命令，支持 --commit/--base/--codebase/--entry/--output/--open flags
- [x] 4.2 实现 `--open` flag：生成后调用 `xdg-open` / `open` 打开浏览器

## 5. Skill 更新

- [x] 5.1 简化 `skills/gsf-review.md`：引导用户运行 `gsf review` 命令
- [x] 5.2 同步 embed + 运行 `gsf init`

## 6. 自举验证

- [x] 6.1 `gsf review --commit HEAD --open` — 验证 diff 模式 HTML 输出
- [x] 6.2 `gsf review --codebase --entry "internal/ast" --open` — 验证 codebase 模式 HTML 输出
- [x] 6.3 确保 `go test ./...` 全部通过
