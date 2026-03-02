## ADDED Requirements

### Requirement: Fix skill reads comment file
The `/gsf:fix` skill SHALL read `review-comments.json` from the project root and present comments to the AI for execution.

#### Scenario: Read and display comments
- **WHEN** user invokes `/gsf:fix`
- **THEN** the skill SHALL read `review-comments.json`, display all comments grouped by file, and begin processing them

#### Scenario: Missing comment file
- **WHEN** user invokes `/gsf:fix` but `review-comments.json` does not exist
- **THEN** the skill SHALL inform the user that no comment file was found and suggest running a review first

### Requirement: Fix skill executes modifications
The `/gsf:fix` skill SHALL process each comment by reading the target file, locating the commented line, and making the requested modification.

#### Scenario: Process comment with clear intent
- **WHEN** a comment says "把这个函数改成返回 error"
- **THEN** the AI SHALL modify the function at the specified file and line to return an error

#### Scenario: Process comment with ambiguous intent
- **WHEN** a comment's intent is unclear
- **THEN** the AI SHALL ask the user for clarification before making changes

#### Scenario: Line number mismatch fallback
- **WHEN** the code at the specified line number does not match `codeContext`
- **THEN** the AI SHALL use `codeContext` to search for the correct location in the file

### Requirement: Fix skill progress tracking
The `/gsf:fix` skill SHALL track and display progress as it processes comments.

#### Scenario: Show progress
- **WHEN** processing comments
- **THEN** the skill SHALL display which comment is being processed (e.g., "Processing 3/7: builder.go:128")

#### Scenario: Completion summary
- **WHEN** all comments have been processed
- **THEN** the skill SHALL display a summary of changes made and suggest cleanup of the comment file
