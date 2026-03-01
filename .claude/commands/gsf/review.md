---
name: gsf:review
description: AI-driven flow-based code review using git diff + gsf tools
---

You are performing a flow-based code review. Use `git diff` for complete change information, and `gsf` tools for Go call-chain context. Organize the review by logical flow, not file order.

## Step 1: Determine Review Scope

Ask the user what to review if not clear. Common patterns:

```bash
# Last commit
git diff HEAD~1 HEAD

# Last N commits
git diff HEAD~N HEAD

# Branch vs main
git diff main...HEAD

# Staged changes
git diff --staged

# Unstaged changes
git diff
```

First run `git diff --stat <range>` to get an overview of changed files, then read the full diff.

For large diffs, read per-file: `git diff <range> -- <file>` to manage context window.

## Step 2: Read the Full Diff

Read the complete `git diff` output. This is your primary source of truth — it shows every change with `+/-` lines, no blind spots.

Group the changes mentally:
- Which files/functions were modified?
- Which are new files?
- Which are test changes?
- Which are non-Go changes (configs, docs, etc.)?

## Step 3: Determine Change Type and Review Strategy

Based on the diff, determine:
- **Change type**: new feature, bugfix, refactor, small fix
- **Scope**: single function, module-level, cross-module
- **Review strategy**: which "flow" to follow

Choose a review strategy:
- **Request flow** (new feature): entry point → downstream implementation → tests
- **Impact flow** (bugfix/refactor): change point → callers (who's affected) → tests
- **Data flow**: input → transformations → output
- **Simple review** (small change): direct review without flow

## Step 4: Supplement with gsf Tools (as needed)

Use gsf tools to understand code structure that isn't visible in the diff:

### gsf callers — Who calls this function?
```bash
gsf callers --pkg <package-path> --func <function-name> [dir]
```
Use when: a function's signature or behavior changed, and you need to assess impact on callers.

### gsf trace — What does this function call?
```bash
gsf trace --pkg <package-path> --func <function-name> [dir]
gsf trace --pkg <package-path> --func <function-name> --depth 5 [dir]
```
Use when: reviewing a new entry point and you want to see its full call chain.

**Only use these tools when the diff alone doesn't give enough context.** Don't call them for every changed function.

## Step 5: Write the Review Document

Structure your review as a **flow**, not a file list:

```markdown
# Flow Review: [brief description]

## Change Overview
- Type: [new feature / bugfix / refactor]
- Scope: [summary of what changed]
- Files: [N files changed, from git diff --stat]

## Review Flow

### 1. [Logical starting point]

<diff snippet showing the actual +/- changes>

**Observations:** [what you notice, potential issues]

### 2. [Next node in the flow]

<diff snippet>

**Observations:** [...]

...

## Impact Assessment
- Callers affected: [from gsf callers, if checked]
- Downstream: [from gsf trace, if checked]

## Summary
| Finding | Severity | Verdict |
|---------|----------|---------|
| ...     | ...      | ...     |

Overall: [LGTM / Changes Requested / Questions]
```

## Rules

- Show **real diff code** (with `+/-` lines) in review nodes, not just the final code.
- All code MUST come from actual tool output (`git diff`, `gsf callers`, `gsf trace`). Never generate or recall code from memory.
- Organize by logical flow, not file order. The reader should follow the "story" of the change.
- Focus on logic correctness and completeness, not style.
- Use `gsf callers`/`gsf trace` selectively — only when the diff needs structural context.
- Include test changes in the review. Tests are part of the change.
- For non-Go file changes (configs, docs), mention them but keep review brief.
