## ADDED Requirements

### Requirement: CLI entry point with subcommand structure
The system SHALL provide a `gsf` CLI binary with a Cobra-based subcommand structure supporting `gsf <command> [flags]` pattern.

#### Scenario: Run gsf with no arguments
- **WHEN** `gsf` is executed with no arguments
- **THEN** the system displays help text listing all available commands

#### Scenario: Run gsf with --help
- **WHEN** `gsf --help` is executed
- **THEN** the system displays detailed usage information with command descriptions

### Requirement: Output format selection
The system SHALL support multiple output formats for all analysis commands.

#### Scenario: Default output is human-readable
- **WHEN** an analysis command is run without format flags
- **THEN** the output is human-readable text suitable for terminal display

#### Scenario: JSON output for machine consumption
- **WHEN** `--format json` flag is provided
- **THEN** the output is valid JSON that can be parsed by other tools and AI assistants

#### Scenario: YAML output
- **WHEN** `--format yaml` flag is provided
- **THEN** the output is valid YAML

### Requirement: gsf routes command
The system SHALL provide a `gsf routes` command that discovers and lists all Hertz HTTP routes in a project.

#### Scenario: List routes for a project
- **WHEN** `gsf routes <project-path>` is executed
- **THEN** the system outputs all HTTP routes with method, path, and handler function

#### Scenario: Use current directory as default
- **WHEN** `gsf routes` is executed without a project path
- **THEN** the system uses the current working directory as the project root

### Requirement: gsf trace command
The system SHALL provide a `gsf trace` command that traces call chains from a specified function entry point.

#### Scenario: Trace a function
- **WHEN** `gsf trace <project-path> --entry <package>.<Function>` is executed
- **THEN** the system outputs the call chain as a tree structure

#### Scenario: Limit trace depth
- **WHEN** `gsf trace --depth <N>` is specified
- **THEN** the trace stops at depth N

### Requirement: gsf analyze command
The system SHALL provide a `gsf analyze` command that outputs the structural overview of a Go project.

#### Scenario: Analyze project structure
- **WHEN** `gsf analyze <project-path>` is executed
- **THEN** the system outputs packages, structs, interfaces, and function signatures

### Requirement: gsf init command
The system SHALL provide a `gsf init` command that installs gsf OpenSpec skills and commands into a project.

#### Scenario: Init into Claude Code project
- **WHEN** `gsf init` is executed in a project directory
- **THEN** the system copies gsf skills to `.claude/skills/` and commands to `.claude/commands/gsf/`

#### Scenario: Init into coco project
- **WHEN** `gsf init --tool coco` is executed
- **THEN** the system copies gsf skills to `.coco/skills/` and commands to `.coco/commands/gsf/`

#### Scenario: Detect tool automatically
- **WHEN** `gsf init` is executed in a project that already has a `.coco/` directory but no `.claude/` directory
- **THEN** the system auto-detects and installs to `.coco/`
