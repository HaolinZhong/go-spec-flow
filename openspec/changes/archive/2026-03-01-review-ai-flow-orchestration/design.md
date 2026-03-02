## Context

`gsf review` 当前直接生成 HTML，树结构按 package/file/function 组织。用户反馈：看到的只是代码堆砌，没有动线引导，不知道从哪看起、每段代码在架构中的角色是什么。

核心架构原则（CLAUDE.md）：**gsf 做精确的脏活，AI 做智能的组织**。当前缺失 AI 组织环节。

## Goals / Non-Goals

**Goals:**
- gsf review 支持 JSON 导出，供 AI 消费
- gsf review 支持从 JSON 渲染 HTML，消费 AI 编排后的数据
- FlowTree/FlowNode 支持 description 字段，承载 AI 生成的解释
- HTML 模板展示 description（动线说明 + 节点注释）
- review skill 实现三步流程：提取 → AI 编排 → 渲染

**Non-Goals:**
- 不做 gsf serve（实时交互留到后续）
- 不做 AI 自动调用（AI 编排由 skill prompt 驱动，不是 gsf 内置 LLM 调用）
- 不改变现有 `gsf review --open` 直接生成 HTML 的行为（保留为快捷方式）

## Decisions

### 1. JSON 管道设计

**选择**: `--json` 标志输出 JSON 到 stdout，`--render <file>` 从文件读取 JSON 渲染 HTML

**理由**:
- stdout 输出适合管道和重定向，AI 可以直接读取
- `--render` 与 `--json` 构成对称的输入/输出接口
- 保持现有 `--open` 行为不变（不加 `--json` 时仍直接输出 HTML）

**替代方案**:
- 固定输出到文件（`--json-output <file>`）→ 多一个临时文件管理，不如 stdout 灵活
- 二合一命令（gsf review-render）→ 破坏命令一致性

### 2. Description 字段位置

**选择**: FlowTree 和 FlowNode 都加 `Description string` 字段

**理由**:
- FlowTree.Description: 整条动线的概述（如"这条动线展示请求从路由到数据库的完整路径"）
- FlowNode.Description: 单个节点的角色说明（如"入口函数，负责参数校验和路由分发"）
- JSON `omitempty` 避免 gsf 原始输出冗余

### 3. AI 编排流程

**选择**: Skill 三步流程，AI 在中间步骤重组 JSON

流程:
1. `gsf review --codebase --json > raw.json`（或 `--commit HEAD --json`）
2. AI 读取 raw.json，理解代码结构，重新组织为动线，添加 description
3. AI 写入 flow.json
4. `gsf review --render flow.json --open`

**理由**:
- AI 有完整的代码上下文（JSON 包含源码），可以做出有意义的组织
- 动线划分是智能决策，不同项目、不同 review 目的，动线不同
- gsf 不需要内置任何 AI 逻辑，保持纯工具定位

### 4. HTML 模板 description 展示

**选择**:
- 根节点（动线）的 description 作为标题下方的段落展示
- 子节点的 description 作为代码面板上方的注释卡片展示

**理由**: description 是引导文字，不应和代码混在一起，也不应喧宾夺主

## Risks / Trade-offs

- **[JSON 体积] → 大项目 JSON 可能很大（>1MB）** → AI 的上下文窗口可能装不下 → skill 中增加提示：大项目建议按 `--entry` 分包 review
- **[AI 编排质量不稳定]** → 不同模型/不同 prompt 产出质量不同 → skill prompt 提供明确的编排指引和示例
- **[现有行为兼容]** → `--json` 和 `--render` 是新增 flag，不影响现有用法 → 零风险
