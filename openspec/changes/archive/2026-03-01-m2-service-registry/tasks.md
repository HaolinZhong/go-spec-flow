## 1. Thrift IDL Parser

- [x] 1.1 Implement Thrift lexer/tokenizer (`internal/thrift/lexer.go`) supporting keywords (service, struct, enum, exception, typedef, include, namespace), identifiers, types (i32, i64, string, bool, list, map, set), and punctuation
- [x] 1.2 Implement Thrift parser (`internal/thrift/parser.go`) that builds an AST with Service, Method, Struct, Enum, Exception, Typedef nodes
- [x] 1.3 Implement include resolution: given an IDL root directory, resolve `include` statements to load referenced files

## 2. Testdata Thrift IDL

- [x] 2.1 Create `testdata/idl/order.thrift` with service definition (CreateOrder, GetOrder), request/response structs, exception, and enum
- [x] 2.2 Create `testdata/idl/common.thrift` with shared types (BaseResponse, Pagination) to test include resolution

## 3. Registry Data Model and Generation

- [x] 3.1 Define registry data model (`internal/registry/model.go`): ServiceRegistry, ServiceInfo, MethodInfo, FieldInfo, TypeInfo
- [x] 3.2 Implement auto.yaml generator (`internal/registry/generator.go`): convert parsed Thrift AST to ServiceInfo, marshal to YAML
- [x] 3.3 Implement context.yaml loader (`internal/registry/context.go`): load and merge human-maintained context with auto-generated data

## 4. Registry Commands

- [x] 4.1 Implement `gsf registry update` command (`internal/cmd/registry.go`): parse IDL directory, generate auto.yaml for each service, update registry-index.yaml
- [x] 4.2 Implement `gsf registry show` command: display merged auto + context info for a service
- [x] 4.3 Implement `gsf registry list` command: list all registered services with summary
