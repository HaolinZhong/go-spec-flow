## ADDED Requirements

### Requirement: gsf review 命令生成 HTML

`gsf review` SHALL 构建 flow tree 并渲染为自包含 HTML 文件。

#### Scenario: Diff 模式生成 HTML
- **WHEN** 用户运行 `gsf review --commit HEAD~3..HEAD`
- **THEN** gsf 生成 `review.html`，用浏览器打开可见左侧 flow tree 和右侧代码面板

#### Scenario: Codebase 模式生成 HTML
- **WHEN** 用户运行 `gsf review --codebase --entry "internal/ast"`
- **THEN** gsf 生成 `review.html`，flow tree 按调用链组织，代码面板展示完整源码

#### Scenario: 自动打开浏览器
- **WHEN** 用户添加 `--open` flag
- **THEN** 生成 HTML 后自动用默认浏览器打开

### Requirement: HTML 左侧 flow tree 导航

HTML 左侧 SHALL 渲染可折叠的 flow tree。每个节点包含函数名、包路径。点击节点在右侧展示对应代码。

#### Scenario: 节点折叠展开
- **WHEN** 用户点击一个有子节点的 tree 节点
- **THEN** 子节点折叠或展开

#### Scenario: 节点选中显示代码
- **WHEN** 用户点击一个 tree 节点
- **THEN** 右侧代码面板滚动到/显示该节点对应的代码

### Requirement: HTML 右侧代码面板

右侧代码面板 SHALL 显示带行号的代码。

Diff 模式下，`+` 行 SHALL 显示绿色背景，`-` 行 SHALL 显示红色背景。

Codebase 模式下，代码 SHALL 有 Go 语法高亮。

#### Scenario: Diff 代码着色
- **WHEN** 代码面板展示 diff 内容
- **THEN** 新增行绿底，删除行红底，上下文行无背景

#### Scenario: Codebase 代码高亮
- **WHEN** 代码面板展示源码
- **THEN** Go 关键字、字符串、注释等有语法高亮

### Requirement: Flow tree 构建逻辑

Diff 模式：
- 输入为 git diff range
- 变更文件作为 tree 的基础节点
- 变更文件中的函数用 gsf trace 向下展开调用链
- 用 gsf callers 向上找一层调用者作为补充

Codebase 模式：
- 输入为入口包或函数
- 从入口用 gsf trace 构建调用链
- 每个节点关联完整的函数源码

#### Scenario: Diff 模式的 tree 包含变更文件
- **WHEN** git diff 显示 3 个文件变更
- **THEN** flow tree 包含这 3 个文件的节点，每个节点包含对应的 diff 代码

#### Scenario: Codebase 模式从入口展开
- **WHEN** 用户指定 `--entry "internal/ast"` 且入口包有 5 个公开函数
- **THEN** flow tree 有 5 个根节点，每个从对应函数的 trace 展开

### Requirement: 自包含 HTML

输出的 HTML 文件 SHALL 是自包含的（单文件），可直接浏览器打开，不依赖本地 server。

代码数据 SHALL 以 JSON 内嵌在 HTML 的 `<script>` 标签中。

#### Scenario: 离线可用
- **WHEN** 用户在无网络环境打开生成的 HTML
- **THEN** flow tree 和代码面板正常工作（语法高亮降级为无样式代码显示）
