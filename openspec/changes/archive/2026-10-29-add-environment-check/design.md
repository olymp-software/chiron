# Design

## Context

- `internal/runtime.Run(ctx, script, dir)` parses a script with `mvdan.cc/sh`, runs it with the guaranteed utilities (`utilitiesHandler`) and returns buffered stdout/stderr and the exit code. External commands go to the interpreter's default exec handler: no process-tree kill, no `CREATE_NO_WINDOW`, no `HOST_OS`/`HOST_ARCH`. Nothing outside tests calls it.
- `manifest.ParseCourse` already validates `check` as a relative path inside the course root (`internal/manifest/check.go`).
- `library.Enter` has a marked place for the check gate; `library.Sync` has a marked place for stopping running scripts. `Entry` in `library.json` has `origin`, `source`, `key`, `title`.
- The frontend already uses `<dialog>.showModal()` for the script warning (`library/ScriptWarning.vue`); `router.ts` sends `EnterCourse` errors to the course overview.

See proposal.md for scope and the delta specs for the behavior.

## Goals / Non-Goals

**Goals:**
- One runner API in `internal/runtime` that the check uses now and controls reuse later without changes to its contract.
- The hash rules from `environment-check` enforced in Go; the UI only reflects them.

**Non-Goals:**
- Exclusion groups, `ACTIVITY_DIR`/`WORKSPACE_DIR`, controls, version switching, course removal.
- Interrupting a long-running guaranteed utility (e.g. `cp -r` of a large tree) mid-call; Stop takes effect when it returns.

## Decisions

### Runner API: bytes in, combined output out

```
runtime.Start(Spec{Script []byte, Name, Dir string, Env []string}, out io.Writer) *Run
Run.Stop()            // idempotent, kills everything, returns at once
Run.Wait() Result     // Result{ExitCode int, Stopped bool}
```

- The caller passes the script bytes, not a path: the library hashes exactly the bytes it runs, so a file changed between hashing and running cannot be remembered under the wrong hash.
- stdout and stderr go to one `out` writer behind a mutex, so the panel shows them interleaved in arrival order as the spec requires. The writer gets raw bytes; decoding is the consumer's job (below).
- A parse error is written to `out` like bash does (`check.sh:3:5: ...`) and the run ends with exit code 2. It is a failed check, not an app error.
- `Env` is built by the caller from `os.Environ()`; the runner adds `HOST_OS`/`HOST_ARCH` from `runtime.GOOS`/`GOARCH` and the check removes `ACTIVITY_DIR`/`WORKSPACE_DIR` if inherited. Stdin is an empty reader.
- *Alternative: separate stdout/stderr callbacks.* Rejected: nothing in the specs distinguishes them, and merging them later loses the order.

### Process-tree termination per OS

The runner replaces the interpreter's default exec handler with its own: it resolves the command with `interp.LookPathDir`, starts it with an OS-specific `SysProcAttr` and registers it with the run.

- **Windows (`_windows.go`)**: each run creates a Job Object with `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`. Commands start with `CREATE_NO_WINDOW | CREATE_SUSPENDED`, are assigned to the job, then resumed (`NtResumeProcess`), so no grandchild can be spawned outside the job. Stop calls `TerminateJobObject`; closing the job handle when the app exits or crashes kills the rest.
- **Unix (`_unix.go`)**: commands start with `Setpgid`; Stop sends `SIGKILL` to each registered process group.
- Stop also cancels the interpreter context, so no new command starts. `exec.Cmd.WaitDelay` is set so a killed child that left pipe handles to a stray process cannot block `Wait`.
- *Alternative: `CommandContext` kill only.* Kills just the direct child, which leaves a started server running: fails the "Stop a server" scenario.
- *Alternative: assign to the job after start without suspending.* A fast child can spawn a grandchild before assignment.

### UTF-8 decoding and batching in Go, rendering in the frontend

- A `runtime` decoder holds back an incomplete trailing UTF-8 sequence until the next write and replaces invalid bytes with `U+FFFD` (`strings.ToValidUTF8`), so a multi-byte character split across reads is not broken.
- A batcher flushes decoded text every 75 ms (and at exit) as one Wails event, so endless output costs at most ~13 events per second.
- The frontend renders: `scripts/ansi.ts` parses SGR into styled spans, applies `\r` as return to line start, strips other CSI/OSC sequences; `scripts/OutputPanel.vue` keeps the last 10,000 lines in a panel of fixed height and renders only the lines in view (virtual scrolling): lines have a fixed height and do not wrap, long lines scroll horizontally. Both are generic, for reuse by controls.
- *Alternative: send bytes (base64) and decode with `TextDecoder` in the frontend.* Works, but the decoder is easier to test in Go and the event payload is smaller.

