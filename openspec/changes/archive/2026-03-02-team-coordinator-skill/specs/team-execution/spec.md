## ADDED Requirements

### Requirement: Coordinator reads Feature Spec and builds dependency graph
The `opsx:team` skill SHALL read the specified Feature Spec from `openspec/features/<name>/feature-spec.md`, parse all changes with their dependencies and statuses, and construct a dependency graph to determine execution batches.

#### Scenario: Feature Spec with multiple independent changes
- **WHEN** the Feature Spec contains changes A, B, C where A and B have no dependencies and C depends on A
- **THEN** the Coordinator SHALL identify batch 1 as [A, B] and batch 2 as [C]

#### Scenario: Feature Spec with all sequential dependencies
- **WHEN** each change depends on the previous one (A → B → C)
- **THEN** the Coordinator SHALL execute them in 3 single-change batches

#### Scenario: Feature Spec not found
- **WHEN** the specified feature name does not match any directory in `openspec/features/`
- **THEN** the Coordinator SHALL inform the user and suggest `/opsx:decompose`

### Requirement: Coordinator presents batch plan for user confirmation
Before executing each batch, the Coordinator SHALL present a summary showing: completed changes, current batch changes, remaining blocked changes, and overall progress. The user MUST confirm before Workers are launched.

#### Scenario: First batch with 3 parallel changes
- **WHEN** 3 changes are unblocked and ready to start
- **THEN** the Coordinator SHALL show all 3 with their summaries and ask the user to confirm starting the batch

#### Scenario: User declines batch
- **WHEN** the user does not confirm the batch plan
- **THEN** the Coordinator SHALL stop and wait for further instructions

### Requirement: Workers execute in isolated worktrees
Each Worker SHALL be launched as a Claude Code Agent with `isolation: "worktree"`, giving it an independent copy of the repository. Workers in the same batch SHALL run in parallel.

#### Scenario: Batch with 2 changes
- **WHEN** the Coordinator starts a batch with changes A and B
- **THEN** two Agent calls SHALL be made in parallel, each with `isolation: "worktree"`

#### Scenario: Worker creates artifacts
- **WHEN** a Worker is launched for propose phase
- **THEN** the Worker SHALL create an OpenSpec change with proposal.md, design.md, specs, and tasks.md in its worktree

### Requirement: Workers receive complete context via prompt
Each Worker SHALL receive a prompt containing: the feature overview, this change's summary/scope/key decisions, completed dependency information, and the full propose or apply workflow steps (since Workers cannot access slash commands).

#### Scenario: Worker prompt for propose
- **WHEN** a Worker is launched for the propose phase
- **THEN** the prompt SHALL include OpenSpec CLI commands (new change, instructions, status) and artifact creation guidelines

#### Scenario: Worker prompt for apply
- **WHEN** a Worker is launched for the apply phase
- **THEN** the prompt SHALL include instructions to read tasks.md and implement each task

### Requirement: Human reviews each proposal sequentially
After all Workers in a batch complete their propose phase, the Coordinator SHALL present each proposal to the user for sequential review. No proposal SHALL be applied without explicit user approval.

#### Scenario: All proposals approved
- **WHEN** the user approves all proposals in a batch
- **THEN** the Coordinator SHALL proceed to parallel apply phase

#### Scenario: Proposal rejected with small adjustments
- **WHEN** the user requests changes to a specific proposal
- **THEN** the Coordinator SHALL re-run that Worker with the feedback to revise the proposal

#### Scenario: Proposal rejected with major concerns
- **WHEN** the user indicates the decomposition itself needs rethinking
- **THEN** the Coordinator SHALL stop the team and suggest revisiting the Feature Spec via `/opsx:decompose`

### Requirement: Workers apply approved changes in parallel
After all proposals in a batch are approved, the Coordinator SHALL launch Workers in parallel to apply (implement) the approved changes, each in its worktree.

#### Scenario: Parallel apply of 2 approved changes
- **WHEN** 2 proposals are approved
- **THEN** 2 Agent calls SHALL be made in parallel for the apply phase, each with `isolation: "worktree"`

### Requirement: Coordinator manages worktree merging
After Workers complete the apply phase, the Coordinator SHALL guide the merging of worktree branches into the current branch. Merge conflicts SHALL cause the Coordinator to pause and inform the user.

#### Scenario: Clean merge
- **WHEN** a Worker's apply completes and its worktree branch has no conflicts with the current branch
- **THEN** the Coordinator SHALL merge the worktree branch

#### Scenario: Merge conflict
- **WHEN** merging a worktree branch produces conflicts
- **THEN** the Coordinator SHALL stop, report the conflicting files, and wait for user resolution

### Requirement: Coordinator updates Feature Spec status
After each batch completes (apply + merge), the Coordinator SHALL update the Feature Spec, setting completed changes' status to `completed` and identifying the next batch of unblocked changes.

#### Scenario: Batch completed
- **WHEN** all changes in a batch are applied and merged
- **THEN** the Feature Spec SHALL be updated with `status: completed` for those changes

#### Scenario: All changes completed
- **WHEN** the last batch completes and no pending changes remain
- **THEN** the Coordinator SHALL announce feature completion

### Requirement: Workers pause on ambiguity
Workers SHALL NOT make unilateral decisions when encountering ambiguity. If a Worker encounters unclear requirements, conflicting code, or any situation requiring human judgment, it SHALL return with a description of the issue rather than proceeding.

#### Scenario: Ambiguous requirement during propose
- **WHEN** a Worker cannot determine the correct approach for a design decision
- **THEN** the Worker SHALL return to the Coordinator describing the ambiguity, and the Coordinator SHALL present it to the user

### Requirement: Compatible with opsx:next
The `opsx:team` skill SHALL use the same Feature Spec format as `opsx:next`. Users SHALL be able to switch between sequential (`opsx:next`) and parallel (`opsx:team`) execution at any point.

#### Scenario: Switch from team to next mid-feature
- **WHEN** a user has completed 2 of 5 changes via `opsx:team` and wants to continue with `opsx:next`
- **THEN** `opsx:next` SHALL correctly read the Feature Spec with 2 completed changes and propose the next unblocked one
