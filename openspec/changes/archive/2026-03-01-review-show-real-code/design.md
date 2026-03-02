## Context

gsf 已有核心分析能力：`LoadProject`（AST 加载）、`Tracer.Trace`（前向调用链）、`FindCallers`（反向调用者）、`DiscoverRoutes`（路由发现）。缺的是把这些组合成 flow tree 并渲染成可交互的 HTML。

## Goals / Non-Goals

**Goals:**
- `gsf review` 命令生成自包含 HTML 文件
- 支持 diff 模式（review 代码变更）和 codebase 模式（理解代码架构）
- HTML 交互：左侧 flow tree 导航，右侧代码面板
- 代码带行号、语法高亮，diff 模式标红/标绿
- 单文件 HTML，无需本地 server，浏览器直接打开
- 数据结构为后续 `gsf serve` 动态交互打好基础

**Non-Goals:**
- 本次不做 `gsf serve`（本地 HTTP server）
- 不做 WebSocket 实时交互
- 不做 AI 问答面板
- 不修改 trace/callers/LoadProject 等已有能力

## Decisions

### 1. Flow Tree 数据结构

```go
type FlowTree struct {
    Mode     string      // "diff" or "codebase"
    Title    string      // review 标题
    Roots    []*FlowNode // 多棵树的根
}

type FlowNode struct {
    ID       string      // 唯一标识
    Label    string      // 显示名（如 "LoadProject"）
    Package  string      // 包路径
    File     string      // 文件路径
    LineStart int        // 代码起始行
    LineEnd   int        // 代码结束行
    Code     string      // 源码或 diff 内容
    NodeType string      // "function" / "method" / "rpc" / "mq" / "file"
    IsNew    bool        // diff 模式：是否新增
    Children []*FlowNode
}
```

这个结构直接映射到 HTML 的 tree + code 面板。后续 `gsf serve` 可以用同样的 JSON 数据驱动 WebSocket 动态加载。

### 2. Diff 模式的 flow tree 构建

1. 运行 `git diff <range>` 获取变更文件列表
2. `LoadProject` 加载 AST
3. 对每个变更的 Go 文件，提取变更函数（通过 diff hunk 行号 → AST 映射）
4. 从变更函数出发，用 `Tracer.Trace` 向下展开调用链
5. 用 `FindCallers` 向上找一层调用者
6. 组装成 flow tree，每个节点关联对应的 git diff 片段

注意：之前 `gsf diff` 的 AST 映射有 FuncLit 盲区。这次不做映射 — 直接用 git diff 的文件粒度作为 tree 的叶子节点，trace/callers 作为补充的结构化节点。变更的具体代码就是 git diff 原文，不经过 AST 过滤。

### 3. Codebase 模式的 flow tree 构建

1. 用户指定入口包或函数（`--entry "internal/ast"`）
2. `LoadProject` 加载 AST
3. 从入口出发，用 `Tracer.Trace` 构建调用链
4. 每个节点关联完整的函数源码（通过 AST 定位 `FuncDecl` 的行范围）
5. 如果入口是包级别，列出包内所有公开函数/类型，每个作为一棵子树

### 4. HTML 渲染

使用 Go `html/template` 生成单文件 HTML：

**结构：**
```html
<div id="app">
  <div id="tree"><!-- 左侧 flow tree --></div>
  <div id="code"><!-- 右侧代码面板 --></div>
</div>
```

**交互（纯 JS，无框架）：**
- 点击 tree 节点 → 右侧显示对应代码
- tree 节点可折叠/展开
- 代码面板带行号
- diff 模式：`+` 行绿底，`-` 行红底
- codebase 模式：Go 语法高亮（highlight.js CDN）

**数据传递：**
- FlowTree JSON 内嵌到 HTML 的 `<script>` 标签中
- JS 读取 JSON 动态渲染 tree 和 code

### 5. 命令行接口

```
gsf review --commit HEAD~3..HEAD              # diff 模式，最近 3 个 commit
gsf review --commit HEAD                      # diff 模式，最近 1 个 commit
gsf review --base main                        # diff 模式，vs main
gsf review --codebase --entry "internal/ast"  # codebase 模式
gsf review --codebase                         # codebase 模式，全部包
gsf review ... --output review.html           # 指定输出文件（默认 review.html）
gsf review ... --open                         # 生成后自动用浏览器打开
```

### 6. `/gsf:review` skill 简化

Skill 简化为引导用户运行 `gsf review` 命令，不再指导 AI 生成 review 文档。

## Risks / Trade-offs

- **[Risk] diff 模式的 hunk → 函数映射仍有 FuncLit 盲区** → 不做函数级映射，以文件为基础单元，diff 原文完整展示，trace/callers 作为补充
- **[Risk] 大项目 flow tree 节点过多** → 通过 `--depth` 控制 trace 深度，通过 `--entry` 限制入口范围
- **[Risk] highlight.js CDN 离线不可用** → 可内嵌一个轻量的语法高亮 JS，或 diff 模式不依赖语法高亮（只标红/绿）
- **[Trade-off] 单文件 HTML 可能很大（包含全部代码）** → 这是设计意图，自包含优先于文件大小
