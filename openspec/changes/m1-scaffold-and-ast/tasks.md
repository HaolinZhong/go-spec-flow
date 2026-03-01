## 1. Project Scaffold

- [x]1.1 Initialize Go module (`go mod init`), create `cmd/gsf/main.go` with Cobra root command, and `internal/cmd/root.go` with version/help
- [x]1.2 Create output formatter (`internal/output/formatter.go`) supporting text, JSON, and YAML output formats via `--format` flag

## 2. Testdata Sample Project

- [x]2.1 Create `testdata/sample-app/` with `go.mod`, Hertz router registration (`router/router.go`), handler, service, dal layers, Kitex client wrapper (`rpc/client.go`), and MQ producer (`mq/producer.go`)

## 3. Go AST Project Structure Analysis

- [x]3.1 Implement project loader (`internal/ast/parser.go`) using `golang.org/x/tools/go/packages` to load packages with type information
- [x]3.2 Extract project structure (packages, structs, interfaces, function signatures) from loaded packages
- [x]3.3 Wire up `gsf analyze` command (`internal/cmd/analyze.go`) that outputs project structure

## 4. Hertz Route Discovery

- [x]4.1 Implement Hertz route parser (`internal/ast/hertz.go`): detect `server.Default()`/`server.New()`, trace `.Group()` chains for path prefix accumulation, extract `.GET()`/`.POST()`/etc. registrations with handler mapping
- [x]4.2 Wire up `gsf routes` command (`internal/cmd/routes.go`) that outputs route table

## 5. Kitex Client Identification

- [x]5.1 Implement Kitex client detector (`internal/ast/kitex.go`): identify method calls on receivers whose type originates from `kitex_gen` packages, extract service name and method name

## 6. Call Chain Tracing

- [x]6.1 Implement call chain tracer (`internal/ast/callgraph.go`): from a given entry function, traverse `CallExpr` nodes using type info to resolve targets, build a call tree across packages, support `--depth` limit, detect cycles
- [x]6.2 Integrate Kitex client detection and MQ producer detection into call chain nodes (mark as `[external-rpc]` and `[mq-producer]`)
- [x]6.3 Wire up `gsf trace` command (`internal/cmd/trace.go`) that outputs call chain tree

## 7. gsf init Command

- [x]7.1 Embed skill and command markdown files using `//go:embed`, implement `gsf init` command (`internal/cmd/init.go`) that detects `.claude/` or `.coco/` and copies files accordingly
