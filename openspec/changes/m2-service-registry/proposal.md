## Problem

In microservice architectures, AI tools cannot see external RPC service implementations. When developing features that involve cross-service calls, the AI lacks context about the called service's interface contracts, behavior, and constraints. This leads to incorrect assumptions, wrong error handling, and missed edge cases.

## Proposed Solution

Build a Service Registry that:
1. Parses Thrift IDL files from the centralized IDL repository to extract service/method/request/response definitions
2. Generates `auto.yaml` files with structured interface information
3. Supports `context.yaml` for human-provided behavioral context (idempotency, timeouts, known issues, error codes)
4. Provides CLI commands to manage and query the registry

## Value

- AI gets accurate, structured RPC interface context during spec generation
- Progressive accumulation: each new development fills in more service context
- Separates auto-generated facts from human-curated knowledge
