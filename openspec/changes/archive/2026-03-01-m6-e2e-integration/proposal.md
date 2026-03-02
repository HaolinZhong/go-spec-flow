## Problem

M1-M5 modules are implemented individually but haven't been validated as a complete pipeline. Need end-to-end testing, unit tests, and self-bootstrapping verification.

## Proposed Solution

1. Add unit tests for core modules (AST, Thrift parser, registry, review)
2. Run full pipeline on testdata/sample-app
3. Verify gsf can analyze its own codebase (self-bootstrap)
4. Final polish: error handling, help text, edge cases
