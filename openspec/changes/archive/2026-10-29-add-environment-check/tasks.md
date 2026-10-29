# Tasks

## 1. Error codes

- [x] 1.1 Add `checkRequired`, `checkRunning`, `checkReadFailed` (param `path`) to `internal/apperr` and `All`, run `go generate ./internal/apperr`, add `errors.<code>` messages to `en.json` and `ru.json`; verify with `task go:test` and `task frontend:check`

## 2. Script runner

- [x] 2.1 Replace `runtime.Run` with `Start(Spec, out) *Run`, `Run.Wait() Result{ExitCode, Stopped}` and idempotent `Run.Stop()`: script bytes in, stdout and stderr to one writer behind a mutex, empty stdin, `HOST_OS`/`HOST_ARCH` added to `Spec.Env`, parse error written to `out` with exit code 2, interpreter context cancelled on Stop; port the existing tests and add tests for interleaved output, env variables, empty stdin, a parse error and exit codes 0/1/127 via `task go:test`
- [x] 2.2 Add the exec handler that resolves commands with `interp.LookPathDir`, sets `WaitDelay` and registers started processes with the run; Unix (`_unix.go`): `Setpgid` and `SIGKILL` to each group on Stop; verify with a test where the script starts a background child that starts a grandchild and Stop kills both, via `task go:test`
- [x] 2.3 Windows (`_windows.go`): a Job Object per run with `KILL_ON_JOB_CLOSE`, commands started with `CREATE_NO_WINDOW | CREATE_SUSPENDED`, assigned to the job and resumed, `TerminateJobObject` on Stop; verify the same process-tree test with `task go:test:windows` and lint with `task go:lint`
- [x] 2.4 Add the streaming UTF-8 decoder (holds back an incomplete trailing sequence, replaces invalid bytes with `U+FFFD`) and the batcher (flush every 75 ms and on close); verify with tests for a multi-byte character split across writes, invalid bytes, and batch timing with a fake clock via `task go:test`

## 3. Check in the library

- [x] 3.1 Add `CheckHash` (`checkHash,omitempty`) to `Entry` and `Overview.Check` (`""`, `notChecked`, `passed`, `changed`) computed from the SHA-256 of the `check` file read through `os.Root`; an unreadable file is logged and counts as `notChecked`; verify with tests for each status, an invalid `course.yaml`, and an old `library.json` without the field via `task go:test`
- [x] 3.2 Add `StartCheck(key, out, done)`, `StopCheck(key)`, `StopAll()`: read and hash the file, run it in the course root without `ACTIVITY_DIR`/`WORKSPACE_DIR`, refuse a second run with `checkRunning`, apply the result on exit under the library lock (0 stores the start hash; non-zero clears only an equal stored hash; stopped changes nothing; removed entry ignored); verify each rule, `checkRunning`, `checkReadFailed` and the working-directory/no-activity-variables scenarios via `task go:test`
- [x] 3.3 Gate `Enroll` and `Enter` with `checkRequired` when `check` is declared and not satisfied, and stop a running check at the start of `Sync`; verify with tests for Enroll/Enter with no check, a satisfied check, a changed `check` file after Sync, and Sync during a running check via `task go:test`

## 4. Wails wiring

- [x] 4.1 Bind `StartCheck(key) StartCheckResult{RunID, Error}` and `CancelCheck(key) ErrorResult` in `app.go`, emitting `check:output {runId, text}` through the decoder and batcher and `check:exit {runId, exitCode, stopped}`; call `lib.StopAll()` from `OnShutdown` in `main.go`; verify with `task go:lint` and `task build`

## 5. Output panel

- [x] 5.1 Add `frontend/src/scripts/ansi.ts`: SGR to styled spans (colors, bold, reset), `\r` returns to line start, other CSI/OSC sequences stripped; verify with vitest for the progress bar, colored text and stripped cursor sequences via `task frontend:test`
- [x] 5.2 Add `frontend/src/scripts/OutputPanel.vue` that appends text chunks, keeps the last 10,000 lines in a fixed-height panel with virtual scrolling (fixed line height, no wrapping) and follows the bottom; verify with vitest that 20,000 appended lines leave 10,000 and the visible line range is computed correctly via `task frontend:test`

## 6. Check UI

- [x] 6.1 Add `overview/CheckModal.vue`: `<dialog>` with `showModal()` and `@cancel.prevent`, subscribes to `check:output`/`check:exit` before `StartCheck` and keeps only events of its run id, Cancel calls `CancelCheck` and closes the modal once it returns (`StopCheck` waits until the run is unregistered), after exit shows the exit code with a success or failure mark (or stopped) and OK closes it; i18n strings in `en.json` and `ru.json`; verify with vitest for output of another run being dropped, Cancel, and OK after exit via `task frontend:test`
- [x] 6.2 In `CourseOverview.vue` show "Check environment" and the status when `overview.check` is set, open the modal, reload the overview after it closes, and disable Enroll/Continue unless the status is `passed` or there is no check; i18n strings; verify with `task frontend:check`

## 7. Integration

- [x] 7.1 Run `task check` and verify it passes
- [x] 7.2 With `task run`, add a local-folder course whose `check` prints colored lines with `\r` progress, sleeps and starts a child process; verify streaming, Cancel killing the child (no console window), exit 0 enabling Enroll, a failing rerun disabling it, and an edited `check` after Sync showing "check changed"
