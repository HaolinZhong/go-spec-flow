## 1. Create opsx:team Skill File

- [x] 1.1 Create `.claude/commands/opsx/team.md` with YAML frontmatter (name, description, category, tags) and the complete Coordinator workflow: Feature Spec selection, dependency graph analysis, batch identification, batch plan presentation, Worker launch (propose phase), proposal review loop, Worker launch (apply phase), worktree merge, Feature Spec status update, next batch loop
- [x] 1.2 Include embedded Worker Prompt Templates within the skill: one for propose phase (OpenSpec CLI commands, artifact creation guidelines, context source discovery) and one for apply phase (tasks.md reading, implementation instructions). These are needed because Workers cannot access slash commands.
- [x] 1.3 Include ambiguity handling instructions: Workers return to Coordinator on uncertainty; Coordinator presents to user; small rejection → re-run Worker with feedback; major rejection → stop team and suggest decompose adjustment
- [x] 1.4 Include worktree merge instructions: after apply, merge Worker branches into current branch; detect and report conflicts; pause for user resolution on conflict

## 2. Mirror to Skills Registry

- [x] 2.1 Create `.claude/skills/openspec-team/SKILL.md` with YAML metadata (name, description, license, compatibility, metadata) mirroring the content of `team.md`
