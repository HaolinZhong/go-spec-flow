---
name: "OPSX: Next"
description: "Find and propose the next unblocked change from a Feature Spec"
category: Workflow
tags: [workflow, feature-spec, sequential, experimental]
---

Find the next change to work on from a Feature Spec and kick off its proposal.

This skill reads a Feature Spec (created by `/opsx:decompose`), identifies which changes are ready to start, and initiates `opsx:propose` for the selected one.

---

**Input**: Optionally specify a feature name (e.g., `/opsx:next coupon-system`). If omitted, auto-selects if only one active feature exists.

**Steps**

1. **Select the Feature Spec**

   a. If a feature name is provided, look for `openspec/features/<name>/feature-spec.md`.

   b. If no name is provided:
      - List directories in `openspec/features/`
      - Filter to features that have at least one change with status other than `completed`
      - If exactly one active feature: auto-select and announce it
      - If multiple active features: use **AskUserQuestion tool** to let the user choose
      - If no features found: inform the user and suggest `/opsx:decompose`

   Announce: "Working on feature: **<name>**"

2. **Read and parse the Feature Spec**

   Read `openspec/features/<name>/feature-spec.md`. Parse the Changes section to extract:
   - Each change's name, summary, scope, dependencies, status

3. **Reconcile status with actual OpenSpec state**

   Check actual state to auto-correct Feature Spec if it's out of sync:

   ```bash
   openspec list --json
   ```

   Also check `openspec/changes/archive/` for archived changes.

   For each change in the Feature Spec:
   - If an archived directory matching `*-<change-name>` exists in `openspec/changes/archive/` but Feature Spec status is not `completed` → update to `completed`
   - If an active change directory `openspec/changes/<change-name>/` exists but Feature Spec status is `pending` → update to `proposed`

   If any corrections were made, update the Feature Spec file and announce: "Auto-corrected status for: <change-names>"

4. **Identify unblocked changes**

   A change is **unblocked** when:
   - Its status is `pending`
   - ALL changes in its `depends_on` list have status `completed`

   Categorize all changes:
   - **Completed**: status = `completed`
   - **In progress**: status = `proposed`
   - **Unblocked**: status = `pending`, dependencies satisfied
   - **Blocked**: status = `pending`, dependencies not satisfied

5. **Present the situation**

   Show a progress summary:

   ```
   ## Feature: <name>

   Progress: N/M changes completed

   ✅ Completed:
     - A: <summary>

   🔄 In progress (proposed):
     - B: <summary>

   🟢 Ready to start:
     - C: <summary>
     - D: <summary>

   🔒 Blocked:
     - E: <summary> (waiting for: C)
   ```

   **Handle each situation:**

   a. **All completed**: Congratulate! Suggest archiving the feature or cleaning up.

   b. **Some in progress (proposed but not completed)**: Remind the user to complete in-progress changes first. Suggest `/opsx:apply <change-name>` for each.

   c. **One unblocked change**: Present it and ask: "Ready to propose **<change-name>**?"

   d. **Multiple unblocked changes**: Present all and use **AskUserQuestion tool** to let the user pick which to work on next.

   e. **No unblocked, some blocked**: Explain the deadlock — which changes are blocking others. This shouldn't normally happen if in-progress changes are completed first.

6. **Kick off propose**

   When the user confirms a change to work on:

   a. **Prepare context** from the Feature Spec:
      - Overall feature goal (from Overview section)
      - This change's summary and scope
      - Completed dependency changes and what they accomplished
      - Any key decisions noted in the Feature Spec for this change

   b. **Update the Feature Spec**: Set the selected change's status to `proposed`

   c. **Run propose**:
      ```bash
      openspec new change "<change-name>"
      ```

      Then follow the standard propose workflow — create artifacts (proposal, design, specs, tasks) using `openspec instructions` and `openspec status`. When writing the **proposal.md**, incorporate the Feature Spec context:

      - Reference the parent feature
      - Include the change summary as the starting point for the "Why" section
      - Note completed dependencies and what they provide
      - Address any key decisions from the Feature Spec

   d. **After propose completes**: Announce completion and remind:
      "Proposal ready! Run `/opsx:apply <change-name>` to implement, then `/opsx:next` for the next change."

**Output On All Completed**

```
## Feature Complete: <name>

All N/N changes completed! 🎉

### Completed Changes
- A: <summary> ✅
- B: <summary> ✅
- C: <summary> ✅

The feature is fully implemented. Consider:
- Running a final integration review
- Cleaning up the feature spec
```

**Guardrails**
- NEVER skip the human confirmation step before proposing
- Always reconcile Feature Spec status with actual OpenSpec state before presenting options
- When proposing, provide Feature Spec context so the proposal is coherent with the overall feature
- If a change with the same name already exists as an active OpenSpec change, ask the user if they want to continue it
- Update Feature Spec status to `proposed` only after the OpenSpec change is created
- Do not modify the Feature Spec's change definitions (summary, scope, dependencies) — only update status fields
