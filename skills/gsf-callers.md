---
name: gsf:callers
description: Find direct callers of a function (one level up)
---

Run `gsf callers` to find which functions call a given function.

Usage:
```
# Find callers of a function
gsf callers --pkg <package-path> --func <function-name> [dir]

# Structured output
gsf callers --pkg <package-path> --func <function-name> --format json [dir]
gsf callers --pkg <package-path> --func <function-name> --format yaml [dir]
```

Returns one level of direct callers only, showing:
- Caller package, function name
- File path and line number of the call site

Use this to understand the impact of changes — who calls the modified function?
To trace deeper, call `gsf callers` again on each caller you want to investigate.
