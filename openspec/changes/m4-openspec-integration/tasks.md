## 1. Context Command

- [x] 1.1 Implement `gsf context` command (`internal/cmd/context.go`): combines investigate report + project structure summary + registry data into unified AI-consumable YAML/JSON output

## 2. Enhanced Skill Files

- [x] 2.1 Create `skills/gsf-propose.md`: enhanced propose skill that instructs AI to run gsf investigate/context before generating specs
- [x] 2.2 Create `skills/gsf-apply.md`: enhanced apply skill that instructs AI to use gsf trace for code context per task
- [x] 2.3 Update embedded skill files in `internal/cmd/embed_data/skills/`

## 3. gsf init Enhancement

- [x] 3.1 Update `gsf init` to also create a sample context.yaml template in service-registry/ if it doesn't exist
