## Context

go-spec-flow 是 OpenSpec 的 Go 后端增强包。它的核心是一个 Go 静态分析引擎（`gsf` CLI），为 AI 工具在 OpenSpec 流程中提供结构化的代码事实（路由、调用链、RPC 依赖等）。

本 milestone 是项目的第一步：建立项目脚手架和 AST 分析基础能力。后续 Investigate（M3）和 Flow Review（M5）都依赖这些基础能力。

目标用户的项目使用 Hertz (HTTP) + Kitex (RPC) 框架，IDL 集中管理使用 Thrift。

## Goals / Non-Goals

**Goals:**
- 建立可扩展的 Go 项目结构和 CLI 框架
- 实现 Hertz 路由发现
- 实现 Kitex RPC client 调用识别
- 实现跨 package 的函数调用链追踪
- 提供 `gsf init` 命令安装 OpenSpec skills
- 创建 testdata sample 项目验证分析正确性

**Non-Goals:**
- Thrift IDL 解析（M2 scope）
- Service Registry 数据管理（M2 scope）
- 调研报告生成（M3 scope）
- Git diff 分析和 Flow Review（M5 scope）
- Web UI（不在 Phase 1 scope）
- MCP server 集成（后续考虑）

## Decisions

### 1. 项目结构采用标准 Go 项目布局

```
go-spec-flow/
├── cmd/gsf/main.go
├── internal/
│   ├── ast/          ← AST 解析核心
│   │   ├── parser.go     (项目结构解析)
│   │   ├── hertz.go      (Hertz 路由识别)
│   │   ├── kitex.go      (Kitex client 识别)
│   │   └── callgraph.go  (调用链追踪)
│   ├── cmd/          ← CLI 命令定义
│   │   ├── root.go
│   │   ├── routes.go
│   │   ├── trace.go
│   │   ├── analyze.go
│   │   └── init.go
│   └── output/       ← 输出格式化
│       └── formatter.go  (text/json/yaml)
├── skills/           ← 待安装的 OpenSpec skills
├── commands/         ← 待安装的 command 定义
└── testdata/         ← sample Hertz/Kitex 项目
    └── sample-app/
```

**Rationale**: `internal/` 保证包不被外部引用，符合 Go 惯例。`ast/` 作为独立包便于后续模块复用。

### 2. AST 分析使用 `go/ast` + `go/types` + `golang.org/x/tools/go/packages`

**Alternatives considered**:
- 纯 `go/ast`: 缺乏类型信息，无法准确识别 Kitex client 调用
- `go/ssa`: 更精确但更复杂，对第一版过重
- tree-sitter: 不支持 Go 类型系统

**Decision**: 使用 `golang.org/x/tools/go/packages` 加载项目（包含类型信息），配合 `go/ast` 遍历。这样既能做 AST 遍历找到调用点，又能通过类型信息判断调用目标是否为 Kitex client。

### 3. Hertz 路由识别基于 AST pattern matching

Hertz 路由注册代码形如：
```go
h := server.Default()
g := h.Group("/api")
v1 := g.Group("/v1")
v1.POST("/orders", handler.CreateOrder)
```

**Approach**:
1. 扫描所有函数体，找到 `server.Default()` / `server.New()` 调用确定 engine 变量
2. 追踪 `.Group()` 调用链，累积 path prefix
3. 识别 `.GET()` / `.POST()` / `.PUT()` / `.DELETE()` 等方法调用，提取 path 和 handler
4. 通过类型信息验证 receiver 确实是 Hertz 的 `route.IRoutes` 接口

**Limitation**: 如果路由注册使用动态构造（变量拼接 path、循环注册等），pattern matching 可能遗漏。第一版聚焦标准写法，动态场景留待后续增强。

### 4. Kitex client 识别基于类型系统

Kitex 生成的 client 实现特定 interface（如 `xxxservice.Client`）。

**Approach**: 检查方法调用的 receiver 类型，判断其是否来自 Kitex 生成的 client package（通常路径包含 `kitex_gen`）。

### 5. 调用链追踪使用 AST 级别分析（非 SSA）

**Alternatives considered**:
- `golang.org/x/tools/go/callgraph` (基于 SSA): 最精确，但构建 SSA 较慢，对大型项目成本高
- AST 级别: 扫描函数体中的 `CallExpr`，通过类型信息解析被调用函数，速度快但不处理函数值/接口动态分发

**Decision**: 第一版使用 AST 级别。对于 Hertz/Kitex 项目，绝大多数调用是直接函数调用或 receiver 方法调用，AST 级别足够覆盖。接口动态分发场景（如 service interface → impl）通过类型信息解析 concrete type。如果后续精度不够再切换到 SSA。

### 6. testdata sample 项目结构

```
testdata/sample-app/
├── go.mod
├── handler/
│   └── order.go        (Hertz handler)
├── service/
│   └── order.go        (业务逻辑)
├── dal/
│   └── order.go        (数据访问)
├── rpc/
│   └── client.go       (Kitex client 封装)
├── mq/
│   └── producer.go     (MQ producer)
└── router/
    └── router.go       (Hertz 路由注册)
```

这个 sample 覆盖了需要验证的所有场景：Hertz 路由、分层调用、Kitex RPC、MQ producer。

### 7. gsf init 使用 embed 打包 skill 文件

skill 和 command 的 markdown 文件通过 `//go:embed` 编译进 binary，`gsf init` 时直接写出。无需额外下载或外部文件依赖。

## Risks / Trade-offs

**[AST pattern matching 覆盖率] → 先聚焦标准写法，收集真实项目反馈后迭代**
Hertz 路由和 Kitex client 的识别基于 pattern matching，非标准写法可能遗漏。通过 testdata 覆盖主流 pattern，真实项目验证时收集遗漏 case。

**[go/types 加载速度] → 可接受，必要时加缓存**
`go/packages` 加载包含类型信息的大型项目可能较慢（数秒到十数秒）。第一版不做缓存，如果后续体验不可接受再引入增量分析或缓存。

**[调用链精度 vs 速度] → 选择速度优先（AST 级别）**
AST 级别无法处理接口动态分发和函数值。但对目标场景（Hertz/Kitex 分层架构）覆盖率足够。保留后续切换到 SSA 的可能。

**[testdata 维护成本] → sample 项目需要有效的 go.mod 和真实的依赖**
sample 项目需要引入真实的 Hertz 和 Kitex 依赖才能通过类型检查。这增加了依赖管理成本，但对正确性验证是必要的。
