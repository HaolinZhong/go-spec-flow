# Go Spec Flow (gsf)

面向大型 Go 后端微服务项目的 Spec-Driven AI 开发框架。基于 OpenSpec，覆盖从需求到交付的全链路。

**核心定位**：不是另一个 AI 写代码的工具，而是让 AI 写代码真正能在大型项目中落地的工程体系。

## 安装

**最简单的方式**：把这个 repo 地址发给你的 coding agent（Claude Code / Coco），让它帮你安装：

> 帮我安装 https://github.com/HaolinZhong/go-spec-flow ，运行 go install 然后 gsf init

**手动安装**：

```bash
# 需要 Go 1.22+
go install github.com/zhlie/go-spec-flow/cmd/gsf@latest
```

或从源码构建：

```bash
git clone https://github.com/HaolinZhong/go-spec-flow.git
cd go-spec-flow
go build -o gsf ./cmd/gsf
# 把 gsf 移到 PATH 下
```

## 快速开始

### 1. 在你的 Go 项目中初始化

```bash
cd your-go-project

gsf init                # 自动检测 .claude/ 或 .coco/
gsf init --target claude  # 指定安装到 .claude/
gsf init --target coco    # 指定安装到 .coco/
```

这会把 gsf 的 slash commands 安装到 `.claude/commands/gsf/`（或 `.coco/commands/gsf/`），之后在 Claude Code 或 Coco 中可以直接用 `/gsf:review` 等命令。

### 2. Review 代码变更

在 Claude Code 中输入：

```
/gsf:review
```

AI 会引导你选择 review 范围，自动提取代码结构，构建交互式 HTML review。

也可以直接用 CLI：

```bash
# 最近一次 commit
gsf review --commit HEAD --json > raw.json

# vs main 分支
gsf review --base main --json > raw.json

# 整个代码库
gsf review --codebase --json > raw.json

# 指定包
gsf review --codebase --entry "internal/ast" --json > raw.json
```

### 3. 渲染并浏览

```bash
# 带评论功能（推荐）— 启动本地服务器，评论自动保存
gsf review --render flow.json --serve

# 静态 HTML
gsf review --render flow.json --open
```

### 4. 处理评论

review 中添加的评论保存在 `review-comments.json`。用 `/gsf:fix` 让 AI 读取评论并执行修改或回答问题。

## 命令一览

### CLI 命令

| 命令 | 说明 |
|------|------|
| `gsf init` | 安装 slash commands 到项目 |
| `gsf review` | 生成交互式 HTML 代码 review |
| `gsf trace <pkg> <func>` | 从入口函数追踪调用链 |
| `gsf callers <pkg> <func>` | 查找函数的直接调用者 |
| `gsf analyze [dir]` | 分析项目结构（包、类型、函数） |
| `gsf routes [dir]` | 发现 Hertz HTTP 路由注册 |
| `gsf investigate` | 生成代码上下文调研报告 |
| `gsf registry` | 管理跨服务 RPC 上下文 |

### Slash Commands（AI 编排）

安装后在 Claude Code / Coco 中可用：

| 命令 | 说明 |
|------|------|
| `/gsf:review` | AI 驱动的交互式代码 review（核心功能） |
| `/gsf:fix` | 处理 review 评论 — 回答问题或执行代码修改 |
| `/gsf:trace` | AI 引导的调用链追踪 |
| `/gsf:callers` | AI 引导的调用者查找 |
| `/gsf:analyze` | AI 引导的项目结构分析 |
| `/gsf:routes` | AI 引导的路由发现 |
| `/gsf:propose` | 增强版 propose（自动收集 Go 代码上下文） |
| `/gsf:apply` | 增强版 apply（每个 task 附带代码上下文） |

## Review 功能详解

`/gsf:review` 是核心功能。流程：

```
选择 review 范围 → gsf 提取代码结构（JSON）→ AI 构建 master flow → 渲染交互式 HTML
```

生成的 HTML 包含：

- **左侧面板**：流式树导航（主动线 → 步骤 → 函数）
- **右侧面板**：源码 + AI 解读卡片
- **Source/Diff 切换**：在语法高亮源码和彩色 diff 之间切换
- **函数名交叉导航**：代码中的函数名可点击跳转
- **行级评论**：点击行号添加评论，`--serve` 模式下自动保存

### 典型 review 场景

```bash
# review 最近 3 次提交
/gsf:review → 选 "最近 N 次 commits" → 输入 3

# 和 main 分支对比
/gsf:review → 选 "vs 某个分支" → 输入 main

# 探索某个包的架构
/gsf:review → 选 "某个包" → 输入 internal/ast
```

## 架构概览

gsf 遵循"gsf 做精确的脏活，AI 做智能的编排"的原则：

- **gsf CLI**：纯静态分析引擎，无 AI，提供精确的代码结构数据
- **Slash commands**：Markdown 格式的 AI 指令，教 AI 如何调用 gsf 工具并组织结果
- **`gsf init`**：把 slash commands 编译进 binary，一条命令安装到任何项目

```
┌─────────────────────────────────────────────┐
│  AI Agent (Claude Code / Coco / 其他)        │
│  读取 slash command → 调用 gsf CLI → 组织结果  │
├─────────────────────────────────────────────┤
│  gsf CLI                                     │
│  ┌──────────┐ ┌───────┐ ┌─────────┐         │
│  │ review   │ │ trace │ │ callers │  ...     │
│  │ (diff +  │ │       │ │         │         │
│  │  AST)    │ │       │ │         │         │
│  └──────────┘ └───────┘ └─────────┘         │
├─────────────────────────────────────────────┤
│  Go AST Analysis (golang.org/x/tools)        │
└─────────────────────────────────────────────┘
```

## 开发

本项目使用 OpenSpec 工作流管理所有变更：

```bash
/opsx:explore    # 探索和讨论
/opsx:propose    # 提出变更
/opsx:apply      # 实施
/opsx:archive    # 归档
```
