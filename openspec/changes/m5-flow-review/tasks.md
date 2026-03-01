## 1. Git Diff Parser

- [x] 1.1 Implement git diff parser (`internal/review/diff.go`): run git diff, parse unified diff output, extract changed file paths and line ranges

## 2. Diff-to-Function Mapping

- [x] 2.1 Implement function mapper (`internal/review/mapper.go`): given changed line ranges and AST, determine which functions/methods were modified or are new

## 3. Flow Review Engine

- [x] 3.1 Implement flow review engine (`internal/review/flow.go`): overlay change markers on call chain tree, collect standalone changes
- [x] 3.2 Implement CLI tree renderer with change annotations

## 4. Review Command

- [x] 4.1 Implement `gsf review` command (`internal/cmd/review.go`) with --commit and --base flags
