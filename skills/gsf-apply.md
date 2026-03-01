---
name: gsf:apply
description: Enhanced apply workflow with Go code context per task
---

Before implementing each task in a Go (Hertz/Kitex) project:

1. **Trace the entry point**: If the task involves a specific handler or route:
   ```
   gsf trace --pkg <package> --func <function> .
   ```
   This shows the current call chain from that entry point.

2. **Check routes**: If adding or modifying HTTP endpoints:
   ```
   gsf routes .
   ```
   Verify the current route table.

3. **Review external dependencies**: If the task involves RPC calls:
   ```
   gsf registry show <service-name>
   ```
   Check the service interface, expected behavior, and known issues.

4. **After implementation**: Run `gsf trace --route .` to verify the call chain matches expectations.

Use gsf output to validate:
- New code follows the existing handler → service → dal pattern
- RPC calls use the correct method signatures
- MQ messages are published at the right points
