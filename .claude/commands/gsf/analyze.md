---
name: gsf:analyze
description: Analyze Go project structure (packages, structs, interfaces, functions)
---

Run `gsf analyze` on the current project to extract its structure.

Usage:
```
gsf analyze [dir] [--format text|json|yaml]
```

This will load all Go packages, extract exported types and functions, and output a structured summary.
