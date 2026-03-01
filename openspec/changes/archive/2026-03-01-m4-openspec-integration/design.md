## Context

This milestone integrates gsf's analysis capabilities into the OpenSpec workflow through enhanced skill files and a unified context command.

## Goals / Non-Goals

**Goals:**
- Create `gsf context` command that outputs unified AI-consumable context
- Create enhanced OpenSpec skill files for gsf-aware propose/apply
- Update `gsf init` to install the enhanced skills
- Support both `.claude/` and `.coco/` target directories

**Non-Goals:**
- Custom OpenSpec schema (using standard spec-driven schema)
- Parent-change L1 splitting (future enhancement)

## Decisions

### 1. gsf context command

Combines investigation report + project structure + relevant registry data into a single document:

```
gsf context --all-routes [dir]
gsf context --route "POST /orders" [dir]
```

Output is a structured YAML document designed for AI consumption.

### 2. Enhanced skill files

Skills instruct the AI to run gsf commands and incorporate results:

- `gsf-propose.md`: Before generating proposal/design, run `gsf investigate` and `gsf context` to gather code facts
- `gsf-apply.md`: Before implementing each task, run `gsf trace` for relevant entry points to understand current code

### 3. gsf init updates

Install enhanced skills alongside the basic ones. Detect target directory automatically.
