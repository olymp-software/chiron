# learner-workspace Specification

## Purpose

Defines the learner folder where the learner does their work, how workspace files get there, how Reset and Full reset restore them, and how the folder is opened in an external editor.

## Requirements

### Requirement: A learner folder per activity
On first open of any activity, regardless of type, the app SHALL create a learner folder for it, outside the course clone. The default root is `~/<App>/`; inside it the course folder mirrors the course source (`github.com/owner/repo/`, or `local/<folder-name>-<short path hash>/` for a local-folder course), with one subfolder per activity slug.

#### Scenario: Repository course
- **WHEN** the learner first opens activity `hello` of course `github.com/owner/repo`
- **THEN** the folder `~/<App>/github.com/owner/repo/hello` exists

#### Scenario: Theory activity
- **WHEN** the learner opens a `theory` activity
- **THEN** it also has a learner folder, available to its controls as `WORKSPACE_DIR`

### Requirement: Only listed workspace files are copied
For a `coding` activity, only the items listed in `workspace` SHALL be copied to the learner folder, preserving their relative paths; a listed folder SHALL be copied whole. Everything else in the activity folder (scripts, hidden tests, solutions, images) SHALL stay in the course and remain available to scripts.

#### Scenario: Hidden tests stay in the course
- **WHEN** `workspace` lists `main.go` and the activity folder also has `tests/edge_test.go`
- **THEN** the learner folder contains `main.go` but not `tests/`

#### Scenario: Nested path
- **WHEN** `workspace` lists `src/app.py`
- **THEN** the learner folder contains `src/app.py`

### Requirement: Hidden files are hidden only from the UI
Files of a `coding` activity that are not in `workspace` SHALL be hidden only from the UI. They remain readable in the course clone on disk; the app provides no protection against the learner reading them.

#### Scenario: Learner opens the clone
- **WHEN** the learner browses the course clone in a file manager
- **THEN** hidden tests and solutions are readable

### Requirement: Missing workspace files are copied on open
Every time a `coding` activity opens, the app SHALL copy the `workspace` items that are missing from the learner folder. Existing files MUST NOT be touched.

#### Scenario: Learner deleted a file
- **WHEN** the learner deleted `main.go` and reopens the activity
- **THEN** `main.go` is copied again from the course

#### Scenario: Learner edited a file
- **WHEN** the learner edited `main.go` and reopens the activity
- **THEN** the edits are kept

### Requirement: Learner may change anything
The learner SHALL be able to change any file in their folder, including tests. A control script that depends on the original content of a file SHALL copy the original next to the code before running.

```sh
# test.sh
cp main_test.go "$WORKSPACE_DIR/main_test.go"   # in case the learner changed it
cp tests/edge_test.go "$WORKSPACE_DIR/edge_test.go"
cd "$WORKSPACE_DIR"
go test ./...
status=$?
rm -f edge_test.go
exit $status
```

#### Scenario: Learner weakened a test
- **WHEN** the learner edited `main_test.go` and runs the check control above
- **THEN** the check runs against the original test

### Requirement: Reset
A `coding` activity SHALL have a built-in Reset button, not declared by the author. Reset SHALL open a confirmation dialog that by default restores the `workspace` files from the course, letting the learner choose which ones. Other files in the learner folder MUST NOT be touched.

#### Scenario: Reset selected files
- **WHEN** the learner resets only `main.go`
- **THEN** `main.go` matches the course version and other files are unchanged

#### Scenario: Getting a changed file after a version switch
- **WHEN** a new course version changed `main.go` and the learner resets it
- **THEN** the learner gets the new `main.go`

### Requirement: Full reset
The Reset dialog SHALL offer a Full reset option that deletes the learner folder entirely, including environment artifacts, and recreates it from the workspace files.

#### Scenario: Full reset
- **WHEN** the learner confirms Full reset and the folder contained build artifacts
- **THEN** the folder contains only the workspace files

### Requirement: Open in external editor
A `coding` activity SHALL have an "Open in external editor" button that opens the learner folder in the editor chosen in app settings: one of the editors found in `PATH` (`code`, `cursor`, `zed`, `subl`) or a custom command with a `{dir}` placeholder, run directly without a shell. Until an editor is chosen, the button SHALL open the folder in the system file manager.

#### Scenario: No editor chosen
- **WHEN** no editor is set and the learner presses the button
- **THEN** the folder opens in the file manager

#### Scenario: Custom command
- **WHEN** the custom command is `idea64 {dir}`
- **THEN** the app runs `idea64` with the learner folder path as its argument

### Requirement: Relocating learner folders
The learner SHALL be able to change the learner folder root in settings. After confirmation the app SHALL stop running scripts and move the contents to the new location. Course folders that already exist at the new location MUST NOT be overwritten; the app SHALL report the conflict. Courses SHALL be unavailable until the move finishes.

#### Scenario: Conflict at destination
- **WHEN** the new root already has `github.com/owner/repo/`
- **THEN** that course folder is not overwritten and the app reports a conflict

### Requirement: Version switch keeps learner folders
Switching course version SHALL NOT modify learner folders. A changed workspace file SHALL reach the learner only through Reset; a newly added workspace item SHALL appear on the next activity open.

#### Scenario: New workspace item
- **WHEN** a new course version adds `README.txt` to `workspace`
- **THEN** it appears in the learner folder the next time the activity opens
