## Why

Review HTML 目前是只读展示工具。用户完成代码 review 后，发现问题需要告诉 AI 修改时，只能手动描述文件名和行号，效率低且容易出错。需要一个从"review 发现问题"到"AI 执行修改"的闭环机制：用户在 review 页面上直接对代码行添加评论，评论自动保存到磁盘，然后通过新 skill 命令让 AI 读取并执行修改。

## What Changes

- Review HTML 新增行级评论 UI：点击行号弹出评论框，输入自由文本评论
- 新增 `gsf review --serve` 模式：启动本地 HTTP 服务器，提供 HTML 页面和评论保存 API，替代生成静态文件+手动打开的方式
- 定义评论文件格式（JSON），包含文件路径、行号、代码上下文、评论内容
- 新增 `/gsf:fix` skill：读取评论文件，驱动 AI 逐条理解评论意图并执行代码修改

## Capabilities

### New Capabilities
- `review-comments`: Review HTML 的行级代码评论能力，包括评论 UI 交互、本地服务器保存、评论文件格式定义
- `review-fix`: 基于 review 评论文件驱动 AI 执行代码修改的 skill

### Modified Capabilities

## Impact

- `internal/review/templates/review.html` — 新增评论 UI（行点击、评论框、评论显示）
- `internal/review/server.go` — 新增本地 HTTP 服务器（serve HTML + 评论保存 API）
- `internal/cmd/review.go` — 新增 `--serve` flag
- `skills/gsf-fix.md` + `internal/cmd/embed_data/skills/gsf-fix.md` — 新增 fix skill
- `internal/cmd/init.go` — 注册 gsf-fix skill 文件
