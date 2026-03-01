## Context

The Investigate module is the AI-facing context engine. It connects M1 (AST analysis) and M2 (Service Registry) to produce structured investigation reports that serve as input for spec generation.

## Goals / Non-Goals

**Goals:**
- Generate investigation reports from specified entry points (routes, functions)
- Include call chain analysis with RPC/MQ markers
- Cross-reference external RPCs with Service Registry data
- Support report merging for multi-entry-point features
- Output YAML format for AI consumption and text format for human review
- Provide `gsf investigate` command

**Non-Goals:**
- NLP-based PRD keyword extraction (manual entry point specification for now)
- Automatic change detection (human specifies what to investigate)

## Decisions

### 1. Report structure

```yaml
investigation:
  target: "Order creation flow"
  entry_points:
    - route: POST /api/v1/orders
      handler: handler.OrderHandler.CreateOrder
  modules:
    - package: sample-app/handler
      role: HTTP handler layer
      functions: [CreateOrder, GetOrder]
    - package: sample-app/service
      role: Business logic
      functions: [CreateOrder, GetOrder]
  call_chains:
    - entry: handler.CreateOrder
      chain:
        - sample-app/service.CreateOrder
        - sample-app/rpc.CreateOrder → [RPC] orderservice.CreateOrder
        - sample-app/dal.Create
        - [MQ] sample-app/service.SendMessage
  external_dependencies:
    - service: OrderService
      methods_used: [CreateOrder]
      context: (from service registry)
  risks:
    - "RPC call to OrderService may timeout under high load"
```

### 2. Investigation workflow

1. User specifies entry points (routes or functions)
2. gsf loads project and traces call chains from each entry
3. Collects all touched packages/modules
4. Identifies external RPC calls and looks up Service Registry
5. Assembles structured report

### 3. CLI interface

```
gsf investigate --route "POST /api/v1/orders" [dir]
gsf investigate --pkg sample-app/handler --func CreateOrder [dir]
gsf investigate --all-routes [dir]
```
