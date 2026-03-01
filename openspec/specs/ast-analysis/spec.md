# ast-analysis Specification

## Purpose
TBD - created by archiving change m1-scaffold-and-ast. Update Purpose after archive.
## Requirements
### Requirement: Parse Go project structure
The system SHALL parse a Go project directory and extract package-level structure including all exported and unexported struct definitions, interface definitions, and function/method signatures.

#### Scenario: Parse a multi-package Go project
- **WHEN** `gsf analyze <project-path>` is executed on a Go project with multiple packages
- **THEN** the system outputs a structured listing of all packages, their structs (with field types), interfaces, and function signatures

#### Scenario: Handle parse errors gracefully
- **WHEN** a Go source file contains syntax errors
- **THEN** the system reports the error for that file and continues parsing remaining files

### Requirement: Discover Hertz HTTP routes
The system SHALL parse Hertz router registration code and extract all HTTP route entries, mapping each route (method + path) to its handler function.

#### Scenario: Discover routes from standard Hertz registration
- **WHEN** `gsf routes <project-path>` is executed on a project using `r.Group()`, `r.GET()`, `r.POST()`, `r.PUT()`, `r.DELETE()` style registration
- **THEN** the system outputs a list of all routes with HTTP method, URL path, and the fully qualified handler function name

#### Scenario: Discover routes with nested groups
- **WHEN** a Hertz project uses nested `r.Group("/api").Group("/v1")` with handlers registered at each level
- **THEN** the system correctly composes the full path (e.g., `/api/v1/orders`) for each handler

#### Scenario: No Hertz routes found
- **WHEN** the project does not contain Hertz route registration code
- **THEN** the system outputs an empty route list with no error

### Requirement: Identify Kitex RPC client calls
The system SHALL detect Kitex RPC client method invocations within Go source code and report them as external dependencies.

#### Scenario: Detect Kitex client calls
- **WHEN** a function calls a Kitex-generated client method (e.g., `orderClient.CreateOrder(ctx, req)`)
- **THEN** the system identifies this as an external RPC call and reports the service name and method name

#### Scenario: Distinguish Kitex clients from regular method calls
- **WHEN** a function calls both Kitex client methods and regular struct methods
- **THEN** only Kitex client calls are identified as external RPC dependencies

### Requirement: Trace call chain from entry point
The system SHALL trace the function call chain starting from a specified entry point, following calls across packages within the project.

#### Scenario: Trace from handler to dal
- **WHEN** `gsf trace <project-path> --entry <package>.<Function>` is executed
- **THEN** the system outputs the call chain tree showing all function calls from the entry point through service layer to dal layer

#### Scenario: Mark external RPC calls in trace
- **WHEN** the call chain encounters a Kitex RPC client call
- **THEN** the call is included in the trace tree and marked as `[external-rpc]` with service and method name

#### Scenario: Mark MQ producer calls in trace
- **WHEN** the call chain encounters a message queue producer call (e.g., Kafka, RocketMQ producer)
- **THEN** the call is included in the trace tree and marked as `[mq-producer]` with topic name if detectable

#### Scenario: Handle circular calls
- **WHEN** the call chain contains circular references (A calls B calls A)
- **THEN** the system detects the cycle, marks it as `[cycle]`, and does not recurse infinitely

#### Scenario: Configurable trace depth
- **WHEN** `gsf trace --depth 3` is specified
- **THEN** the system traces only up to 3 levels deep from the entry point

