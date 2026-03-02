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

## Step 3: AI Orchestration — Build Master Flow

Read `/tmp/gsf-raw.json`. This contains the raw structural tree (packages/files/functions with source code).

Your job: **build a single master flow (主动线) that tells the story of the code's lifecycle, then nest sub-flows under each step**.

### Think like a tour guide:

1. **Understand the code** — Read through every node's code to understand what each function does
2. **Identify the master flow** — What is the end-to-end lifecycle?
   - For codebase mode: "When a command/request comes in, what happens step by step?" (e.g., CLI parses args → loads config → runs analysis → generates output)
   - For diff mode: "What is this change doing, and in what order does it flow?" (e.g., new flag added → new function parses input → existing renderer updated)
3. **Identify sub-flows** — Within each step of the master flow, group the relevant functions
4. **Build a three-layer tree**:
   - **Layer 1 (root)**: Single master flow node — the overall story. **Its `code` field MUST contain the primary entry function's complete code** (e.g., the main handler, the top-level exported function). This is what readers see first when opening the review
   - **Layer 2 (steps)**: Each major step in the lifecycle, in execution order. **Each step's `code` field MUST contain the entry function's complete code** (copy from the original node) — this lets readers see real code at the master flow level
   - **Layer 3 (functions)**: The actual code nodes under each step (including the entry function with its detailed description)
5. **Add descriptions at every level**:
   - `FlowTree.description`: 1-2 sentence overview of the entire review
   - Master flow root's `description`: The complete lifecycle narrative — "When X happens, first A, then B, then C..."
   - Each step's `description`: What this step does, why it matters, how it connects to the previous/next step
   - Each function node's `description`: This function's specific role within its step

### Small package rule:

If there are fewer than 5 functions total, skip layer 2 (steps) and put functions directly under the master flow root (two-layer structure).

### Output format:

Write the orchestrated JSON to `/tmp/gsf-flow.json`. The JSON structure is:

```json
{
  "mode": "codebase",
  "title": "Architecture Review: <project>",
  "description": "Overview of the review...",
  "roots": [
    {
      "id": "master-flow",
      "label": "主动线: Complete Lifecycle Name",
      "description": "When a request arrives, the system first does X, then Y, then Z...",
      "nodeType": "file",
      "code": "func MainEntry(...) { ... }  // ← primary entry function code here",
      "children": [
        {
          "id": "step-1",
          "label": "1. Step Name",
          "description": "This step handles X. It receives Y from the previous step and produces Z for the next step.",
          "nodeType": "file",
          "code": "func EntryFunction(...) { ... }  // ← COPY the entry function's full code here",
          "children": [
            {
              "id": "original-node-id",
              "label": "FunctionName",
              "description": "Role of this function within this step...",
              "package": "...",
              "file": "...",
              "lineStart": 0,
              "lineEnd": 0,
              "code": "... (keep original code) ...",
              "nodeType": "function",
              "children": []
            }
          ]
        },
        {
          "id": "step-2",
          "label": "2. Next Step Name",
          "description": "After step 1 completes, this step...",
          "nodeType": "file",
          "code": "func NextEntryFunc(...) { ... }  // ← entry function code",
          "children": [...]
        }
      ]
    }
  ]
}
```

**Rules:**
- The `roots` array MUST have exactly ONE element — the master flow
- Keep ALL original code in nodes — never truncate or summarize code
- In diff mode, nodes have both `code` (source) and `diff` (raw diff) fields. Keep both — the HTML toggle lets users switch between views
- Each step node's `code` MUST contain its entry function's complete code (copied from the original node). This lets readers see code at the master flow level and click function names to navigate deeper
- Steps (layer 2) should be numbered and in execution/logical order
- **EVERY node at EVERY level MUST have a description** — including deep trace chain nodes (layer 3, 4, 5...). No node should be left without a description. For deep call chain nodes, a brief one-sentence description is fine (e.g., "解析 diff 文本，按文件拆分为独立的 GitDiffFile 列表")
- Step descriptions should explain the connection to previous/next steps
- Descriptions should be in the same language as the user's conversation

### For large codebases:

If the JSON is too large (>500KB), tell the user and suggest reviewing by package:
```
The codebase is large. I recommend reviewing by package:
gsf review --codebase --entry "internal/ast" --json
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
