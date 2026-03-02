# Feature: <Feature Name>

## PRD Source
[prd.md](./prd.md)

## Overview

<!-- 1-2 sentence summary of the overall feature -->

## Dependency Diagram

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│ A: <name>    │────▶│ B: <name>    │────▶│ C: <name>    │
└──────────────┘     └──────────────┘     └──────────────┘
                                                │
                                                ▼
                                          ┌──────────────┐
                                          │ D: <name>    │
                                          └──────────────┘
```

## Changes

### A: <change-name-kebab-case>
- **Summary**: What this change accomplishes
- **Scope**: [handler, service, dal, model, rpc-client, job, config, ...]
- **Depends on**: []
- **Status**: pending
- **Key decisions**: Any open questions or decisions to make during propose

### B: <change-name-kebab-case>
- **Summary**: What this change accomplishes
- **Scope**: [service, dal]
- **Depends on**: [A: <change-name>]
- **Status**: pending

### C: <change-name-kebab-case>
- **Summary**: What this change accomplishes
- **Scope**: [service, rpc-client]
- **Depends on**: [B: <change-name>]
- **Status**: pending

### D: <change-name-kebab-case>
- **Summary**: What this change accomplishes
- **Scope**: [job, dal]
- **Depends on**: [A: <change-name>]
- **Status**: pending

---

## Status Legend
- **pending**: Not yet started
- **proposed**: `opsx:propose` completed, proposal exists in `openspec/changes/<name>/`
- **completed**: `opsx:apply` + `opsx:archive` done
