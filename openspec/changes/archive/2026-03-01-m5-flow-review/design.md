## Context

Flow-Based Review is the final major module. It uses git diff parsing + AST call chain analysis to present code changes organized by request flow rather than by file.

## Goals / Non-Goals

**Goals:**
- Parse git diff to extract changed functions
- Map diff hunks to function/method boundaries using AST
- Overlay change markers onto call chain tree
- CLI tree rendering with change annotations
- Show relevant Service Registry context

**Non-Goals:**
- Web UI (future)
- Inline code diff display (just markers and summary)
- Spec coverage tracking (future enhancement)

## Decisions

### 1. Git diff parsing

Parse unified diff format to extract:
- Changed files
- Hunk line ranges
- Map line ranges to function declarations using AST

### 2. Change classification

For each function in the call chain:
- `[modified]` — function body changed in the diff
- `[new]` — function is newly added
- `[unchanged]` — function exists but not changed
- `[external-rpc]` — external RPC call (from M1)
- `[mq-producer]` — MQ call (from M1)

### 3. CLI output

```
POST /api/v1/orders → handler.CreateOrder
├── [modified] handler.CreateOrder
│   └── [modified] service.CreateOrder
│       ├── [new] service.ValidateStock
│       ├── [unchanged] rpc.CreateOrder
│       │   └── [RPC] orderservice.CreateOrder
│       ├── [modified] dal.Create
│       └── [unchanged] [MQ] SendMessage
```

### 4. Standalone changes

Functions changed but not in any call chain (e.g., utility functions, model changes) are shown separately as "Standalone Changes".

### 5. CLI commands

```
gsf review [dir]                    # review uncommitted changes
gsf review --commit HEAD [dir]      # review specific commit
gsf review --base main [dir]        # review changes vs base branch
```
