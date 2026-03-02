---
name: "OPSX: Decompose"
description: "Decompose a PRD into a Feature Spec with multiple changes and dependency tracking"
category: Workflow
tags: [workflow, decompose, feature-spec, experimental]
---

Decompose a large requirement (PRD) into multiple focused OpenSpec changes.

I'll generate a Feature Spec (L1 artifact) that:
- Breaks the requirement into manageable changes
- Defines dependencies between changes
- Tracks execution status

When ready to execute, use `/opsx:next` to work through changes sequentially.

---

**Input**: The argument after `/opsx:decompose` can be:
- A file path to a PRD document (e.g., `/opsx:decompose docs/prd.md`)
- A text description of the requirement (e.g., `/opsx:decompose add coupon system with creation, redemption, and expiry`)
- Nothing (will prompt for input)

**Steps**

1. **Get the requirement**

   - If a file path is provided: read the file as the PRD
   - If text is provided: treat it as the requirement description
   - If nothing is provided: use **AskUserQuestion tool** to ask:
     > "What requirement do you want to decompose? You can paste a description or provide a path to a PRD document."

   Derive a kebab-case feature name from the requirement (e.g., "商品优惠券系统" → `coupon-system`).

2. **Gather context**

   Before analyzing the requirement, gather available context:

   a. **Check for context sources**: If `openspec/context/index.md` exists, read it and any relevant context source files. Use the described tools/files to understand the codebase structure. This helps produce a better decomposition grounded in the actual code.

   b. **If no context sources**: Explore the codebase directly — read key files, search for relevant code, understand the project structure as needed.

   c. **Check existing OpenSpec state**:
      ```bash
      openspec list --json
      ```
      Understand if there are active changes that might relate to or conflict with this feature.

3. **Generate the decomposition**

   Analyze the requirement and codebase context to produce a Feature Spec. Apply these principles:

   **Decomposition principles:**
   - Each change should be **"one propose can explain it, AI can execute it in one go"** — this is a cognitive complexity limit, not a line count limit
   - Prefer changes that touch **different files/packages** to minimize conflicts if later executed in parallel
   - Put **foundational work first** (data models, interfaces) as dependencies of higher-level changes (business logic, integrations)
   - External integrations (RPC, MQ, third-party APIs) should be **separate changes** due to their distinct context needs
   - If uncertain about granularity, **err on the side of smaller changes** — they can always be merged later

   **Dependency principles:**
   - A change depends on another only if it **cannot be proposed or implemented** without the other's output
   - Maximize parallelism — avoid unnecessary sequential dependencies
   - Common pattern: data model → business logic → API layer → integrations

4. **Present the Feature Spec to the user**

   Show the generated Feature Spec in full. Include:
   - Feature name and summary
   - ASCII dependency diagram
   - Each change with name, summary, scope, dependencies
   - Total count: "N changes, M can start immediately (no dependencies)"

   Ask: "How does this decomposition look? You can ask me to merge, split, add, remove, or reorder changes."

5. **Handle adjustments**

   If the user requests changes:
   - **Merge**: Combine two changes into one, update dependencies
   - **Split**: Break one change into multiple, add dependencies
   - **Add**: Insert a new change at the right position in the dependency graph
   - **Remove**: Delete a change, update any dependencies that pointed to it
   - **Reorder**: Adjust dependency relationships

   Present the updated Feature Spec after each adjustment.
   Repeat until the user approves.

6. **Save the Feature Spec**

   Once approved:

   ```bash
   mkdir -p openspec/features/<feature-name>
   ```

   - Write the Feature Spec to `openspec/features/<feature-name>/feature-spec.md`
   - If the input was a file, copy it to `openspec/features/<feature-name>/prd.md`
   - If the input was text, save it as `openspec/features/<feature-name>/prd.md`

**Output**

After saving, summarize:
```
## Feature Decomposed: <feature-name>

**Location:** openspec/features/<feature-name>/
**Changes:** N total
**Ready to start:** M changes (no dependencies)
**Dependency depth:** D levels

Run `/opsx:next` to start working through changes sequentially.
Run `/opsx:next <feature-name>` if you have multiple features.
```

**Feature Spec Format**

Use this structure for the generated Feature Spec:

```markdown
# Feature: <Feature Name>

## PRD Source
[prd.md](./prd.md)

## Overview
<!-- 1-2 sentence summary -->

## Dependency Diagram
<!-- ASCII diagram showing change dependencies -->

## Changes

### A: <change-name>
- **Summary**: What this change accomplishes
- **Scope**: [package/module areas]
- **Depends on**: []
- **Status**: pending
- **Key decisions**: Any open questions for the propose phase

### B: <change-name>
- **Summary**: What this change accomplishes
- **Scope**: [package/module areas]
- **Depends on**: [A: <change-name>]
- **Status**: pending
```

**Guardrails**
- NEVER implement code — this skill only produces the Feature Spec artifact
- Always present the decomposition for human review before saving
- If the requirement is small enough for a single change, say so and suggest using `/opsx:propose` directly instead
- Each change name must be unique and kebab-case
- Dependencies must not form cycles
- All changes start with status `pending`
- Check `openspec/context/` for context sources — use them if available, gracefully skip if not
