## Context

Review HTML 是一个单文件静态页面，通过 Go template 将 FlowTree JSON 嵌入 `<script>` 中。当前 `gsf review --render flow.json --open` 生成静态 `review.html` 并用浏览器打开。HTML 中的代码展示已具备行号（`.line-num` / `.diff-line-num` span），可作为评论的锚点。

FlowNode 已有 `File`、`LineStart`、`LineEnd` 字段，可映射评论到原始文件的绝对行号。

浏览器安全模型不允许静态 HTML 直接写磁盘，因此需要本地 HTTP 服务器作为桥梁。

## Goals / Non-Goals

**Goals:**
- 用户可在 review HTML 中点击代码行号添加自由文本评论
- 评论通过本地 HTTP 服务器自动保存到磁盘（无需手动下载）
- 评论文件格式结构化，包含足够信息让 AI 精确定位修改位置
- 新增 `/gsf:fix` skill 读取评论文件并驱动 AI 执行修改

**Non-Goals:**
- 不做多人协作评论（只有单用户本地使用）
- 不做评论的版本管理或历史记录
- 不改变现有 `--open` 模式（`--serve` 是新增选项，两者共存）
- 不做评论的编辑/删除 UI 的复杂化（基础 CRUD 即可）

## Decisions

### Decision 1: 本地 HTTP 服务器架构

**选择**：`gsf review --serve` 启动 `net/http` 服务器，随机端口，提供两个端点。

```
GET  /           → 渲染 review HTML（内存中，不写文件）
POST /comments   → 接收评论 JSON，写入 review-comments.json
```

**理由**：Go 标准库 `net/http` 零依赖；随机端口避免冲突；内存渲染避免产生临时文件。`--serve` 隐含 `--open`（自动打开浏览器）。

**替代方案**：File System Access API（浏览器兼容性差）；WebSocket（过度设计）。

### Decision 2: 评论文件格式

**选择**：JSON 文件，结构如下：

```json
{
  "reviewTitle": "Diff Review: ...",
  "mode": "diff",
  "exportedAt": "2026-03-01T...",
  "comments": [
    {
      "file": "internal/review/builder.go",
      "line": 128,
      "codeContext": "seen[pkgPath+\".\"+funcName] = true",
      "comment": "这里的 key 应该考虑 receiver type"
    }
  ]
}
```

**关键字段**：
- `file`: 原始文件路径（相对项目根目录）
- `line`: 原始文件的绝对行号（非 diff 行号）
- `codeContext`: 该行的代码内容，帮助 AI 在文件变化时仍能定位
- `comment`: 用户的自由文本评论

**理由**：JSON 被 AI 工具原生理解；`codeContext` 提供容错定位；绝对行号最直接。

### Decision 3: 评论 UI 交互

**选择**：点击行号区域触发评论。行号 span（`.line-num` / `.diff-line-num`）添加点击事件，弹出内联评论输入框。已有评论的行号高亮显示。

**评论存储**：JavaScript 中维护 `comments` Map（key: `file:line`），每次增删改通过 POST `/comments` 同步到服务器。

**理由**：行号点击是最直观的交互模式（GitHub PR review 同款）；内联输入框避免弹窗打断阅读流；实时同步确保不丢失。

### Decision 4: 行号映射策略

**选择**：评论始终关联原始文件的绝对行号。

- Source 模式：行号直接就是文件行号（来自 `lineStart + i`），无需转换
- Diff 模式：从 hunk header 解析的 `newLine` 已经是文件绝对行号，也无需转换
- `+` 行（新增行）：使用 new side 行号
- `-` 行（删除行）：不可评论（该行已不存在于当前文件中）

**理由**：AI 修改代码时需要当前文件的行号；删除行评论无意义（行已不存在）。

### Decision 5: `/gsf:fix` skill 设计

**选择**：独立 skill，不整合到 `/opsx:apply`。

流程：
1. 读取 `review-comments.json`
2. 按文件分组评论
3. 对每个文件，读取源码，定位评论行
4. 逐条理解评论意图，执行代码修改
5. 修改完成后标记评论为已处理

**理由**：评论驱动的修改和 OpenSpec task 语义不同；独立 skill 更简单、用户体验更直观。

### Decision 6: 评论文件保存位置

**选择**：保存到项目根目录 `review-comments.json`。

**理由**：服务器运行时不一定知道当前 OpenSpec change 是哪个；项目根目录是最确定的位置；`/gsf:fix` skill 可以读取后决定是否融入 OpenSpec 工作流。

## Risks / Trade-offs

- **服务器端口冲突** → 使用 `:0` 随机端口，启动后打印实际端口
- **评论与代码不同步** → `codeContext` 字段提供行内容作为 fallback 定位；AI 可根据上下文模糊匹配
- **浏览器关闭未保存** → 每次评论操作实时 POST 同步，不依赖关闭前保存
- **Diff 模式下删除行不可评论** → 明确设计决策，删除行评论无可操作意义
