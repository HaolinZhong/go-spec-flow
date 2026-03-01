## Why

主动线的步骤节点目前没有代码（`code: ""`），只有文字描述，太抽象——开发者需要看到入口函数的真实代码才能理解每一步做了什么。此外，代码中调用的函数与树中的节点之间没有关联，开发者需要手动在树上找对应函数，打断阅读流。

## What Changes

- **Skill 编排指令**：要求 AI 在步骤节点上放入该步骤的入口函数代码（复用原始节点的 code），使步骤不再是空壳
- **HTML 代码跳转**：代码渲染后，自动识别树中存在的函数名并渲染为可点击链接；点击后跳转到对应节点（展开父节点、高亮、显示代码+描述）
- **CSS**：为可点击的函数名添加链接样式

## Capabilities

### New Capabilities

- `code-cross-navigation`: HTML 模板中的代码跳转功能——渲染代码时自动将树中已有的函数名变为可点击链接，点击后导航到对应树节点

### Modified Capabilities

## Impact

- `skills/gsf-review.md` — Step 3 编排指令更新（步骤节点带入口函数代码）
- `internal/cmd/embed_data/skills/gsf-review.md` — 同步
- `internal/review/templates/review.html` — JS: 节点索引构建 + 代码后处理 + 跳转逻辑；CSS: 链接样式
- 需重新构建 gsf 二进制
