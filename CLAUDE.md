# Go Spec Flow

## Project Overview

面向大型 Go 后端微服务项目的 Spec-Driven AI 开发框架。基于 OpenSpec，扩展上下游能力，覆盖从 PRD 到交付的全链路。

**核心定位**：不是另一个 AI 写代码的工具，而是让 AI 写代码真正能在大型项目中落地的工程体系。

**一句话**：让"需求 → 拆解 → 上下文准备 → AI 编码 → Review → 测试"这条链路标准化、半自动化，让任何水平的程序员都能高效驱动 AI 开发。

## Tech Stack

- **语言**: Go
- **目标用户的框架**: Hertz (HTTP API) + Kitex (RPC)
- **IDL 管理**: 集中在一个仓库，Thrift IDL
- **本框架自身**: Go CLI 工具
- **集成**: 基于 OpenSpec 工作流

## Team Pain Points (Why This Project Exists)

1. **历史代码多，逻辑复杂** — 项目涉及的上下文代码量很大
2. **需求拆解困难** — 从 PRD 到 AI 可执行的 micro task，中间需要大量 codebase 认知，目前纯靠人脑
3. **内部 AI 工具模型弱** — 没有 API key，只能通过 ACP 协议整合；必须把 task 拆得极细才能让弱模型写对
4. **Review 困难** — AI 生成的代码量大，不是自己写的看不懂，修改交互也痛苦
5. **测试负担重** — AI 无法做 E2E 测试，人需要自己想 test case，需要大量上下文
6. **团队标准化缺失** — 不同程序员水平和 AI 熟悉度不同，没有统一流程规范
7. **跨服务 RPC 上下文缺失** — 微服务架构下，AI 看不到外部 RPC 服务的代码和行为，设计和开发质量受限

## Red Lines

- **绝对不能把"并发 Agent 开发"作为核心卖点**。已有团队在做，重复做是恶性竞争。框架可以自然支持，但宣传时不提。

## Architecture - Four P0 Modules

### Module 1: Investigate (代码上下文调研引擎)

**解决痛点**: 拆需求前需要大量人工阅读代码理解上下文

**核心能力**:
- 解析 Go AST，提取项目结构（package 依赖、interface、struct/method 签名）
- 从 Hertz 路由注册解析入口（`r.Group()`, `r.POST()` 等），定位 handler
- 从入口向下追踪调用链：handler → service → dal → Kitex client 调用 → MQ producer
- 遇到 Kitex client 调用时，标记为外部 RPC，从 Service Registry 拉取上下文
- 输出结构化调研报告（YAML 格式），包含：涉及模块、现有逻辑摘要、需要变更的点、外部依赖、风险点

**AI 调研模式**: 不只是被动接受人给的上下文，而是 AI 主动在 codebase 中调研：
- 从 PRD 关键词定位代码入口
- 顺着调用链理解现有逻辑
- 识别外部依赖并补充上下文
- 输出调研报告供人 review 和纠正

### Module 2: Service Registry (跨服务 RPC 上下文注册)

**解决痛点**: 微服务 RPC 调用时 AI 不知道外部服务的接口和行为

**核心能力**:
- 从集中 IDL 仓库解析 Thrift IDL，自动提取 service/method/request/response 定义
- 生成 auto.yaml（自动解析的接口信息）
- 支持 context.yaml（人工补充的业务上下文：幂等性、超时建议、已知坑、错误码等）
- 渐进式积累：每次开发涉及新 RPC 时提示补充

**结构**:
```
service-registry/
├── <service-name>/
│   ├── auto.yaml      ← 自动从 Thrift IDL 解析
│   └── context.yaml   ← 人工补充的上下文（渐进积累）
└── registry-index.yaml
```

### Module 3: OpenSpec Integration + Spec Templates

**解决痛点**: 拆需求靠个人经验，没标准，质量差异大

**核心能力**:
- 定制 Go 后端（Hertz/Kitex）专用 spec 模板体系：
  - L1: Feature Spec（对应 PRD 子项）
  - L2: Design Spec（技术设计：模块/接口/数据变更）
  - L3: Task Spec（AI 可执行的 micro task，自带代码上下文和验收标准）
