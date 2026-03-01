## 1. Unit Tests

- [x] 1.1 Add tests for Thrift lexer/parser (`internal/thrift/parser_test.go`)
- [x] 1.2 Add tests for AST analysis (`internal/ast/parser_test.go`) - project loading, route discovery, call chain tracing
- [x] 1.3 Add tests for registry generation (`internal/registry/generator_test.go`)

## 2. Self-Bootstrap Verification

- [x] 2.1 Run gsf analyze/routes/trace on itself, verify it works on a real Go project (not just testdata stubs)
- [x] 2.2 Run gsf review on its own commits

## 3. Polish

- [x] 3.1 Add missing error handling and edge case coverage
- [x] 3.2 Ensure all commands have consistent help text and examples
