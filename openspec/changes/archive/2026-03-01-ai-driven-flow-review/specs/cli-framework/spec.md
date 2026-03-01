## ADDED Requirements

### Requirement: gsf diff command
The system SHALL provide a `gsf diff` command that outputs function-level change analysis with complete code snippets.

#### Scenario: Run gsf diff
- **WHEN** `gsf diff [dir]` is executed
- **THEN** the system outputs changed functions with package, name, file, line range, and complete function body

#### Scenario: Diff with commit flag
- **WHEN** `gsf diff --commit <ref> [dir]` is executed
- **THEN** the system analyzes the specified commit's changes

#### Scenario: Diff with base flag
- **WHEN** `gsf diff --base <branch> [dir]` is executed
- **THEN** the system analyzes changes between HEAD and the base branch

### Requirement: gsf callers command
The system SHALL provide a `gsf callers` command that finds direct callers of a specified function.

#### Scenario: Run gsf callers
- **WHEN** `gsf callers --pkg <package> --func <name> [dir]` is executed
- **THEN** the system outputs direct callers of the specified function

#### Scenario: Callers with format flag
- **WHEN** `gsf callers --format json` is run
- **THEN** the output is valid JSON

## REMOVED Requirements

### Requirement: gsf review command
**Reason**: Replaced by AI-orchestrated `/gsf:review` skill combined with `gsf diff`, `gsf callers`, and `gsf trace` tool commands. The hardcoded Hertz-based flow review does not generalize to non-Hertz projects.
**Migration**: Use `/gsf:review` skill which dynamically selects review strategy, or use `gsf diff` + `gsf callers` + `gsf trace` individually.
