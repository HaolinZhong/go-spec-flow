## 1. Feature Spec Template

- [x] 1.1 Create `openspec/features/` directory and a Feature Spec template file at `openspec/features/TEMPLATE.md` showing the expected structure (feature name, PRD reference, dependency diagram, changes list with name/summary/scope/depends_on/status/key decisions)

## 2. Context Sources Convention

- [x] 2.1 Create `openspec/context/index.md` explaining the convention: what context sources are, how they're structured, how skills consume them
- [x] 2.2 Create `openspec/context/codebase-go.md` as the first context source — describes gsf tools (analyze, routes, trace, callers) and when to use them
- [x] 2.3 Create `openspec/context/service-registry.md` as a context source — describes the Service Registry at `service-registry/` and when to use it

## 3. `opsx:decompose` Skill

- [x] 3.1 Create `.claude/commands/opsx/decompose.md` skill file with YAML frontmatter and full skill instructions: input handling (file path, inline description, no input), PRD reading, context source discovery, one-shot Feature Spec generation, human adjustment loop, final save to `openspec/features/<name>/`
- [x] 3.2 Mirror the skill to `.claude/skills/openspec-decompose/SKILL.md` for the structured skill registry

## 4. `opsx:next` Skill

- [x] 4.1 Create `.claude/commands/opsx/next.md` skill file with YAML frontmatter and full skill instructions: Feature Spec selection (auto/manual), dependency graph analysis, unblocked change identification, status reconciliation with actual OpenSpec changes, kick-off of propose with Feature Spec context
- [x] 4.2 Mirror the skill to `.claude/skills/openspec-next/SKILL.md` for the structured skill registry

## 5. Integration

- [x] 5.1 Add context source discovery directive to existing `opsx:propose` skill — a single line instructing the AI to check `openspec/context/` for available context sources before generating artifacts
