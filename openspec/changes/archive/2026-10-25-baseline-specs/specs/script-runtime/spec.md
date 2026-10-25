# Spec Delta

## Purpose

Defines the contract between the app and course scripts: which interpreter runs them, which commands they can rely on, the environment they run in, how output and exit codes are shown, and how scripts are stopped and serialized.

## ADDED Requirements

### Requirement: Embedded interpreter
Control scripts and the environment check SHALL be run by a shell interpreter embedded in the app that behaves like non-interactive bash. A system bash MUST NOT be required on any OS, including Windows. The same script SHALL behave the same on Linux, macOS and Windows.

#### Scenario: Windows without bash
- **WHEN** a learner on Windows without Git Bash or WSL presses a control
- **THEN** its script runs

### Requirement: No sandbox
Course scripts SHALL run with the learner's permissions and without a sandbox, like running `make` in someone else's repository. The app provides no protection from a malicious author; the learner accepts this once when adding the course.

#### Scenario: Script touches files outside the course
- **WHEN** a script writes to a file in the learner's home folder
- **THEN** the write succeeds

### Requirement: Guaranteed commands
On every OS a script SHALL be able to rely on bash builtins (`echo`, `printf`, `cd`, `test`/`[`, `export`, `exit`, etc.) and the utilities `cat chmod cp find ls mkdir mv rm touch xargs`. On Windows these utilities SHALL be provided by the app. Anything else (`grep`, `sed`, `awk`, `head`, `diff`, tools like `go`, `docker`, `kubectl`) is not guaranteed and SHALL be verified by the course's environment check.

#### Scenario: Guaranteed utility on Windows
- **WHEN** a script calls `cp a b` on Windows
- **THEN** the file is copied

#### Scenario: Non-guaranteed utility on Windows
- **WHEN** a script calls `grep` on Windows without grep installed
- **THEN** the command fails as not found

### Requirement: Run environment
A control script SHALL run with the activity folder in the course as working directory, no arguments, empty stdin, and stdout/stderr as pipes rather than a terminal. It SHALL receive `ACTIVITY_DIR` (activity folder in the course), `WORKSPACE_DIR` (the learner folder of the activity), `HOST_OS` (`linux`, `darwin`, `windows`) and `HOST_ARCH` (`amd64`, `arm64`), plus the inherited environment (`PATH`, `HOME`, ...).

#### Scenario: Hidden test fixture
- **WHEN** a script reads `tests/edge_test.go` by relative path
- **THEN** it reads the file from the activity folder in the course

#### Scenario: Architecture-specific download
- **WHEN** a script on Windows ARM reads `HOST_OS` and `HOST_ARCH`
- **THEN** they are `windows` and `arm64`

#### Scenario: Reading stdin
- **WHEN** a script reads from stdin
- **THEN** it immediately gets end of input

### Requirement: Scripts must not write to the course
The course clone SHALL be owned by the app and overwritten on version switches. Scripts SHALL keep anything that must persist between runs in `$WORKSPACE_DIR`. Authors SHALL account for implicit writes (e.g. Python `__pycache__`, `go build` output); `cd "$WORKSPACE_DIR"` first is the usual remedy.

#### Scenario: State written to the course folder
- **WHEN** a script writes a file into the activity folder and the learner switches course version
- **THEN** the file is gone

### Requirement: Output display
stdout and stderr SHALL appear in the output panel as they arrive. Output SHALL be decoded as UTF-8 with invalid bytes replaced by `�`. ANSI SGR sequences SHALL be rendered as styles, `\r` SHALL return to the start of the current line, and other control sequences SHALL be stripped.

```sh
# typical run.sh preamble for a Python course
export PYTHONUNBUFFERED=1 PYTHONUTF8=1 PYTHONDONTWRITEBYTECODE=1
cd "$WORKSPACE_DIR"
python3 main.py
```

#### Scenario: Progress bar
- **WHEN** a script prints `10%\r20%\r30%`
- **THEN** the panel shows a single line ending in `30%`

#### Scenario: Colored output
- **WHEN** a script prints text with SGR color codes
- **THEN** the text is shown in color without escape characters

#### Scenario: Invalid bytes
- **WHEN** a script prints bytes that are not valid UTF-8
- **THEN** they are shown as the replacement character

### Requirement: No color forcing
The app MUST NOT set `NO_COLOR` or enable colors. Through a pipe most tools write without color; enabling it is the author's choice (`--color=always`, `FORCE_COLOR=1`).

#### Scenario: Environment inspection
- **WHEN** a script prints `$NO_COLOR` and the learner has not set it
- **THEN** the value is empty

### Requirement: Bounded output panel
The output panel SHALL keep only the latest ~10,000 lines and drop older ones silently. Output SHALL be delivered to the UI in batches every 50–100 ms so that unbounded output does not freeze the window.

#### Scenario: Infinite loop printing
- **WHEN** a script prints lines in an endless loop
- **THEN** the window stays responsive and the panel holds the latest lines

### Requirement: Exit code
Exit code `0` SHALL be shown as success (green mark) and any other code as failure (red mark). A control's result MUST NOT affect progress. There is no structured result (check list) in v1.

#### Scenario: Failing tests
- **WHEN** a control script exits with code 1
- **THEN** the control shows a red mark and progress is unchanged

### Requirement: Immediate stop
There SHALL be no timeout. Every running script SHALL have a "Stop" button that terminates it with all its child processes at once, with no graceful signal. Scripts SHALL also be terminated this way when the learner leaves the activity, switches course version, or closes the app. Child processes SHALL not open console windows.

#### Scenario: Stop a server
- **WHEN** a script started a server process and the learner presses "Stop"
- **THEN** the script and the server process are both terminated

#### Scenario: Leaving the activity
- **WHEN** a control is running and the learner opens another activity
- **THEN** the running script is terminated

### Requirement: Cleanup at start, not at stop
Because scripts get no chance to clean up when stopped, a script SHALL remove leftovers of an interrupted run (containers, temp files, busy ports) at the beginning of its next run.

#### Scenario: Leftover container
- **WHEN** a previous run was stopped while a container was running
- **THEN** the next run's `docker rm -f` before `docker run` lets it start cleanly

### Requirement: Exclusion groups
The same control MUST NOT run twice at once. Controls with the same `concurrency` value SHALL be mutually exclusive: while one runs, the others' buttons are disabled; a new start is blocked and the running one is not interrupted. Controls without `concurrency` SHALL share the activity's default group. Groups SHALL apply only within one activity.

```yaml
controls:
  - label: Start server
    script: server.sh
    concurrency: server   # runs alongside the others
  - label: Prepare
    script: setup.sh      # these two exclude each other
  - label: Check API
    script: check-api.sh
```

#### Scenario: Default group
- **WHEN** an activity has two controls without `concurrency` and one is running
- **THEN** the other's button is disabled

#### Scenario: Separate group
- **WHEN** "Start server" with `concurrency: server` is running
- **THEN** "Check API" can be started
