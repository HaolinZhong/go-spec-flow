## Why

当前 `/gsf:review` 是一个 AI skill（Markdown 文档），依赖 AI 读代码后生成总结。两个根本问题：

1. **AI 倾向于总结代码而非展示代码** — 无论 prompt 怎么写，AI 都会压缩代码、扩展分析，用户看到的是 AI 的读后感
2. **CLI 不适合展示大量代码** — 几千行代码在终端里滚动，交互极差

需要从"AI prompt 工程"转向"工具生成"：gsf 确定性地构建 flow tree + 关联代码，渲染成可交互的 HTML，在浏览器中展示。

## What Changes

- **新增 `gsf review` 命令**：构建 flow tree，关联源码/diff，渲染自包含 HTML 文件
  - Diff 模式：`gsf review --commit HEAD~3..HEAD` — flow tree 按变更组织，代码显示 +/- diff
  - Codebase 模式：`gsf review --codebase --entry "internal/ast"` — flow tree 按调用链组织，代码显示完整源码
  - 输出单个 HTML 文件，浏览器打开即用
  - HTML 交互：左侧可折叠 flow tree，右侧代码面板（带行号、语法高亮）
- **保留 `/gsf:review` skill** 但简化为引导用户运行 `gsf review` 命令
- **未来可演进**：静态 HTML → `gsf serve` 本地 HTTP server → WebSocket 实时交互

## Capabilities

### New Capabilities

- `review-html-render`: `gsf review` 命令，构建 flow tree + 渲染 HTML
- `flow-tree-builder`: flow tree 数据结构和构建逻辑（diff 模式 + codebase 模式）

### Modified Capabilities

（无）

## Impact

- **新增代码**：`internal/review/` 包（flow tree 构建 + HTML 渲染）、`internal/cmd/review.go`
- **新增依赖**：Go `html/template`（标准库，无外部依赖）、highlight.js CDN（语法高亮）
- **Skill 变更**：`skills/gsf-review.md` 简化为引导运行 `gsf review`
- **HTML 模板**：内嵌在 Go 代码中（`embed`），单文件自包含输出
