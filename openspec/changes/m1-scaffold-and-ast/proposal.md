## Why

go-spec-flow 项目需要一个可运行的基础框架和核心的 Go AST 分析引擎。AST 分析能力是所有后续模块（Investigate、Flow Review）的共享地基。没有它，后续模块无法开发。作为第一个 milestone，需要建立项目脚手架、CLI 入口和基础的代码分析能力。

## What Changes

- 建立 Go 项目结构：`cmd/gsf/main.go` 入口，`internal/` 分包
- 基于 Cobra 搭建 CLI 框架，支持子命令扩展
- 实现 Go AST 解析基础能力：
  - 解析项目结构（package、struct、interface、function 签名）
  - 从 Hertz 路由注册代码中识别 HTTP 路由入口（`r.Group()`, `r.POST()` 等），定位 handler 函数
  - 识别 Kitex RPC client 调用（标记为外部依赖）
  - 从指定函数入口向下追踪调用链（handler → service → dal → RPC client → MQ producer）
- 提供 CLI 命令暴露这些能力：`gsf routes`、`gsf trace`、`gsf analyze`
- 创建 sample Hertz/Kitex 项目用于开发期测试验证

## Capabilities

### New Capabilities
- `ast-analysis`: Go AST 解析引擎 — 项目结构提取、Hertz 路由识别、Kitex client 识别、调用链追踪
- `cli-framework`: gsf CLI 框架 — Cobra 基础结构、子命令体系、输出格式（YAML/JSON）

### Modified Capabilities

(none)

## Impact

- 新建整个 Go 项目结构，引入 cobra、go/ast、go/types、golang.org/x/tools 等依赖
- 需要一个 sample Hertz/Kitex 项目（testdata/ 目录下）用于验证 AST 解析的正确性
- 后续所有模块（Investigate M3、Flow Review M5）都将依赖此 milestone 的 AST 分析能力
