# Context Source: Service Registry

## What It Provides

Cross-service RPC interface and behavior context from a centralized Service Registry:
- **Auto-generated interface definitions** (`auto.yaml`): Service methods, request/response types, field definitions — parsed from Thrift IDL
- **Human-curated business context** (`context.yaml`): Idempotency guarantees, timeout recommendations, known pitfalls, error codes, SLA expectations

## When to Use

| Stage | Use Case |
|-------|----------|
| **decompose** | Identify which external services are involved, inform change boundaries |
| **propose** | Understand RPC interfaces for design decisions |
| **apply** | Get precise request/response types, error handling guidance |

## How to Access

### Browse the registry
```
service-registry/
├── <service-name>/
│   ├── auto.yaml      ← Auto-generated from Thrift IDL
│   └── context.yaml   ← Human-curated business context (may not exist yet)
└── registry-index.yaml ← Index of all registered services
```

### Check the index
Read `service-registry/registry-index.yaml` for a list of all registered services.

### Get service details
Read `service-registry/<service-name>/auto.yaml` for interface definitions.
Read `service-registry/<service-name>/context.yaml` for business context (if it exists).

### Update from IDL (if Thrift IDL repo is available)
```bash
gsf registry update --idl-dir <path-to-idl-repo> [--service <name>]
```

## Notes

- `context.yaml` is progressively accumulated — it may not exist for all services
- When a decomposition or proposal involves an RPC call to a service without `context.yaml`, consider noting this as a gap to fill
- `auto.yaml` provides structural information; `context.yaml` provides behavioral information (both are valuable)
