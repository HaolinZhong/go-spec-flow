---
name: gsf:fix
description: Apply code modifications from review comments
---

Apply code changes based on review comments exported from `gsf review`.

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

For each comment:

1. **Read the target file** at the specified path
2. **Locate the line** using the `line` number. If the code at that line doesn't match `codeContext`, search for `codeContext` in the file to find the correct location
3. **Understand the intent** of the comment — it's free-form text describing what the user wants changed
4. **Make the modification** as described
5. **Show what was changed** with before/after

```
### Processing 3/7: internal/review/builder.go:128
Comment: "这里的 key 应该考虑 receiver type"
Code context: seen[pkgPath+"."+funcName] = true

→ Modified: Added receiver type to seen key
```

If a comment's intent is unclear, ask the user for clarification before proceeding.

## Step 4: Completion Summary

After all comments are processed:

```
## Fix Complete

**Processed:** 7/7 comments
**Files modified:** 3

### Changes Made
- internal/review/builder.go: 3 modifications
- internal/review/diff.go: 2 modifications
- internal/cmd/review.go: 2 modifications

You can delete review-comments.json if no longer needed.
```

## Rules

- Read each target file before modifying — never guess at code structure
- Use `codeContext` as a fallback when line numbers are stale
- Keep modifications minimal and focused on what the comment asks
- If a comment describes a problem without a clear solution, ask the user
- Process comments in file order to minimize file re-reads
- Do NOT auto-delete review-comments.json — let the user decide
