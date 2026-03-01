## MODIFIED Requirements

### Requirement: Parse git diff and map to function-level changes
The system SHALL parse git diff output and map changed lines to Go function/method declarations, outputting each changed function with its metadata and complete source code. The system SHALL support staged changes, unstaged changes, specific commits, and base branch comparisons.

#### Scenario: Diff uncommitted changes
- **WHEN** `gsf diff [dir]` is executed with unstaged changes in the working tree and no staged changes
- **THEN** the system outputs a list of changed functions from unstaged changes

#### Scenario: Diff staged changes
- **WHEN** `gsf diff [dir]` is executed with staged changes (files added via `git add`)
- **THEN** the system outputs a list of changed functions from staged changes

#### Scenario: Diff a specific commit
- **WHEN** `gsf diff --commit HEAD [dir]` is executed
- **THEN** the system outputs changed functions between that commit and its parent

#### Scenario: Diff against a base branch
- **WHEN** `gsf diff --base main [dir]` is executed
- **THEN** the system outputs all changed functions between HEAD and the base branch

#### Scenario: No Go file changes
- **WHEN** `gsf diff` is executed but no `.go` files were changed
- **THEN** the system outputs an empty list with no error

#### Scenario: New file with multiple functions
- **WHEN** a new `.go` file is added containing multiple function declarations
- **THEN** all functions in the new file SHALL be listed with `is_new: true` and their complete code

#### Scenario: Include untracked files
- **WHEN** `gsf diff --include-untracked [dir]` is executed with new untracked `.go` files
- **THEN** the untracked files SHALL be treated as new files and all their functions listed with `is_new: true`

### Requirement: Extract complete function body from source
The system SHALL extract the full source code of each changed function using AST position information, producing exact code without truncation or modification. When a package contains both a standalone function and a method with the same name, the system SHALL correctly match based on receiver type.

#### Scenario: Extract a standalone function
- **WHEN** a standalone function `func Foo()` at lines 10-25 is changed
- **THEN** the output includes the complete source from line 10 to line 25 inclusive, including the function signature, body, and closing brace

#### Scenario: Extract a method with receiver
- **WHEN** a method `func (s *Service) Handle()` is changed
- **THEN** the output includes the receiver type in the `receiver` field and the complete method source in the `code` field

#### Scenario: Disambiguate same-name function and method
- **WHEN** a package contains both `func Create()` and `func (s *Service) Create()` and only the method is changed
- **THEN** the system SHALL extract the method's code (not the standalone function's code), matching by both name and receiver type
