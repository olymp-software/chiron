# Spec Delta

## MODIFIED Requirements

### Requirement: Scripts must not write to the course
The course clone, or the course copy of a local-folder course, SHALL be owned by the app and overwritten on version switches and on Sync. Scripts SHALL keep anything that must persist between runs in `$WORKSPACE_DIR`. Authors SHALL account for implicit writes (e.g. Python `__pycache__`, `go build` output); `cd "$WORKSPACE_DIR"` first is the usual remedy.

#### Scenario: State written to the course folder
- **WHEN** a script writes a file into the activity folder and the learner switches course version
- **THEN** the file is gone

#### Scenario: State written to a local-folder course
- **WHEN** a script of a local-folder course writes a file into the activity folder and the learner presses Sync
- **THEN** the file is gone from the course copy and the author's folder is unchanged
