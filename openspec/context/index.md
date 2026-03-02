# Context Sources

Pluggable context sources provide codebase and environment awareness to OpenSpec skills. Each source is a markdown file in this directory describing what information it provides and how to access it.

## How Skills Use Context Sources

Skills that benefit from codebase awareness (decompose, propose, apply) include a directive:

> Check `openspec/context/` for available context sources and use them as appropriate for the current task.

When an AI agent encounters this directive, it:
1. Reads this `index.md` to discover available sources
2. Reads relevant source files for details
3. Uses the described tools/files as appropriate
4. Falls back to direct code reading if no sources are available

## Available Sources

| Source | File | Best For |
|--------|------|----------|
| Go Codebase (gsf) | [codebase-go.md](codebase-go.md) | Project structure, call chains, route discovery |
| Service Registry | [service-registry.md](service-registry.md) | External RPC service interfaces and context |

## Adding a Context Source

Create a new markdown file in this directory with:
1. **What it provides** — What kind of information
2. **When to use it** — Which workflow stages benefit
3. **How to invoke** — Commands to run or files to read
4. Update the table above
