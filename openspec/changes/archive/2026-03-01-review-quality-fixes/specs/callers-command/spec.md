## MODIFIED Requirements

### Requirement: Build reverse call index via AST
The system SHALL traverse all function declarations AND package-level variable declarations containing function literals in the project AST, recording call expressions to build a mapping from called function to list of callers.

#### Scenario: Identify function call targets
- **WHEN** a function body contains `review.MapDiffToFunctions(diffs, pkgs)`
- **THEN** the system records the containing function as a caller of `review.MapDiffToFunctions`

#### Scenario: Identify method call targets
- **WHEN** a function body contains `tracer.Trace(pkg, fn)`
- **THEN** the system records the containing function as a caller of `Tracer.Trace` (using the concrete type name)

#### Scenario: Identify calls inside package-level function literals
- **WHEN** a package-level variable declaration contains a function literal that calls a target function (e.g., `var cmd = &cobra.Command{ RunE: func(...) { review.ExtractDiffEntries(...) } }`)
- **THEN** the system records the variable name as the caller context (e.g., caller name includes the variable name)

#### Scenario: Identify calls inside init function
- **WHEN** a package `init()` function calls a target function
- **THEN** the system records `init` as the caller name

## ADDED Requirements

### Requirement: Consistent call target resolution
The system SHALL use a single shared call resolution function for both callers lookup and call chain tracing, ensuring consistent results between `gsf callers` and `gsf trace`.

#### Scenario: Callers and trace agree on call targets
- **WHEN** `gsf trace` identifies function A calling function B
- **THEN** `gsf callers` for function B SHALL include function A as a caller
