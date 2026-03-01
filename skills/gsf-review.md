---
name: gsf:review
description: AI-driven flow-based code review using gsf tools
---

You are performing a flow-based code review. Use gsf tools to gather precise code facts, then organize them into a human-friendly review document.

## Available Tools

### gsf diff
Lists all changed functions with their complete source code.
```
gsf diff [dir]                          # uncommitted changes
gsf diff --commit HEAD [dir]            # last commit
gsf diff --base main [dir]             # changes vs base branch
gsf diff --format yaml [dir]           # structured output
```

### gsf callers
Finds direct callers of a function (one level up).
```
gsf callers --pkg <package-path> --func <function-name> [dir]
gsf callers --pkg <package-path> --func <function-name> --format yaml [dir]
```

### gsf trace
Traces call chains downward from a function.
```
gsf trace --pkg <package-path> --func <function-name> [dir]
gsf trace --pkg <package-path> --func <function-name> --depth 5 [dir]
```

## Review Flow

### Step 1: Get the change overview
Run `gsf diff --format yaml` to get all changed functions with their code.

### Step 2: Analyze change characteristics
Based on the diff output, determine:
- **Change type**: new feature, bugfix, refactor, small fix
- **Scope**: single function, module-level, cross-module
- **Entry points**: which changed functions are likely entry points

### Step 3: Gather context based on change type

**New feature** (new functions/methods added):
- Use `gsf trace` on new entry-point functions to see what they call
- Verify the implementation chain is complete

**Bugfix** (existing functions modified):
- Use `gsf callers` on modified functions to assess impact
- Use `gsf trace` to follow the fix path and verify correctness

**Refactor / signature change**:
- Use `gsf callers` to verify all call sites are updated
- Check for any missing adaptations

**Small change** (minor modifications):
- Direct review of the changed code may be sufficient
- Use `gsf callers` only if the change could affect callers

### Step 4: Organize the review document

Structure your review as a **flow** (not a file-by-file diff):
1. Start from the entry point or most significant change
2. Follow the logical flow (request path, data flow, or impact path)
3. At each node, show the code and your observations
4. Note risks, suggestions, and questions

## Output Format

```markdown
# Flow Review: [brief description of what changed]

## Change Overview
- Type: [new feature / bugfix / refactor / small fix]
- Scope: [N functions across M packages]
- Key changes: [1-2 sentence summary]

## Review Flow

### 1. [Entry point or most significant change]
[Code snippet from gsf diff]
[Your observations, risks, suggestions]

### 2. [Next step in the flow]
[Code snippet]
[Observations]

...

## Impact Assessment
- Callers affected: [from gsf callers output]
- Downstream dependencies: [from gsf trace output]

## Summary
- [Key findings]
- [Suggestions]
- [Questions for the author]
```

## Rules
- All code snippets MUST come from gsf tool output. Never generate or recall code from memory.
- Use `gsf callers` to understand who is affected, not to list every caller.
- One level of callers at a time. Decide whether to recurse based on what you see.
- Focus on logic correctness, not style. The review should help the author catch real issues.
