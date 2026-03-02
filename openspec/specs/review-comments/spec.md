## ADDED Requirements

### Requirement: Line-level comment UI
The review HTML SHALL allow users to add free-text comments on individual code lines by clicking the line number area.

#### Scenario: Add comment on source line
- **WHEN** user clicks a line number in source view
- **THEN** an inline text input SHALL appear below that line, allowing the user to type a comment

#### Scenario: Add comment on diff added line
- **WHEN** user clicks a line number on a `+` (added) line in diff view
- **THEN** an inline text input SHALL appear below that line

#### Scenario: Deleted lines not commentable
- **WHEN** user views a `-` (deleted) line in diff view
- **THEN** the line number area SHALL NOT be clickable for comments

#### Scenario: Comment indicator on commented lines
- **WHEN** a line has an associated comment
- **THEN** the line number area SHALL display a visual indicator (highlight or icon)

### Requirement: Comment editing and deletion
The review HTML SHALL allow users to edit or delete existing comments.

#### Scenario: Edit existing comment
- **WHEN** user clicks on a line that already has a comment
- **THEN** the comment text SHALL be shown in an editable input, allowing modification

#### Scenario: Delete comment
- **WHEN** user clears the comment text and confirms
- **THEN** the comment SHALL be removed and the visual indicator SHALL disappear

### Requirement: Comment counter display
The review HTML SHALL display a count of total comments in the header area.

#### Scenario: Comment count updates
- **WHEN** user adds, edits, or deletes comments
- **THEN** the header SHALL display the updated total comment count

### Requirement: Local HTTP server mode
The `gsf review` command SHALL support a `--serve` flag that starts a local HTTP server instead of generating a static HTML file.

#### Scenario: Start server with serve flag
- **WHEN** user runs `gsf review --render flow.json --serve`
- **THEN** a local HTTP server SHALL start on a random available port and automatically open the browser to the server URL

#### Scenario: Server serves review HTML
- **WHEN** browser requests `GET /` from the server
- **THEN** the server SHALL respond with the rendered review HTML (same content as static file)

#### Scenario: Server saves comments
- **WHEN** browser sends `POST /comments` with JSON comment data
- **THEN** the server SHALL write the data to `review-comments.json` in the project root directory

#### Scenario: Server shutdown
- **WHEN** user presses Ctrl+C in the terminal
- **THEN** the server SHALL shut down gracefully

### Requirement: Comment auto-save
The review HTML SHALL automatically save comments to the server on every change when running in serve mode.

#### Scenario: Auto-save on comment add
- **WHEN** user adds a new comment while in serve mode
- **THEN** the full comment set SHALL be POSTed to `POST /comments` automatically

#### Scenario: Auto-save on comment edit
- **WHEN** user modifies an existing comment while in serve mode
- **THEN** the updated comment set SHALL be POSTed automatically

#### Scenario: Fallback when not in serve mode
- **WHEN** review HTML is opened as a static file (not via server)
- **THEN** comments SHALL still work in-memory, and a download button SHALL be available to export comments manually

### Requirement: Comment file format
The exported comment file SHALL use the following JSON structure.

#### Scenario: Comment file structure
- **WHEN** comments are saved (via server or manual export)
- **THEN** the file SHALL contain: `reviewTitle` (string), `mode` (string), `exportedAt` (ISO timestamp), and `comments` array where each entry has `file` (string, relative path), `line` (integer, absolute line number in current file), `codeContext` (string, code content of that line), and `comment` (string, user's free text)

### Requirement: Comment association to absolute line numbers
Comments SHALL be associated with the absolute line number in the original source file.

#### Scenario: Source mode line mapping
- **WHEN** user comments on line N in source view
- **THEN** the comment SHALL record the file's absolute line number (lineStart + offset)

#### Scenario: Diff mode line mapping for added lines
- **WHEN** user comments on a `+` line in diff view showing new-side line number M
- **THEN** the comment SHALL record line number M as the absolute line in the current file
