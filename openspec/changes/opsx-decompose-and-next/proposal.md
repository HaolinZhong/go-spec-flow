## Why

Large requirements often fail when fed directly into `opsx:propose` — the resulting change becomes unwieldy with too many tasks, and the AI struggles to maintain coherence. We need an intermediate mechanism that breaks a PRD into multiple focused changes, each small enough for a single propose→apply cycle. This also lays the groundwork for semi-automatic execution (sequential via `opsx:next`) and future parallel execution (`opsx:team`).

## What Changes

- **New `opsx:decompose` skill**: Reads a PRD, analyzes the codebase (leveraging available context sources), and generates a Feature Spec (L1 artifact) that decomposes the requirement into multiple changes with dependency relationships.
- **New `opsx:next` skill**: Reads a Feature Spec, identifies the next unblocked change, and kicks off `opsx:propose` with Feature Spec context. Enables semi-automatic sequential execution of decomposed changes.
- **New Feature Spec artifact**: A formal markdown artifact at `openspec/features/<name>/feature-spec.md` that tracks changes, dependencies, and status for a decomposed requirement.
- **New pluggable context sources convention**: An `openspec/context/` directory where context providers (like gsf for Go, Service Registry for RPC) register themselves as markdown descriptions. Any skill can discover and use them without hard coupling.

## Capabilities

### New Capabilities
- `feature-decompose`: PRD decomposition into Feature Spec with multiple changes and dependency tracking
- `feature-execute`: Semi-automatic sequential execution of Feature Spec changes via `opsx:next`
- `context-sources`: Pluggable context source discovery convention for skill-agnostic codebase awareness

### Modified Capabilities
<!-- No existing capability requirements are changing -->

## Impact

- **New files**: 2 skill files (`.claude/commands/opsx/decompose.md`, `.claude/commands/opsx/next.md`), context convention (`openspec/context/index.md`), Feature Spec template
- **No code changes**: All deliverables are markdown skills and conventions — no Go code modifications
- **Dependencies**: Requires `openspec` CLI for change management (already present)
- **Ecosystem**: Establishes the L1 artifact layer between PRD and OpenSpec changes, completing the PRD→Spec→Code pipeline described in the project blueprint
