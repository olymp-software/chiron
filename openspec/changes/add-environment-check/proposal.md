# Proposal

## Why

A course can declare `check` in `course.yaml`, but the app ignores it: Enroll and Continue are available whenever `course.yaml` is valid, so a learner can enter a course whose tools are missing and only find out from a failing control. The script runner is also only a placeholder (output after exit, no Stop, no `HOST_OS`/`HOST_ARCH`), and the environment check is its first real consumer, so the runner that controls will reuse is built here too.

## What Changes

- A script runner in `internal/runtime` that implements the `script-runtime` contract for one script: the embedded interpreter with the guaranteed utilities, the run environment, stdout and stderr streamed together in batches, empty stdin, no timeout, and Stop that kills the script with all its child processes at once (no console windows on Windows). Exclusion groups, `ACTIVITY_DIR`/`WORKSPACE_DIR` and controls are not in this change.
- An output panel component that renders streamed output as the `script-runtime` spec defines (UTF-8 with replacement, ANSI SGR styles, `\r`, other control sequences stripped, ~10,000 lines kept). The check is its first user; controls reuse it later.
- The environment check in the course overview: if `check` is declared, the overview shows a "Check environment" button and the check status. Pressing it opens a modal with the output panel and a Cancel button; when the script exits, the modal shows the exit code with a success or failure mark and Cancel becomes OK, which closes the modal. The check can be rerun at any time.
- The result is remembered as the SHA-256 of the `check` file in the course's entry in `library.json`. Exit code 0 stores the hash taken when the run started. A non-zero exit code clears the stored hash only if it equals the hash taken when the run started; otherwise nothing changes. Cancel and closing the app change nothing.
- The check status shown in the overview is derived from the stored hash and the current `check` file: not checked, passed, or check changed. Enroll and Continue in the overview and entering a course (dropdown, app start) require the hashes to match; Go enforces it, not only the UI.
- Sync stops a running check of the course before replacing the course copy.
- **BREAKING (spec)**: the remembered check result is part of the library entry and is deleted with it; "Keep progress" on removal no longer keeps it, because the learner may have changed the machine between removing and re-adding the course.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `environment-check`: the output panel moves from the course overview into a modal with Cancel/OK; the rules for storing and clearing the hash on exit code 0, non-zero and Cancel; the result is stored in the library entry.
- `course-library`: the course overview shows the "Check environment" button and the check status, not an output panel.
- `course-source`: removing a course always deletes the remembered check result; "Keep progress" keeps only the enrollment and progress.

## Impact

- Code: `internal/runtime` gets a streaming runner with process-tree termination (`_windows.go` job objects, `_unix.go` process groups); `internal/library` gets the check hash in `Entry`, check status in `Overview`, starting and stopping a check, and the check gate in `Enroll` and `Enter`; `app.go` binds starting and cancelling a check and emits output and exit events; `main.go` stops running scripts on shutdown.
- `internal/apperr`: new codes for a required check and a check that cannot be started, with `en.json`/`ru.json` messages.
- Frontend: a reusable output panel component and its ANSI/`\r` renderer, a check modal, check button and status in `CourseOverview`, Enroll/Continue gated on the status; i18n strings in `en.json` and `ru.json`.
- Data folder: `library.json` entries gain an optional `checkHash` field; the format version stays 1 because older entries without it read as not checked.
