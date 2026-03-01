## 1. 数据结构 — Description 字段

- [x] 1.1 `internal/review/tree.go`：FlowTree 和 FlowNode 增加 `Description string` 字段（json omitempty）

## 2. JSON 输出模式

- [x] 2.1 `internal/cmd/review.go`：增加 `--json` flag，输出 FlowTree JSON 到 stdout
- [x] 2.2 验证 diff 模式 `--json` 输出有效 JSON
- [x] 2.3 验证 codebase 模式 `--json` 输出有效 JSON

## 3. JSON 渲染模式

- [x] 3.1 `internal/cmd/review.go`：增加 `--render <file>` flag，从 JSON 文件读取 FlowTree 并渲染 HTML
- [x] 3.2 `--render` 与 `--commit/--base/--codebase` 互斥校验

## 4. HTML 模板 — 描述展示

- [x] 4.1 `internal/review/templates/review.html`：FlowTree.Description 显示为标题下方的概述段落
- [x] 4.2 `internal/review/templates/review.html`：FlowNode.Description 显示为代码面板上方的注释卡片
- [x] 4.3 Description 为空时不显示空白区域

## 5. Review Skill 重写

- [x] 5.1 `skills/gsf-review.md`：重写为三步流程（gsf --json → AI 编排 → gsf --render）
- [x] 5.2 同步 embed 副本 + 重新构建 + gsf init

## 6. 自举验证

- [x] 6.1 `/gsf:review` 全流程验证：codebase 模式，AI 编排动线，HTML 输出含描述
- [x] 6.2 `go test ./...` 全部通过
