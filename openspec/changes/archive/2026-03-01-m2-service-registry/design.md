## Context

Service Registry is the second core module of go-spec-flow. It provides cross-service RPC context for AI tools by parsing Thrift IDL files and maintaining a structured registry of service interfaces.

The registry has two layers:
- **auto.yaml**: Automatically generated from Thrift IDL parsing (services, methods, request/response types)
- **context.yaml**: Manually maintained behavioral context (idempotency, timeouts, error codes, known issues)

## Goals / Non-Goals

**Goals:**
- Parse Thrift IDL files to extract service/method definitions with full type information
- Generate structured auto.yaml with service interface facts
- Support context.yaml for human annotations
- Provide `gsf registry` commands to update, query, and display the registry
- Create testdata Thrift IDL files for validation

**Non-Goals:**
- Runtime service discovery (this is static analysis only)
- Automatic context.yaml generation (that's human work)
- Integration with Investigate module (M3 scope)

## Decisions

### 1. Thrift IDL Parser using go-thrift or custom parser

**Decision**: Use a lightweight custom Thrift parser focused on extracting service/method/struct definitions. Full Thrift compilation is not needed — we only need interface signatures.

**Rationale**: Real Thrift compilers (like Apache Thrift) are heavy and require installation. A focused parser that handles the subset we need (service, struct, enum, typedef, include) is simpler and has zero external dependencies.

### 2. Registry directory structure

```
service-registry/
├── registry-index.yaml          ← Index of all registered services
├── order-service/
│   ├── auto.yaml               ← Auto-generated from IDL
│   └── context.yaml            ← Human-maintained context
├── user-service/
│   ├── auto.yaml
│   └── context.yaml
└── ...
```

### 3. auto.yaml schema

```yaml
service: OrderService
idl_path: idl/order.thrift
methods:
  - name: CreateOrder
    request:
      - name: user_id
        type: i64
      - name: product_id
        type: i64
    response:
      - name: order_id
        type: string
      - name: status
        type: string
    exceptions:
      - name: OrderError
        type: OrderException
  - name: GetOrder
    ...
types:
  - name: OrderException
    kind: exception
    fields:
      - name: code
        type: i32
      - name: message
        type: string
```

### 4. context.yaml schema

```yaml
service: OrderService
notes: "Core order management service, high QPS"
methods:
  CreateOrder:
    idempotent: false
    timeout_ms: 3000
    notes: "Creates order and triggers payment flow"
    known_issues:
      - "May timeout under high load, retry with backoff"
    error_codes:
      1001: "Invalid product ID"
      1002: "Insufficient stock"
  GetOrder:
    idempotent: true
    timeout_ms: 1000
```

### 5. CLI commands

- `gsf registry update <idl-dir>` — Parse IDL files and update auto.yaml
- `gsf registry show <service-name>` — Display service info (auto + context merged)
- `gsf registry list` — List all registered services

## Risks / Trade-offs

**[Custom Thrift parser coverage] → Focus on common patterns, extend as needed**
A custom parser won't handle every Thrift feature (e.g., complex includes, constants, annotations). Cover the common case first.

**[IDL path resolution] → Relative paths from IDL root**
Include statements in Thrift files use relative paths. The parser needs an IDL root directory to resolve them.
