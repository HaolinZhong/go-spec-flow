## ADDED Requirements

### Requirement: AI-orchestrated flow-based code review skill
The system SHALL provide a `/gsf:review` skill that teaches AI to orchestrate flow-based code review using gsf tool commands, dynamically choosing review strategy based on change characteristics.

#### Scenario: Review with gsf diff as starting point
- **WHEN** `/gsf:review` is invoked
- **THEN** the skill instructs AI to first run `gsf diff --format yaml` to get the full picture of changed functions with their code

#### Scenario: AI determines review strategy
- **WHEN** AI receives the diff output
- **THEN** AI analyzes the change nature (new feature, bugfix, refactor, small fix) and decides which gsf tools to call next (trace for downstream impact, callers for upstream impact, or direct review)

#### Scenario: New feature review strategy
- **WHEN** the change introduces new functions/methods
- **THEN** AI uses `gsf trace` on new entry-point functions to verify implementation completeness and follow the call chain

#### Scenario: Bugfix review strategy
- **WHEN** the change modifies existing functions
- **THEN** AI uses `gsf callers` to check impact on callers, and `gsf trace` to verify the fix path

#### Scenario: Refactor/signature change strategy
- **WHEN** the change renames or modifies function signatures
- **THEN** AI uses `gsf callers` to verify all call sites have been updated

### Requirement: Code snippets from tools only
The skill SHALL instruct AI to use only code snippets provided by gsf tool output in the review document, never generating or hallucinating code from memory.

#### Scenario: Code reference in review output
- **WHEN** AI writes a review document referencing a changed function
- **THEN** the code snippet is taken directly from `gsf diff` output's `code` field

#### Scenario: Upstream context code
- **WHEN** AI needs to show a caller's code for context
- **THEN** AI runs `gsf callers` or `gsf trace` to obtain the code, rather than generating it

### Requirement: Skill file installed by gsf init
The system SHALL include the `/gsf:review` skill file in the set of skills installed by `gsf init`.

#### Scenario: gsf init installs review skill
- **WHEN** `gsf init` is executed in a project directory
- **THEN** the review skill file is copied to `.claude/skills/` (or `.coco/skills/` for coco)

#### Scenario: Skill references available gsf commands
- **WHEN** the review skill file is installed
- **THEN** it documents the available commands (`gsf diff`, `gsf callers`, `gsf trace`) with their flags and output formats
