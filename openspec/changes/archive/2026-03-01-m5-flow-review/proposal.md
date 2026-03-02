## Problem

Traditional file-based diff review doesn't match how developers actually understand changes. When reviewing AI-generated code, you need to follow the request flow (handler → service → dal → RPC) to verify correctness, not jump between random files.

## Proposed Solution

Build a Flow-Based Review that:
1. Parses git diff to identify changed functions/methods
2. Maps changes onto the call chain tree from AST analysis
3. Presents changes organized by request flow, not by file
4. Marks each node with change type (modified/new/unchanged)
5. Shows external RPC context from Service Registry

## Value

- Review follows the natural request flow
- Immediately see which parts of the chain changed
- Spot missing changes (e.g., handler changed but service not updated)
- External RPC context visible at point of use
