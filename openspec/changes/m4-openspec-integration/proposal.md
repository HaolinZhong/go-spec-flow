## Problem

OpenSpec provides generic spec-driven workflow but lacks Go/Hertz/Kitex-specific awareness. Spec generation relies entirely on manual context input. Without structured codebase facts, generated specs miss important constraints and produce lower-quality task breakdowns.

## Proposed Solution

Create OpenSpec skill files that teach AI tools to use gsf commands during the propose/apply workflow:
1. Enhanced propose skill: runs `gsf investigate` before generating proposals, injects code context into design/task generation
2. Enhanced apply skill: provides code context per task, validates changes against spec
3. Context command: `gsf context` that combines investigate report + registry data into a single AI-consumable context document

## Value

- Bridges gsf analysis capabilities with OpenSpec workflow
- Skills work with both Claude Code (.claude/) and internal tools (.coco/)
- Context document gives AI all necessary facts for high-quality spec generation
