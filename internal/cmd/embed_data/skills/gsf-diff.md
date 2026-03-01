---
name: gsf:diff
description: Show function-level code changes with complete source
---

Run `gsf diff` to see which functions changed and their complete code.

Usage:
```
# Uncommitted changes
gsf diff [dir]

# Specific commit
gsf diff --commit HEAD [dir]

# Changes vs base branch
gsf diff --base main [dir]

# Structured output for further analysis
gsf diff --format json [dir]
gsf diff --format yaml [dir]
```

Output includes for each changed function:
- Package path, function name, receiver type (if method)
- File path and line range
- Whether the function is new or modified
- Complete function source code
