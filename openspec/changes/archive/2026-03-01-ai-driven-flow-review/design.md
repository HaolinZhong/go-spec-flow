## Context

M5 实现了 Flow-Based Review，但 `gsf review` 硬编码 Hertz 路由发现作为入口策略。自举测试暴露了关键问题：gsf 自身是 CLI 项目，没有 Hertz 路由，所有变更都被归为 "Standalone Changes"。

核心认知：**选入口、定动线**是智能决策，不应硬编码在 Go 代码里。静态分析擅长精确提取（diff mapping、调用链追踪、调用者查找），AI 擅长理解意图和组织信息。

现有可复用资产：
- `internal/review/diff.go` — git diff 解析，提取文件级 hunk
- `internal/review/mapper.go` — hunk 到函数级变更映射
- `internal/ast/` — 项目加载、调用链追踪（Tracer）
- `gsf trace --pkg X --func Y` — 已支持指定函数入口的调用链追踪

## Goals / Non-Goals

**Goals:**
- gsf 提供三个精确的工具命令（diff, callers, trace），输出结构化数据供 AI 消费
- AI 通过 `/gsf:review` skill 编排这些工具，根据变更性质动态决定 review 策略
- 适用于任何类型的 Go 项目（Hertz API、CLI 工具、库），不再绑定特定框架
- 输出的代码片段来自 gsf 工具，完整准确，无 AI 幻觉

**Non-Goals:**
- 不实现 AI 模型调用（skill 由 Claude Code / coco 执行，gsf 只是工具）
- 不做跨语言支持（仅 Go）
- 不做 diff 的 before/after 对比（只输出当前版本的完整函数体）
- 不做多层 callers 递归（只返回一层，AI 自行决定是否递归）

## Decisions

### D1: gsf diff — 函数级变更分析命令

**决策**: 新增 `gsf diff` 命令，复用 `diff.go` + `mapper.go`，增加函数体代码提取。

**输出结构** (JSON/YAML):
```yaml
changed_functions:
  - package: "github.com/zhlie/go-spec-flow/internal/review"
    name: "MapDiffToFunctions"
    receiver: ""
    file: "internal/review/mapper.go"
    line_start: 27
    line_end: 74
    is_new: false
    code: |
      func MapDiffToFunctions(diffs []*FileDiff, ...) []*ChangedFunc {
        ...
      }
```

**Flag 设计**:
- `gsf diff [dir]` — 未提交的变更（默认）
- `gsf diff --commit HEAD [dir]` — 指定 commit
- `gsf diff --base main [dir]` — 对比 base branch
- `--format text|json|yaml` — 输出格式

**实现路径**: 新增 `internal/cmd/diff.go`，调用 `review.GetDiff()` → `review.MapDiffToFunctions()` → 读取源文件提取函数体。函数体提取通过 AST 的 `Pos()/End()` 获取行号范围，然后从源文件读取对应行。

**替代方案**: 考虑过复用 `gsf review` 并加 `--format` 输出函数列表，但 review 命令语义是"生成 review 文档"，diff 的语义是"列出变更"，职责不同。

### D2: gsf callers — 直接调用者查找

**决策**: 新增 `gsf callers` 命令，基于 AST 构建反向调用索引。

**实现策略**: 遍历项目所有函数体的 AST，记录每个 `*ast.CallExpr` 的调用目标，构建 `被调用函数 → 调用者列表` 的反向映射。

**输出结构**:
```yaml
target:
  package: "github.com/zhlie/go-spec-flow/internal/review"
  name: "MapDiffToFunctions"
callers:
  - package: "github.com/zhlie/go-spec-flow/internal/cmd"
    name: "reviewCmd.RunE"
    file: "internal/cmd/review.go"
    line: 57
```

**Flag 设计**:
- `gsf callers --pkg <package> --func <name> [dir]`
- `--format text|json|yaml`

**只返回一层的理由**: 多层 callers 的组合爆炸问题由 AI 控制——AI 看到第一层后决定是否继续追踪某个 caller。gsf 保持工具的简单性和确定性。

**实现位置**: `internal/ast/callers.go` — 新增 `FindCallers(project, pkg, funcName)` 函数，返回 `[]CallerInfo`。

### D3: 删除硬编码 review

**决策**: 删除 `gsf review` 命令、`internal/review/flow.go`。

**理由**: `BuildFlowReview()` 的入口发现（Hertz 路由）和动线组织逻辑现在由 AI 在 skill 中完成。保留它会造成两套 review 逻辑共存的混乱。

**保留**: `diff.go`, `mapper.go` 被 `gsf diff` 命令复用，不删除。

### D4: /gsf:review skill 设计

**决策**: 新增 `/gsf:review` skill 文件 (`skills/gsf-review.md`)，教 AI 如何编排工具做 flow review。

**Skill 编排流程**:
1. 运行 `gsf diff --format yaml` 获取变更函数全貌
2. AI 分析变更性质（新 feature / bugfix / 重构 / 小修改）
3. 根据性质选择策略：
   - **新 feature**: 从新增函数向下 `gsf trace`，看实现完整性
   - **bugfix**: `gsf callers` 看影响面，`gsf trace` 看修复路径
   - **重构/改签名**: `gsf callers` 看所有调用方是否适配
   - **小修改**: 直接 review 变更函数代码
4. 按需调用 `gsf callers` / `gsf trace` 补充上下游上下文
5. 组织成人类友好的 review 文档（markdown）

**Skill 不硬编码策略**: skill 文件描述可用工具和示例策略，但明确告诉 AI "根据变更性质自行判断"，不写死 if-else。

**代码引用原则**: review 文档中所有代码片段必须来自 `gsf diff` 输出，不允许 AI 自行生成或记忆代码。

## Risks / Trade-offs

- **[AI 质量依赖]** review 质量取决于 AI 模型能力。弱模型可能选错策略或遗漏关键上下文 → 缓解：skill 提供清晰的策略指引和示例，降低对模型推理能力的要求
- **[callers 精度]** AST 级别的调用者查找无法处理接口方法动态分派和反射调用 → 缓解：这与现有 trace 的限制一致，标记为已知局限；接口方法可通过类型信息辅助匹配
- **[Breaking Change]** 删除 `gsf review` 是破坏性变更 → 缓解：项目还在 Phase 2 内部验证阶段，无外部用户
- **[Skill 维护]** skill markdown 文件需要随工具输出格式变化而更新 → 缓解：skill 引用工具的输出结构，格式变更时同步更新