- L1→L2 需人参与决策，框架提供脚手架
- L2→L3 高度自动化
- 每个 Task Spec 针对弱模型优化：明确、短、有示例、有约束

### Module 4: Flow-Based Review (流式代码审查)

**解决痛点**: 传统 file-based diff 不符合人类 review 习惯

**核心理念**: 人类 review 的真实习惯是：找到入口 → 跟着请求流走 → 在每个节点检查逻辑 → 验证是否符合预期

**核心能力**:
- 解析 git diff，映射到函数/方法级别
- 利用 Investigate 的调用链数据，把 diff 节点挂到调用树上
- 从 Hertz handler 入口开始，按调用顺序组织变更展示
- 每个节点标注：变更类型（modified/new/unchanged/external）、Spec 对应关系、风险提示
- 外部 RPC 节点展示 Service Registry 中的接口信息
- Spec 覆盖度检测：哪些验收标准已实现，哪些遗漏
- 输出：CLI 树状渲染（第一版），未来可扩展 Web UI

## P2 Modules (Future)

- **Code Prompt Adapter**: Task Spec → 结构化 prompt，适配弱模型，支持 ACP 协议
- **Test Case Generator**: 从 Spec 验收标准推导测试用例大纲 + 测试骨架

## Development Flow (Using OpenSpec)

本项目自身使用 OpenSpec 管理开发流程：
1. `/opsx:explore` — 探索和讨论
2. `/opsx:propose` — 提出变更，生成 proposal + design + tasks
3. `/opsx:apply` — 按 tasks 实施
4. `/opsx:archive` — 归档完成的变更

## Schedule

### Phase 1: 框架开发（AI 主力编码，尽快交付）

按模块交付，每个模块完成即可 review：

```
Milestone 1: 项目脚手架 + Go AST 基础能力 (Hertz 路由解析、Kitex client 识别、调用链追踪)
Milestone 2: Service Registry (Thrift IDL parser、auto.yaml 生成)
Milestone 3: Investigate 模块 (调研报告输出)
Milestone 4: OpenSpec 整合 (spec 模板体系 L1/L2/L3)
Milestone 5: Flow-Based Review (git diff → AST 映射 → 调用链叠加 → CLI 渲染)
Milestone 6: 端到端串联跑通
```

人的主要工作：review spec/design、提供真实 Hertz/Kitex 项目代码做验证、决策设计方向。

### Self-Bootstrapping Strategy (自举)

模块开发顺序已针对自举优化。核心模块完成后，立刻用于加速后续开发：
- **OpenSpec**: 从 M1 起全程使用，每个 milestone 走 propose → apply → archive
- **Investigate**: M3 完成后，用于调研自身 Go 代码，辅助开发 M4、M5
- **Flow-Based Review**: M5 完成后，用于 review 后续所有变更
- 自举本身就是最好的 demo："我们用这个框架开发了这个框架"

### Phase 2: 实战验证 + 出成果
```
用真实需求跑完整流程，收集效率数据和 demo 截图
整理文档 + 团队使用规范 + 汇报材料
```

## Presentation Narrative

"我们做的不是又一个 AI 写代码的工具，而是一套让 AI 写代码真正能落地的工程体系。解决的是从需求到交付全链路的效率问题 — 自动分析代码上下文、标准化拆解需求、适配内部 AI 工具、辅助 review 和测试。任何水平的开发者，按照这套流程，都能高效驱动 AI 完成开发。"

## Differentiation

| Dimension | OpenSpec | Concurrent Agents (others) | **This Framework** |
|---|---|---|---|
| Focus | Spec → Code | Parallel coding | **Full pipeline: PRD → Spec → Code → Review → Test** |
| Codebase awareness | Weak (manual) | N/A | **Go AST analysis + auto context extraction** |
| Weak model support | No | No | **Standardized prompt engineering for weak models** |
| Team standardization | Generic templates | N/A | **Go/Hertz/Kitex-specific templates + workflow norms** |
| Review/Test | N/A | N/A | **Flow-based review + spec traceability + test scaffolding** |
| Cross-service RPC | N/A | N/A | **Service Registry from centralized Thrift IDL** |
