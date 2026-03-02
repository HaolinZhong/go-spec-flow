## Context

The current OpenSpec workflow (`explore → propose → apply → archive`) handles individual changes well but has no mechanism for decomposing large requirements. When a PRD is too big for a single change, users must manually figure out how to split it — which requires deep codebase knowledge and is error-prone.

The project blueprint defines a three-level spec hierarchy (L1 Feature Spec → L2 Design Spec → L3 Task Spec) but only L2 and L3 are currently implemented via `opsx:propose` (which generates design.md and tasks.md). The L1 layer — decomposing a PRD into multiple changes — is missing.

All deliverables are **markdown skills and conventions** (no Go code). Skills are AI-readable instructions that guide the agent through a workflow.

## Goals / Non-Goals

**Goals:**
- Enable PRD → multiple OpenSpec changes decomposition via `opsx:decompose`
- Provide semi-automatic sequential execution via `opsx:next`
- Establish a pluggable context source convention that any skill can leverage
- Keep everything generic (opsx-level), not tied to Go or gsf

**Non-Goals:**
- Parallel execution (`opsx:team`) — future phase, not this change
- Automatic dependency resolution or smart scheduling
- Go code or CLI tool changes
- Modifying existing opsx skills (propose/apply/archive)

## Decisions

### 1. Feature Spec format: Markdown with structured sections

**Decision**: Feature Spec is a markdown file with a defined section structure (overview, dependency diagram, per-change sections with status/scope/dependencies).

**Why not YAML?** Markdown is human-readable and writable, AI-friendly, and consistent with all other OpenSpec artifacts. YAML would add a parsing dependency.

**Why not a formal schema?** The Feature Spec is consumed by AI skills (which parse markdown naturally) and humans. Adding schema validation would be over-engineering at this stage.

### 2. Feature Spec location: `openspec/features/<name>/`

**Decision**: Feature Specs live in `openspec/features/<name>/` alongside a copy of or link to the original PRD.

```
openspec/
├── features/              ← NEW
│   └── <feature-name>/
│       ├── feature-spec.md
│       └── prd.md         ← Original PRD (copied)
├── changes/               ← Existing (each change from decomposition)
└── specs/                 ← Existing
```

**Why separate from changes?** A feature is a level above changes — it contains multiple changes. Mixing them would confuse the hierarchy.

**Why copy the PRD?** Rather than referencing an external path that may move, keep a copy alongside the Feature Spec for self-contained context.

### 3. Decompose process: One-shot generation + human adjustment

**Decision**: AI reads the PRD, optionally uses context sources, and generates a complete Feature Spec in one pass. User then adjusts (merge changes, add new ones, reorder dependencies).

**Why not conversational?** The conversational approach (AI asks questions iteratively) is slower and what `opsx:explore` already provides. Users can explore first, then decompose. One-shot is faster and the user can correct errors more efficiently than answering questions.

### 4. `opsx:next` scope: Propose only

**Decision**: `opsx:next` finds the next unblocked change and kicks off `opsx:propose`. It does NOT auto-run `opsx:apply`.

**Why?** The user explicitly chose "semi-automatic with human confirmation at each step." Propose and apply are separate decisions — the user needs time to review the proposal before committing to implementation. The loop is: `next → (review proposal) → apply → next → ...`

### 5. Context sources: File-based discovery via `openspec/context/`

**Decision**: Context sources are registered as markdown files in `openspec/context/`. Each file describes what the source provides and how to invoke it. An `index.md` provides an overview.

```
openspec/context/
├── index.md              ← Overview of available context sources
├── codebase-go.md        ← "gsf analyze/routes/trace available"
└── service-registry.md   ← "Service Registry at service-registry/"
```

**Why markdown files?** Skills already instruct AI agents via markdown. Context source descriptions are the same — telling the AI what tools are available and when to use them. No plugin mechanism needed.

**How skills consume them**: Each skill that benefits from context (decompose, propose, etc.) includes one line: "Check `openspec/context/` for available context sources and use them as appropriate." The AI reads index.md, discovers what's available, and decides what to use.

### 6. Change status tracking in Feature Spec

**Decision**: Each change in the Feature Spec has a `status` field managed by `opsx:next`. Statuses: `pending` → `proposed` → `completed`.

`opsx:next` determines "unblocked" by checking: status is `pending` AND all `depends_on` entries have status `completed`. When a user finishes `opsx:apply` + `opsx:archive` for a change, they (or `opsx:next` at next invocation) update the Feature Spec.

**Why not use OpenSpec's own change status?** The Feature Spec tracks a higher-level view. An OpenSpec change might be in various internal states (artifacts incomplete, partially applied), but from the Feature Spec's perspective, it's either pending, actively being worked on (proposed), or done.

## Risks / Trade-offs

**[Feature Spec gets out of sync with actual changes]** → `opsx:next` always reads both the Feature Spec and checks actual OpenSpec change status (`openspec list --json`) to reconcile. If a change has been archived but the Feature Spec still says "proposed", next auto-corrects.

**[AI generates poor decomposition]** → Mitigated by human review. The one-shot approach means the user sees the full picture at once and can fix it. Context sources (when available) improve decomposition quality by grounding it in actual code structure.

**[Feature Spec format ambiguity]** → Mitigated by providing a clear template. The skill instructions include the expected format with examples. AI agents are reliable at following markdown templates.

**[Context sources not available]** → Graceful degradation. If `openspec/context/` is empty or doesn't exist, skills proceed without enhanced context. The AI falls back to reading code directly.
