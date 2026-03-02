## ADDED Requirements

### Requirement: Same-file sibling trace filtering
The builder SHALL NOT include a function as a trace child if that function is already declared as a top-level FlowNode in the same file.

#### Scenario: Sibling function removed from trace children
- **WHEN** file `builder.go` contains functions `BuildDiffTree` and `buildFuncNodesFromDiff`, and `BuildDiffTree` calls `buildFuncNodesFromDiff`
- **THEN** `buildFuncNodesFromDiff` SHALL appear as a top-level function node but SHALL NOT appear in `BuildDiffTree`'s children

#### Scenario: Cross-file calls preserved
- **WHEN** `BuildDiffTree` in `builder.go` calls `RunGitDiff` in `diff.go`
- **THEN** `RunGitDiff` SHALL still appear as a trace child of `BuildDiffTree`

### Requirement: Cross-branch seen deduplication
The builder SHALL use a per-file seen set to avoid emitting the same function with full code multiple times across different trace branches within the same file.

#### Scenario: First occurrence is complete
- **WHEN** function `readFileLines` first appears as a trace child in file `builder.go`
- **THEN** the node SHALL include full code, diff, and children fields

#### Scenario: Subsequent occurrences are references
- **WHEN** function `readFileLines` appears again as a trace child of a different function in `builder.go`
- **THEN** the node SHALL retain id, label, package, nodeType, and description, but SHALL have empty code, diff, and children

### Requirement: Dedup applies to both build modes
The deduplication logic SHALL apply to both `buildFuncNodesFromDiff` (diff mode) and `BuildCodebaseTree` (codebase mode).

#### Scenario: Diff mode dedup
- **WHEN** `gsf review --commit HEAD --json` is run
- **THEN** the output FlowTree SHALL have no duplicate function labels within any single file's trace tree

#### Scenario: Codebase mode dedup
- **WHEN** `gsf review --codebase --json` is run
- **THEN** the output FlowTree SHALL have no duplicate function labels within any single file's trace tree
