## 1. Investigation Report Model

- [x] 1.1 Define report data model (`internal/investigate/report.go`): InvestigationReport, ModuleInfo, CallChainInfo, ExternalDependency, with YAML/JSON tags and text formatting

## 2. Report Generator

- [x] 2.1 Implement report generator (`internal/investigate/generator.go`): from entry points, trace call chains, collect modules, identify external dependencies, cross-reference Service Registry
- [x] 2.2 Support route-based entry point resolution (match route pattern to discovered routes)

## 3. CLI Command

- [x] 3.1 Implement `gsf investigate` command (`internal/cmd/investigate.go`) with --route, --pkg/--func, and --all-routes flags, --registry-dir for Service Registry lookup
