## ADDED Requirements

### Requirement: AI 编排产出单一主动线 root

AI 编排 Step 3 产出的 JSON SHALL 只有一个 root 节点（主动线），其 children 为各步骤节点，每个步骤的 children 为该步骤涉及的函数/方法节点。

#### Scenario: Codebase 模式编排

- **WHEN** AI 读取 codebase 模式的 raw JSON 并执行 Step 3 编排
- **THEN** 输出 JSON 的 `roots` 数组长度为 1，该 root 的 `label` 描述整体流程，`description` 概述完整生命周期，`children` 包含 2+ 个步骤节点

#### Scenario: Diff 模式编排

- **WHEN** AI 读取 diff 模式的 raw JSON 并执行 Step 3 编排
- **THEN** 输出 JSON 的 `roots` 数组长度为 1，该 root 的 `label` 描述变更主题，`children` 包含按变更逻辑分组的步骤节点

### Requirement: 步骤节点包含子动线函数

主动线下的每个步骤节点 SHALL 作为子动线容器，其 children 包含该步骤涉及的具体函数/方法节点。

#### Scenario: 步骤展开显示函数

- **WHEN** 用户在树形面板展开某个步骤节点
- **THEN** 该步骤下显示属于该子动线的函数节点列表，每个函数保留原始 code 和 description

### Requirement: 主动线步骤节点描述展示

当用户点击无 code 的 group 节点时，HTML 右侧面板 SHALL 展示该节点的 description 而非 "No code available"。

#### Scenario: 点击主动线 root

- **WHEN** 用户点击主动线 root 节点（无 code，有 description）
- **THEN** 右侧面板显示 description 卡片，展示该主动线的概述

#### Scenario: 点击步骤节点

- **WHEN** 用户点击某个步骤节点（无 code，有 description）
- **THEN** 右侧面板显示 description 卡片，展示该步骤在整体流程中的角色

#### Scenario: 节点无 code 也无 description

- **WHEN** 用户点击无 code 且无 description 的节点
- **THEN** 右侧面板显示 "No code available"（保持现有行为）

### Requirement: Skill 指令引导三层结构

`skills/gsf-review.md` Step 3 的编排指令 SHALL 明确要求 AI 构建三层结构（主动线 → 步骤 → 函数），并提供示例 JSON。

#### Scenario: 小包简化

- **WHEN** 被 review 的代码只有少量函数（<5 个）
- **THEN** AI 可以简化为两层（主动线 → 函数），跳过中间步骤层
