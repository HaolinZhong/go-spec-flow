## ADDED Requirements

### Requirement: Decompose PRD into Feature Spec
The system SHALL provide an `opsx:decompose` skill that reads a PRD document and generates a Feature Spec (L1 artifact) decomposing the requirement into multiple OpenSpec changes.

#### Scenario: Decompose with PRD file path
- **WHEN** user invokes `/opsx:decompose path/to/prd.md`
- **THEN** the skill reads the PRD file, analyzes it, and generates a Feature Spec at `openspec/features/<derived-name>/feature-spec.md`
- **AND** copies the PRD to `openspec/features/<derived-name>/prd.md`

#### Scenario: Decompose with inline description
- **WHEN** user invokes `/opsx:decompose` with a text description instead of a file path
- **THEN** the skill treats the description as the requirement and generates a Feature Spec

#### Scenario: Decompose without input
- **WHEN** user invokes `/opsx:decompose` with no argument
- **THEN** the skill asks the user what requirement they want to decompose

### Requirement: Feature Spec structure
The Feature Spec artifact SHALL contain: a feature name, PRD reference, dependency overview diagram, and a list of changes. Each change entry SHALL include: name (kebab-case), summary, scope, dependencies (list of other change names), status, and optional key decisions.

#### Scenario: Feature Spec with independent changes
- **WHEN** a PRD decomposes into changes with no inter-dependencies
- **THEN** the Feature Spec lists all changes with empty `depends_on` fields

#### Scenario: Feature Spec with dependent changes
- **WHEN** a PRD decomposes into changes where some depend on others
- **THEN** the Feature Spec lists dependencies using the change name references
- **AND** includes an ASCII dependency diagram in the overview section

### Requirement: Context source utilization during decompose
The `opsx:decompose` skill SHALL check `openspec/context/` for available context sources and use them to improve decomposition quality.

#### Scenario: Context sources available
- **WHEN** `openspec/context/index.md` exists and lists context sources
- **THEN** the skill reads relevant context source descriptions and uses the described tools/files to gather codebase information before generating the decomposition

#### Scenario: No context sources
- **WHEN** `openspec/context/` does not exist or is empty
- **THEN** the skill proceeds with decomposition using only the PRD and direct code reading

### Requirement: Human adjustment of decomposition
After generating the initial Feature Spec, the skill SHALL present it to the user and accept adjustments (merge changes, split changes, add new changes, reorder dependencies, rename changes).

#### Scenario: User adjusts decomposition
- **WHEN** the skill presents the generated Feature Spec
- **AND** user requests modifications (e.g., "merge B and C", "add a new change for X")
- **THEN** the skill updates the Feature Spec accordingly and presents the updated version

#### Scenario: User approves decomposition
- **WHEN** the skill presents the generated Feature Spec
- **AND** user approves it
- **THEN** the skill saves the final Feature Spec to `openspec/features/<name>/feature-spec.md`

### Requirement: Change status tracking
Each change in the Feature Spec SHALL have a status field with values: `pending`, `proposed`, `completed`.

#### Scenario: Initial status
- **WHEN** a Feature Spec is first created
- **THEN** all changes have status `pending`

#### Scenario: Status after propose
- **WHEN** a change has been proposed via `opsx:propose`
- **THEN** its status in the Feature Spec is updated to `proposed`

#### Scenario: Status after completion
- **WHEN** a change has been applied and archived
- **THEN** its status in the Feature Spec is updated to `completed`