### Check orchestration in `internal/library`

- `Library.StartCheck(key, out io.Writer, done func(runtime.Result)) error`: reads `course.yaml` and the `check` file (via `os.Root` on the course copy), computes the SHA-256, starts the run in the course root and returns. Errors: `courseNotFound`, the course error for an invalid `course.yaml`, `checkReadFailed` if the file cannot be read, `checkRunning` if a check of the course is running. Running checks are kept in a map by storage key under its own mutex; the course copy lock is not held during the run.
- On exit a goroutine applies the result under the library lock (`l.update`) to the entry with that key, if it still exists:
  - exit 0 → `checkHash = startHash`;
  - non-zero and not stopped → if `checkHash == startHash` then clear it;
  - stopped → nothing.
- `Library.StopCheck(key)` stops the run; `Sync` calls it before building the new copy (a running script holds its working directory, which blocks the rename on Windows); `Library.StopAll()` is called from Wails `OnShutdown`.
- `Entry` gains `CheckHash string \`json:"checkHash,omitempty"\``. The hash is stored as lowercase hex.

### Check status and the gate

- `Overview` gains `Check string`: `""` (no `check` or invalid `course.yaml`), `"notChecked"`, `"passed"`, `"changed"`, computed by hashing the current file. An unreadable `check` file is logged and counts as `"notChecked"`, so the gate stays closed.
- `Enroll` and `Enter` call the same `checkSatisfied(entry, course)` and return the new code `checkRequired` when it is false. The UI disables the buttons from `Overview.Check`; the Go check covers races and the dropdown/app start path, which already redirects to the overview on any `EnterCourse` error.
- *Alternative: store a "required" flag.* Not needed: deriving from the hash survives restarts by construction.

### Bound methods and events

- `App.StartCheck(key) StartCheckResult{RunID int, Error}` and `App.CancelCheck(key) ErrorResult`.
- Events: `check:output` `{runId, text}` and `check:exit` `{runId, exitCode, stopped}`. The run id is a process-wide counter; the modal ignores events of other runs, so late output of a previous run cannot leak into a new modal.
- Output can arrive before `StartCheck` returns. The modal subscribes to both events before calling it and buffers events until it knows its run id, then keeps the matching ones and drops the rest.

### Frontend

- `overview/CheckModal.vue`: `<dialog>` opened with `showModal()` (blocks the rest of the window), `@cancel.prevent` so Escape does not close it; Cancel calls `CancelCheck`, which returns once the run has stopped and is no longer registered, then closes the modal and emits `closed` (if `CancelCheck` fails, the modal stays open with the error and OK); after the check exits on its own the modal shows the exit code with a green or red mark, and OK closes the modal and emits `closed`, after which `CourseOverview` reloads the overview to get the new status.
- `CourseOverview.vue`: shows "Check environment" and the status when `overview.check` is set; Enroll/Continue are disabled unless the status is `passed` or there is no check.

### Error codes

New `apperr` codes: `checkRequired`, `checkRunning`, `checkReadFailed` (param `path`), listed in `All`, generated into `codes.ts`, with messages in `en.json` and `ru.json`.

## Risks / Trade-offs

- [`NtResumeProcess` is undocumented] → it is stable since Windows XP and widely used for exactly this; covered by a Windows test that a grandchild started by the script dies on Stop (`task go:test:windows`).
- [A process that calls `setsid` on Unix escapes the process group] → Unix is not a v1 target; accepted.
- [An app older than this change rewrites `library.json` without `checkHash`] → the check is merely required again after a downgrade; no format version bump.
- [Hashing on every overview open reads the `check` file] → the file is a small script; negligible.
- [Output arriving faster than the frontend renders] → bounded by batching and the 10,000-line panel; covered by an endless-output test in the frontend.

## Migration Plan

None: `checkHash` is optional, existing entries read as not checked. `runtime.Run` is replaced; its only callers are its tests.
