---
name: gsf:propose
description: Enhanced propose workflow with Go codebase analysis
---

Before generating a proposal or design for a Go (Hertz/Kitex) project, gather code context using gsf:

1. **Analyze project structure**: Run `gsf analyze .` to understand packages, types, and functions.

2. **Investigate relevant code paths**:
   - If the change involves specific routes: `gsf investigate --route "<METHOD> <path>" .`
   - For broader changes: `gsf investigate --all-routes .`

3. **Check service registry**: Run `gsf registry list` to see available external service context.
   For specific services: `gsf registry show <service-name>`

4. **Generate unified context**: `gsf context --all-routes .` for a complete context document.

Use the gathered context to:
- Identify which modules/packages need changes
- Understand existing call chains that will be affected
- Note external RPC dependencies and their constraints
- Identify risks from cross-service interactions

Include this analysis in the proposal and design artifacts.
