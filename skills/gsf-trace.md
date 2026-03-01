---
name: gsf:trace
description: Trace call chains from entry points, marking RPC and MQ calls
---

Run `gsf trace` to build call chain trees from entry functions.

Usage:
```
# Trace from all Hertz routes
gsf trace --route [dir] [--format text|json|yaml]

# Trace from a specific function
gsf trace --pkg <package-path> --func <function-name> [dir]

# Control depth
gsf trace --route --depth 5 [dir]
```

Builds a call tree showing:
- Direct function/method calls across packages
- `[RPC]` markers for Kitex client calls (from kitex_gen packages)
- `[MQ]` markers for message queue producer calls
- Cycle detection and depth limiting
