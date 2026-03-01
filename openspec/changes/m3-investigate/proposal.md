## Problem

Before writing specs or tasks, developers need to understand the existing codebase context: what modules exist, how they connect, what external services are called, what patterns are used. Currently this is done entirely by manual code reading, which is slow and error-prone.

## Proposed Solution

Build an Investigate module that automatically generates structured investigation reports by:
1. Starting from PRD keywords or specified code entry points
2. Using the AST engine (M1) to trace call chains and map code structure
3. Cross-referencing with Service Registry (M2) for external RPC context
4. Outputting a YAML investigation report with: involved modules, existing logic summary, change points, external dependencies, risks

## Value

- Reduces hours of manual code reading to minutes of automated analysis
- Provides AI with structured context for higher-quality spec generation
- Creates a reusable artifact that the whole team can reference
