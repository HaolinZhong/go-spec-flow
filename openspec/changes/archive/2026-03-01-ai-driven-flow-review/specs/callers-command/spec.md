## ADDED Requirements

### Requirement: Find direct callers of a function
The system SHALL find all direct callers (one level) of a specified function within the project, using AST analysis to build a reverse call index.

#### Scenario: Find callers of a function with known callers
- **WHEN** `gsf callers --pkg <package-path> --func <function-name> [dir]` is executed for a function that is called by other functions
- **THEN** the system outputs a list of callers, each with package, function name, file path, and line number of the call site

#### Scenario: Function with no callers
- **WHEN** `gsf callers` is executed for a function that has no callers in the project (e.g., a main function or unused function)
- **THEN** the system outputs an empty callers list with no error

#### Scenario: Method callers
- **WHEN** `gsf callers --pkg X --func "TypeName.MethodName"` is executed for a method
- **THEN** the system finds callers that invoke `receiver.MethodName()` where receiver type matches

#### Scenario: Callers across packages
- **WHEN** a function in package A is called by functions in packages B and C
- **THEN** the system finds callers from all packages in the project, not just the same package

### Requirement: Build reverse call index via AST
The system SHALL traverse all function bodies in the project AST, recording call expressions to build a mapping from called function to list of callers.

#### Scenario: Identify function call targets
- **WHEN** a function body contains `review.MapDiffToFunctions(diffs, pkgs)`
- **THEN** the system records the containing function as a caller of `review.MapDiffToFunctions`

#### Scenario: Identify method call targets
- **WHEN** a function body contains `tracer.Trace(pkg, fn)`
- **THEN** the system records the containing function as a caller of `Tracer.Trace` (using the concrete type name)

### Requirement: Support multiple output formats
The system SHALL support text, JSON, and YAML output formats for the callers command.

#### Scenario: Default text output
- **WHEN** `gsf callers` is run without `--format` flag
- **THEN** the output is human-readable text showing the target function and its callers

#### Scenario: JSON output for AI consumption
- **WHEN** `gsf callers --format json` is run
- **THEN** the output is valid JSON containing `target` and `callers` fields
