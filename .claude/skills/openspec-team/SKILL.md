---
name: openspec-team
description: Parallel execution of Feature Spec changes using Coordinator + Workers in isolated worktrees. Use when the user wants to execute multiple Feature Spec changes concurrently instead of sequentially.
license: MIT
compatibility: Requires Claude Code with Agent tool support. Not available for weak model/tool environments.
metadata:
  author: openspec
  version: "1.0"
  generatedBy: "1.2.0"
---

Execute Feature Spec changes in parallel using a Coordinator + Workers model.

Workers run in isolated git worktrees via Claude Code's Agent tool. The Coordinator manages batches, human review, and state.

This is an enhancement over `/opsx:next` (sequential). Both use the same Feature Spec format and can be used interchangeably.

**Requires**: Claude Code with Agent tool support. Not available for weak model/tool environments (use `/opsx:next` instead).

---

**Input**: Specify a feature name (e.g., `/opsx:team coupon-system`). If omitted, auto-selects if only one active feature exists.

**Steps**

1. **Select the Feature Spec**

   a. If a feature name is provided, look for `openspec/features/<name>/feature-spec.md`.

   b. If no name is provided:
      - List directories in `openspec/features/`
      - Filter to features that have at least one change with status other than `completed`
      - If exactly one active feature: auto-select and announce it
      - If multiple active features: use **AskUserQuestion tool** to let the user choose
      - If no features found: inform the user and suggest `/opsx:decompose`

   Announce: "Team mode for feature: **<name>**"

2. **Read and parse the Feature Spec**

   Read `openspec/features/<name>/feature-spec.md`. Parse the Changes section to extract:
   - Each change's letter, name, summary, scope, dependencies, status, key decisions

3. **Reconcile status with actual OpenSpec state**

   Check actual state to auto-correct Feature Spec if it's out of sync:

   ```bash
   openspec list --json
   ```

   Also check `openspec/changes/archive/` for archived changes.

   For each change in the Feature Spec:
   - If an archived directory matching `*-<change-name>` exists → update status to `completed`
   - If an active change directory `openspec/changes/<change-name>/` exists → update status to `proposed`

   If any corrections were made, update the Feature Spec file and announce corrections.

4. **Build dependency graph and identify batches**

   Construct batches using topological ordering:
   - **Batch 1**: All changes with status `pending` and no dependencies (or all dependencies `completed`)
   - **Batch 2**: Changes whose dependencies are all in batch 1 or already `completed`
   - Continue until all changes are assigned to batches

   If only one change is unblocked at a time (fully sequential), suggest using `/opsx:next` instead for simplicity.

5. **Present batch plan**

   Show the full execution plan:

   ```
   ## Team Execution Plan: <feature-name>

   Progress: N/M changes completed

   ### Batch 1 (parallel)
   - A: <change-name> — <summary>
   - B: <change-name> — <summary>

   ### Batch 2 (after batch 1)
   - C: <change-name> — <summary> (depends on: A)

   ### Batch 3 (after batch 2)
   - D: <change-name> — <summary> (depends on: C)

   Total: K batches, up to P changes in parallel
   ```

   Ask: "Ready to start batch 1? I'll launch N Workers in parallel worktrees."

   If the user declines, stop and wait for instructions.

6. **Execute current batch — Propose phase**

   For each change in the current batch, launch a Worker Agent **in parallel** using the Agent tool:

   ```
   Agent tool call:
     subagent_type: "general-purpose"
     isolation: "worktree"
     description: "Propose <change-name>"
     prompt: <WORKER_PROPOSE_PROMPT — see Worker Prompt Templates below>
   ```

   Launch ALL Workers in the current batch simultaneously (multiple Agent calls in a single message).

   Wait for all Workers to complete. Each Worker will return:
   - The artifacts it created (proposal.md, design.md, specs, tasks.md)
   - Or a description of ambiguity/blockers encountered

7. **Handle Worker results**

   For each completed Worker:
   - If the Worker reports **ambiguity or a blocker**: present the issue to the user and ask for guidance. Then re-run that Worker with the clarification.
   - If the Worker completed successfully: note its worktree path and branch for later use.

   Announce: "All N proposals ready for review."

8. **Human review — sequential**

   For each proposal in the batch, **one at a time**:

   a. Read the Worker's proposal artifacts from its worktree:
      - `openspec/changes/<change-name>/proposal.md`
      - `openspec/changes/<change-name>/design.md`
      - `openspec/changes/<change-name>/specs/*/spec.md`
      - `openspec/changes/<change-name>/tasks.md`

   b. Present a summary to the user:
      ```
      ### Reviewing: <change-name> (N of M in this batch)

      **Proposal**: <1-2 sentence summary from proposal.md>
      **Design**: <key decisions summary>
      **Tasks**: K implementation tasks

      [Full artifacts available at: <worktree-path>/openspec/changes/<change-name>/]
      ```

   c. Ask the user:
      - **Approve** — proceed to apply phase
      - **Request changes** — specify what to adjust (small adjustment)
      - **Reject / Rethink** — stop team, suggest revisiting Feature Spec

   d. Handle responses:
      - **Approve**: Mark as approved, continue to next proposal
      - **Request changes**: Re-run that Worker with the user's feedback appended to the prompt. Present the revised proposal again.
      - **Reject**: Stop the team. Announce: "Team paused. Consider re-running `/opsx:decompose` to adjust the feature breakdown."

   All proposals must be approved before proceeding to apply.

