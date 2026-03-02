## ADDED Requirements

### Requirement: FlowTree and FlowNode support description field
FlowTree SHALL have a `Description` field (string, JSON `omitempty`) for the overall review/flow description.
FlowNode SHALL have a `Description` field (string, JSON `omitempty`) for per-node AI commentary.

#### Scenario: gsf raw output has no descriptions
- **WHEN** gsf generates a FlowTree without AI orchestration
- **THEN** Description fields SHALL be empty and omitted from JSON output

#### Scenario: AI-orchestrated JSON has descriptions
- **WHEN** AI reads raw JSON and adds descriptions to FlowTree and FlowNode
- **THEN** the descriptions SHALL be preserved when read back by `gsf review --render`

### Requirement: JSON output mode via --json flag
`gsf review` SHALL support a `--json` flag that outputs the FlowTree as JSON to stdout instead of generating HTML.
The `--json` flag SHALL work with both diff mode and codebase mode.
When `--json` is used, `--output` and `--open` flags SHALL be ignored.

#### Scenario: Diff mode JSON output
- **WHEN** user runs `gsf review --commit HEAD --json`
- **THEN** gsf SHALL output valid JSON to stdout representing the diff FlowTree with code content

#### Scenario: Codebase mode JSON output
- **WHEN** user runs `gsf review --codebase --json`
- **THEN** gsf SHALL output valid JSON to stdout representing the codebase FlowTree with all packages

#### Scenario: JSON output to file via redirect
- **WHEN** user runs `gsf review --codebase --json > raw.json`
- **THEN** the file SHALL contain valid, parseable JSON identical to stdout output

### Requirement: HTML rendering from JSON via --render flag
`gsf review` SHALL support a `--render <file>` flag that reads a JSON file and renders it as HTML.
The `--render` flag SHALL be mutually exclusive with `--commit`, `--base`, and `--codebase`.
The `--render` flag SHALL work with `--output` and `--open` flags.

#### Scenario: Render AI-orchestrated JSON
- **WHEN** user runs `gsf review --render flow.json --open`
- **THEN** gsf SHALL read flow.json, parse it as a FlowTree, render HTML, and open in browser

#### Scenario: Render JSON with descriptions
- **WHEN** the JSON contains FlowTree and FlowNode with non-empty Description fields
- **THEN** the rendered HTML SHALL display descriptions in the appropriate UI locations

#### Scenario: Invalid JSON file
- **WHEN** user runs `gsf review --render invalid.json`
- **THEN** gsf SHALL exit with a clear error message indicating the JSON is invalid

### Requirement: HTML template displays descriptions
The HTML template SHALL render FlowTree.Description as a summary paragraph below the review title.
The HTML template SHALL render FlowNode.Description as a commentary block above the code panel when a node is selected.
When Description is empty, no commentary block SHALL be shown.

#### Scenario: Flow-level description display
- **WHEN** a FlowTree has a non-empty Description
- **THEN** the HTML SHALL show the description as a styled paragraph below the title in the header area

#### Scenario: Node-level description display
- **WHEN** a user clicks a FlowNode that has a non-empty Description
- **THEN** the right panel SHALL show the description text above the code, visually distinct from code

#### Scenario: No description
- **WHEN** a FlowNode has empty Description
- **THEN** the right panel SHALL show only the code, with no empty commentary block

### Requirement: Review skill orchestrates three-step flow
The `/gsf:review` skill SHALL implement a three-step pipeline:
1. Run `gsf review` with `--json` to extract structural data
2. AI reads the JSON, reorganizes into meaningful flows, adds descriptions
3. AI writes the orchestrated JSON and runs `gsf review --render` to produce HTML

The skill SHALL ask the user for review scope before starting.

#### Scenario: Full codebase review with AI orchestration
- **WHEN** user invokes `/gsf:review` and selects "整个 codebase"
- **THEN** the skill SHALL run `gsf review --codebase --json`, AI SHALL reorganize the tree into named flows with descriptions, and render via `gsf review --render`

#### Scenario: Diff review with AI orchestration
- **WHEN** user invokes `/gsf:review` and selects "最近一次 commit"
- **THEN** the skill SHALL run `gsf review --commit HEAD --json`, AI SHALL add flow context and descriptions, and render via `gsf review --render`
