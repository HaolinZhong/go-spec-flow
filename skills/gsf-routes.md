---
name: gsf:routes
description: Discover Hertz HTTP route registrations in a Go project
---

Run `gsf routes` to discover all Hertz route registrations.

Usage:
```
gsf routes [dir] [--format text|json|yaml]
```

Detects `server.Default()`/`server.New()`, traces `.Group()` chains for path prefix accumulation, and extracts HTTP method registrations (GET, POST, PUT, DELETE, etc.) with their handler mappings.
