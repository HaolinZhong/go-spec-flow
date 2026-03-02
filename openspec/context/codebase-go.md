# Context Source: Go Codebase (gsf)

## What It Provides

Static analysis of Go codebases via the `gsf` CLI tool:
- **Project structure**: Packages, structs, interfaces, functions with signatures
- **HTTP route discovery**: Hertz route registrations (groups, handlers, middleware)
- **Call chain tracing**: Forward tracing from any function through the call graph
- **Caller lookup**: Find direct callers of any function (one level up)

## When to Use

| Stage | Use Case | Precision Needed |
|-------|----------|-----------------|
| **decompose** | Understand project modules and boundaries | Coarse (package-level) |
| **propose** | Identify affected packages, interfaces, data models | Medium (interface/struct-level) |
| **apply** | Understand exact function signatures, call chains | Fine (function-level) |
| **review** | Trace change impact, find upstream/downstream | Fine (function-level) |

## How to Invoke

### Project structure analysis
```bash
gsf analyze [project-dir]
```
Returns: packages, structs (with fields), interfaces, function signatures.

### HTTP route discovery (Hertz projects)
```bash
gsf routes [project-dir]
```
Returns: route groups, HTTP methods, paths, handler function references.

### Forward call chain tracing
```bash
gsf trace --pkg <package> --func <function> [project-dir]
```
Returns: call tree from the specified function, marking RPC calls, MQ producers, and external dependencies.

### Caller lookup (one level up)
```bash
gsf callers --pkg <package> --func <function> [project-dir]
```
Returns: all direct callers of the specified function.

## Prerequisites

- `gsf` binary must be installed and on PATH
- Target project must be a Go project with valid Go modules
