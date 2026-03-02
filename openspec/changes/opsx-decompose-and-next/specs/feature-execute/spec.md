## ADDED Requirements

### Requirement: Find next unblocked change
The `opsx:next` skill SHALL read a Feature Spec and identify changes that are eligible for execution: status is `pending` and all `depends_on` entries have status `completed`.

#### Scenario: Single unblocked change
- **WHEN** user invokes `/opsx:next`
- **AND** exactly one change is unblocked
- **THEN** the skill presents that change and asks if the user wants to proceed with propose

#### Scenario: Multiple unblocked changes
- **WHEN** user invokes `/opsx:next`
- **AND** multiple changes are unblocked
- **THEN** the skill presents all unblocked changes and lets the user choose which to work on next

#### Scenario: No unblocked changes
- **WHEN** user invokes `/opsx:next`
- **AND** all pending changes have unsatisfied dependencies
- **THEN** the skill reports the blocking situation and shows which changes need to complete first

#### Scenario: All changes completed
- **WHEN** user invokes `/opsx:next`
- **AND** all changes in the Feature Spec have status `completed`
- **THEN** the skill congratulates the user and suggests archiving the feature

### Requirement: Feature Spec selection
The `opsx:next` skill SHALL identify which Feature Spec to work from.

#### Scenario: Single active feature
- **WHEN** only one Feature Spec exists in `openspec/features/` with incomplete changes
- **THEN** the skill auto-selects it and announces which feature it's using

#### Scenario: Multiple active features
- **WHEN** multiple Feature Specs exist with incomplete changes
- **THEN** the skill asks the user to select which feature to work on

#### Scenario: Feature name provided
- **WHEN** user invokes `/opsx:next <feature-name>`
- **THEN** the skill uses the specified feature

### Requirement: Kick off propose with Feature Spec context
When the user confirms a change selection, `opsx:next` SHALL initiate the propose workflow with additional context from the Feature Spec.

#### Scenario: Propose with Feature Spec context
- **WHEN** user confirms they want to propose a change
- **THEN** the skill invokes `opsx:propose` for that change name
- **AND** provides the Feature Spec context (overall feature goal, this change's summary and scope, dependency context from completed changes)

#### Scenario: Propose for change with completed dependencies
- **WHEN** the selected change depends on other changes that are already completed
- **THEN** the skill includes a summary of what the completed dependency changes accomplished, so the proposal can build on them

### Requirement: Status reconciliation
The `opsx:next` skill SHALL reconcile Feature Spec status with actual OpenSpec change status.

#### Scenario: Change archived but Feature Spec says proposed
- **WHEN** `opsx:next` detects that a change exists in `openspec/changes/archive/` but the Feature Spec shows status `proposed`
- **THEN** the skill auto-corrects the Feature Spec status to `completed`

#### Scenario: Change exists as active but Feature Spec says pending
- **WHEN** `opsx:next` detects that a change directory exists in `openspec/changes/` but the Feature Spec shows status `pending`
- **THEN** the skill auto-corrects the Feature Spec status to `proposed`
