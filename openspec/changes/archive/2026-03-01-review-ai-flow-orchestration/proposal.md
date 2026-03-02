## Why

`gsf review` 生成的 HTML 按 package/file/function 结构组织代码，用户看到的只是文件系统的映射。缺少"动线"概念 — 没有 AI 对代码架构的解读、没有每个节点的角色说明、没有 review 的引导。用户面对一大堆代码仍然不知道从哪里看起、为什么这些函数放在一起。

核心矛盾：gsf 做了"精确的脏活"（提取代码），但"智能的组织"（动线编排 + 解释）完全缺失。需要让 AI 参与到 review 的组织环节。

## What Changes

- **FlowTree/FlowNode 增加 description 字段** — 支持在树和节点上附加 AI 生成的解释文字
- **`gsf review --json` 输出模式** — 输出结构化 JSON 而非直接生成 HTML，供 AI 读取和编排
- **`gsf review --render <file>` 渲染模式** — 读取 AI 编排后的 JSON，渲染为 HTML
- **HTML 模板增加描述展示区** — 动线标题下显示说明，节点代码上方显示注释
- **review skill 重写** — 从"运行一条命令"变为三步流程：gsf 提取 JSON → AI 编排动线 → gsf 渲染 HTML

## Capabilities

### New Capabilities
- `review-json-pipeline`: gsf review 的 JSON 输入/输出管道 — `--json` 导出结构数据，`--render` 消费 AI 编排后的 JSON 生成 HTML

### Modified Capabilities
（无）

## Impact

- `internal/review/tree.go` — FlowTree/FlowNode 增加 Description 字段
- `internal/review/render.go` + `templates/review.html` — HTML 模板需展示描述文字
- `internal/cmd/review.go` — 增加 `--json` 和 `--render` flag
- `skills/gsf-review.md` — 重写为三步编排流程
- `internal/cmd/embed_data/skills/gsf-review.md` — 同步