9. **Execute current batch — Apply phase**

   For each approved change, launch a Worker Agent **in parallel**:

   ```
   Agent tool call:
     subagent_type: "general-purpose"
     isolation: "worktree"
     description: "Apply <change-name>"
     prompt: <WORKER_APPLY_PROMPT — see Worker Prompt Templates below>
   ```

   **IMPORTANT**: The apply Worker should resume in the SAME worktree where the propose was done (use the `resume` parameter with the previous agent ID if possible, or provide the worktree path context).

   Wait for all Workers to complete.

10. **Merge worktree results**

    For each completed apply Worker:

    a. Get the worktree branch name from the Agent result.

    b. Merge the branch into the current branch:
       ```bash
       git merge <worktree-branch> --no-edit
       ```

    c. **If merge conflict**:
       - Stop immediately
       - Report the conflicting files:
         ```bash
         git diff --name-only --diff-filter=U
         ```
       - Announce: "Merge conflict detected in: <files>. Please resolve the conflicts, then tell me to continue."
       - Wait for user to resolve and confirm before continuing

    d. **If clean merge**: Announce success and continue to next Worker's branch.

11. **Update Feature Spec status**

    After all merges complete for the current batch:
    - Update each completed change's status to `completed` in the Feature Spec
    - Save the Feature Spec file

12. **Next batch loop**

    Re-analyze the Feature Spec:
    - If all changes are `completed`: announce feature completion (see output below)
    - If more unblocked changes exist: go back to Step 5 with the next batch
    - If changes are blocked with no path forward: report the deadlock

---

## Worker Prompt Templates

### WORKER_PROPOSE_PROMPT

Use this template when launching a Worker for the propose phase. Fill in the `<placeholders>`:

```
You are a Worker executing a propose for one change in a Feature Spec.

## Feature Context
- Feature: <feature-name>
- Feature goal: <overview from Feature Spec>
- Your change: <change-letter>: <change-name>
- Summary: <change summary from Feature Spec>
- Scope: <change scope from Feature Spec>
- Key decisions: <key decisions from Feature Spec, if any>

## Completed Dependencies
<For each completed dependency: name, what it accomplished, key outputs>
<If no dependencies: "This change has no dependencies.">

## Your Task: Create an OpenSpec Change with All Artifacts

Run these steps:

1. Create the change:
   ```bash
   openspec new change "<change-name>"
   ```

2. Check artifact build order:
   ```bash
   openspec status --change "<change-name>" --json
   ```

3. For each artifact in dependency order (proposal → design + specs → tasks):

   a. Get instructions:
      ```bash
      openspec instructions <artifact-id> --change "<change-name>" --json
      ```

   b. Read the instruction's `template` for structure, `instruction` for guidance
   c. Read any completed dependency artifacts for context
   d. Create the artifact file at the `outputPath`
   e. Apply `context` and `rules` as constraints — do NOT copy them into the file

4. Check if all applyRequires artifacts are done:
   ```bash
   openspec status --change "<change-name>" --json
   ```

5. When all artifacts are complete, report back with a brief summary of what you created.

## Context Source Discovery
If `openspec/context/index.md` exists, read it to discover available context sources. Use the described tools/files to understand the codebase before creating artifacts.

## Guardrails
- Follow the template structure from openspec instructions
- `context` and `rules` from instructions are constraints for you, NOT content for the file
- If you encounter ambiguity or are unsure about a design decision, STOP and return with a description of the issue. Do NOT guess.
- Keep proposals grounded in the actual codebase structure
```

### WORKER_APPLY_PROMPT

Use this template when launching a Worker for the apply phase. Fill in the `<placeholders>`:

```
You are a Worker implementing an OpenSpec change.

## Change: <change-name>

## Your Task: Implement All Tasks

1. Get apply instructions:
   ```bash
   openspec instructions apply --change "<change-name>" --json
   ```

2. Read the context files listed in the response (proposal, design, specs, tasks).

3. For each pending task in tasks.md:
   - Implement the code changes required
   - Keep changes minimal and focused on the task
   - Mark the task complete: `- [ ]` → `- [x]`
   - Continue to the next task

4. When all tasks are complete, report back with:
   - List of tasks completed
   - List of files created or modified
   - Any issues encountered

## Context Source Discovery
If `openspec/context/index.md` exists, read it. Use the described tools to understand the codebase.

## Guardrails
- If a task is unclear, STOP and return with a description of what's unclear. Do NOT guess.
- If implementation reveals a design issue, STOP and report the issue. Do NOT work around it.
- Keep code changes minimal — only what the task requires
- Do not refactor or "improve" code beyond the task scope
```

---

## Output On Feature Completion

```
## Feature Complete: <feature-name>

All N/N changes completed!

### Execution Summary
- Batches executed: K
- Changes completed: N
- <list each change with its summary>

The feature is fully implemented. Consider:
- Running `/gsf:review` for a final integration review
- Merging to main when ready
- Archiving changes with `/opsx:archive`
```

## Guardrails
- NEVER skip human review of proposals — this is the core safety mechanism
- Workers MUST NOT proceed when encountering ambiguity — they return to Coordinator
- Always reconcile Feature Spec status before planning batches
- If all changes are sequential (no parallelism opportunity), suggest using `/opsx:next` instead
- Feature Spec is the single source of truth — Coordinator updates it after each batch
- On merge conflict, STOP and let the user resolve — do not force-resolve
- This skill requires Claude Code Agent tool support — inform the user if not available
- Compatible with `opsx:next`: same Feature Spec format, users can switch freely
