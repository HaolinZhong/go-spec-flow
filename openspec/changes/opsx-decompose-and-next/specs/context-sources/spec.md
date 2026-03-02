## ADDED Requirements

### Requirement: Context source registration
The system SHALL support registering context sources as markdown files in `openspec/context/`. Each context source file describes what information it provides and how to access it.

#### Scenario: Context source file structure
- **WHEN** a context source is registered
- **THEN** it exists as a markdown file at `openspec/context/<source-name>.md`
- **AND** contains sections: what it provides, when to use it, and how to invoke it

### Requirement: Context source index
The system SHALL maintain an `openspec/context/index.md` file that lists all available context sources with brief descriptions.

#### Scenario: Index with multiple sources
- **WHEN** multiple context sources are registered
- **THEN** `openspec/context/index.md` lists each source with its name, brief description, and applicable workflow stages

#### Scenario: Empty or missing index
- **WHEN** `openspec/context/index.md` does not exist
- **THEN** skills treat this as "no context sources available" and proceed without enhanced context

### Requirement: Skill consumption of context sources
Skills that benefit from codebase awareness (decompose, propose, apply) SHALL include an instruction to check `openspec/context/` for available context sources.

#### Scenario: Skill discovers context sources
- **WHEN** a skill's instructions include the context source discovery directive
- **AND** `openspec/context/index.md` exists
- **THEN** the AI agent reads the index and relevant source files, and uses the described tools/files as appropriate for the current task

#### Scenario: Graceful degradation
- **WHEN** a skill's instructions include the context source discovery directive
- **AND** `openspec/context/` does not exist
- **THEN** the AI agent proceeds without enhanced context, using direct code reading as fallback
