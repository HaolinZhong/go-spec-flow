---
name: gsf:fix
description: Process review comments — answer questions or apply code modifications
---

Process review comments from `gsf review`. Comments can be **questions** (asking about code behavior, design decisions, etc.) or **modification requests** (asking for code changes). Handle each appropriately.

## Step 1: Read Comment File

Read `review-comments.json` from the project root:

```bash
cat review-comments.json
```

If the file doesn't exist, inform the user:
```
No review-comments.json found. Run a review first:
1. gsf review --render flow.json --serve
2. Add comments by clicking line numbers
3. Comments are auto-saved to review-comments.json
```

## Step 2: Display Comment Summary

Show all comments grouped by file:

```
## Review Comments (N total)

### file1.go (M comments)
- Line 42: "comment text here"
- Line 85: "another comment"

### file2.go (K comments)
- Line 10: "comment text"
```

Ask the user: "Process all comments, or select specific ones?"

## Step 3: Process Comments

For each comment, first **classify** it, then handle accordingly:

### Questions (asking about code)

If the comment is a question (contains "?", "为什么", "是不是", "有没有", "什么作用", "怎么", etc.), **answer it**:

1. **Read the target file** and surrounding context
2. **Research** — trace call chains, find callers, read related code as needed
3. **Answer thoroughly** — explain the code's purpose, design rationale, trade-offs

```
### 1/3: internal/ast/callers.go:11
Comment: "这里查找调用者，起到了什么样的作用？"

→ Answer: FindCallers 在 review 模块中用于...
```

### Modification Requests (asking for changes)

If the comment requests a change ("应该", "改成", "添加", "删除", "重构", etc.):

1. **Read the target file** at the specified path
2. **Locate the line** using the `line` number. If the code at that line doesn't match `codeContext`, search for `codeContext` in the file to find the correct location
3. **Understand the intent** of the comment
4. **Make the modification** as described
5. **Show what was changed** with before/after

```
### 3/7: internal/review/builder.go:128
Comment: "这里的 key 应该考虑 receiver type"
Code context: seen[pkgPath+"."+funcName] = true

→ Modified: Added receiver type to seen key
```

### Ambiguous Comments

If a comment's intent is unclear (could be a question or a request), ask the user for clarification before proceeding.

## Step 4: Completion Summary

After all comments are processed:

```
## Processing Complete

**Processed:** 7/7 comments
**Questions answered:** 3
**Files modified:** 2

### Answers
- internal/ast/callers.go:11 — explained FindCallers' role in review context
- ...

### Changes Made
- internal/review/builder.go: 2 modifications
- internal/cmd/review.go: 1 modification

You can delete review-comments.json if no longer needed.
```

## Rules

- Read each target file before modifying — never guess at code structure
- Use `codeContext` as a fallback when line numbers are stale
- Keep modifications minimal and focused on what the comment asks
- If a comment describes a problem without a clear solution, ask the user
- Process comments in file order to minimize file re-reads
- Do NOT auto-delete review-comments.json — let the user decide
