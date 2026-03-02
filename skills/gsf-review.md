---
name: gsf:review
description: Generate interactive HTML code review with flow tree navigation
---

Generate an AI-orchestrated interactive code review. You are a "tour guide" — organize code into meaningful flows and explain each node's role.

## Step 1: Ask Review Scope

Ask the user to choose one:
1. **最近一次 commit** — review 最新提交
2. **最近 N 次 commits** — review 最近几次提交（问 N 是多少）
3. **vs 某个分支** — 和指定分支对比（问分支名，默认 main）
4. **整个 codebase** — 探索完整代码架构
5. **某个包** — 探索特定包（问包路径）

## Step 2: Extract Structural Data

Ensure `go` is in PATH, then run gsf to get raw JSON:

```bash
export PATH=$HOME/sdk/go1.22.5/bin:$PATH
```

| Scope | Command |
|-------|---------|
| 最近一次 commit | `gsf review --commit HEAD --json > /tmp/gsf-raw.json` |
| 最近 N 次 | `gsf review --commit HEAD~N..HEAD --json > /tmp/gsf-raw.json` |
| vs 分支 | `gsf review --base <branch> --json > /tmp/gsf-raw.json` |
| 整个 codebase | `gsf review --codebase --json > /tmp/gsf-raw.json` |
| 某个包 | `gsf review --codebase --entry "<pkg>" --json > /tmp/gsf-raw.json` |
| 全部历史 | `gsf review --commit $(git rev-list --max-parents=0 HEAD)..HEAD --json > /tmp/gsf-raw.json` |

Optional: add `--depth N` to increase call chain trace depth (default: 2 for diff, 4 for codebase). Use higher values (e.g., `--depth 6`) if important functions are cut off by depth limits.

## Step 3: AI Orchestration — Add Descriptions and Organize

Read `/tmp/gsf-raw.json`. The JSON is **pre-organized by call chains**:

- **Chain flows**: Groups of changed functions connected by call relationships, with bridge functions (unchanged functions that connect them) in between
- **Isolated flows**: Changed functions with no call relationship to other changed functions
- **Non-code Files**: Non-Go changed files grouped together

### For diff mode:

The JSON already has the structure. Your job is to **add descriptions and optionally reorganize**:

1. **Read every node's code** to understand what each function does
2. **Add `FlowTree.description`**: 1-2 sentence overview of the entire change
3. **For each flow root**: Add a `description` explaining what this flow does as a whole
4. **For each function node**: Add a `description` explaining this function's role
5. **Bridge nodes** (`isBridge: true`): Add a brief description of how they connect the changed functions
6. **Optionally reorganize**: You may merge related flows, reorder nodes, or create a master flow wrapper if it tells a better story

### For codebase mode:

The JSON is organized by file/package. Your job is heavier — **build a master flow (主动线)**:

1. **Understand the code** — Read through every node's code
2. **Identify the master flow** — "When a command/request comes in, what happens step by step?"
3. **Build a three-layer tree**:
   - **Layer 1 (root)**: Single master flow node — the overall story. Its `code` field MUST contain the primary entry function's complete code
   - **Layer 2 (steps)**: Each major step in the lifecycle, in execution order
   - **Layer 3 (functions)**: The actual code nodes under each step
4. **Add descriptions at every level**

### Small change rule:

If there are fewer than 5 function nodes total, skip reorganization — just add descriptions to existing nodes.

### Output format:

Write the orchestrated JSON to `/tmp/gsf-flow.json`. Keep the same JSON structure:

```json
{
  "mode": "diff",
  "title": "Diff Review: ...",
  "description": "Overview of the review...",
  "roots": [
    {
      "id": "flow-0",
      "label": "主动线: Entry Function Name",
      "description": "This flow covers how X calls Y through Z...",
      "nodeType": "file",
      "code": "func Entry(...) { ... }",
      "children": [
        {
          "id": "flow-0-entry-pkg-Entry",
          "label": "Entry",
          "description": "Entry point that initiates the flow...",
          "package": "...",
          "file": "...",
          "code": "...",
          "diff": "...",
          "nodeType": "function"
        },
        {
          "id": "flow-0-bridge-pkg-Middle",
          "label": "Middle",
          "description": "Bridge: connects Entry to Target by...",
          "isBridge": true,
          "code": "...",
          "nodeType": "function"
        },
        {
          "id": "flow-0-func-pkg-Target",
          "label": "Target",
          "description": "Handles the actual change...",
          "code": "...",
          "diff": "...",
          "nodeType": "function"
        }
      ]
    }
  ]
}
```

**Rules:**
- Keep ALL original code in nodes — never truncate or summarize code
- In diff mode, nodes have both `code` (source) and `diff` (raw diff) fields. Keep both
- **EVERY node MUST have a description** — including bridge nodes. For simple nodes, a brief one-sentence description is fine
- Descriptions should be in the same language as the user's conversation
- Do NOT remove `isBridge` flags from bridge nodes

### For large diffs:

If the JSON is too large (>500KB), tell the user and suggest narrowing the scope:
```
The diff is large. I recommend reviewing fewer commits or a specific package.
```

## Step 4: Render and Open

Use `--serve` to enable interactive commenting (auto-saves to disk), or `--open` for static HTML:

```bash
# With comment support (recommended):
gsf review --render /tmp/gsf-flow.json --serve

# Static HTML (no auto-save, but has manual export):
gsf review --render /tmp/gsf-flow.json --open
```

Tell the user:
- Left panel: flow tree with master flow → steps → functions (click to navigate)
- Right panel: source code with AI commentary above each function
- Blue card above code = AI's explanation of the function's role
- Underlined function names in code are clickable — click to jump to that function's detail in the tree
- In diff mode: Source/Diff toggle in header — switch between syntax-highlighted source and colored diff view (green=add, red=delete). In codebase mode the toggle is disabled
- Click any line number to add a comment — comments are auto-saved to `review-comments.json` (in serve mode) or can be exported via the Export button (in static mode)
- After reviewing, use `/gsf:fix` to have AI read the comments and apply code changes
