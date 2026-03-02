## ADDED Requirements

### Requirement: Parse git diff and map to function-level changes
The system SHALL parse git diff output and map changed lines to Go function/method declarations, outputting each changed function with its metadata and complete source code.

#### Scenario: Diff uncommitted changes
- **WHEN** `gsf diff [dir]` is executed with uncommitted changes in the working tree
- **THEN** the system outputs a list of changed functions with package, name, receiver, file path, line range, is_new flag, and complete function body code

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

### Requirement: Extract complete function body from source
The system SHALL extract the full source code of each changed function using AST position information, producing exact code without truncation or modification.

#### Scenario: Extract a standalone function
- **WHEN** a standalone function `func Foo()` at lines 10-25 is changed
- **THEN** the output includes the complete source from line 10 to line 25 inclusive, including the function signature, body, and closing brace

#### Scenario: Extract a method with receiver
- **WHEN** a method `func (s *Service) Handle()` is changed
- **THEN** the output includes the receiver type in the `receiver` field and the complete method source in the `code` field

### Requirement: Support multiple output formats
The system SHALL support text, JSON, and YAML output formats for the diff command.

#### Scenario: Default text output
- **WHEN** `gsf diff` is run without `--format` flag
- **THEN** the output is human-readable text listing changed functions with their code

#### Scenario: JSON output for AI consumption
- **WHEN** `gsf diff --format json` is run
- **THEN** the output is valid JSON containing a `changed_functions` array with all metadata and code fields

#### Scenario: YAML output
- **WHEN** `gsf diff --format yaml` is run
- **THEN** the output is valid YAML with the same structure as JSON output
