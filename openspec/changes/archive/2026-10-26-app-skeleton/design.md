# Design

## Context

No code exists yet. The stack and the embedded components are fixed by the baseline design (`openspec/changes/archive/2026-10-25-baseline-specs/design.md`): Go + Wails v2 + Vue 3, `go-git`, `mvdan.cc/sh`, Markdown with HTML disabled, Shiki, `vue-i18n`, data in `%LOCALAPPDATA%\Chiron`. This change replaces the baseline's `moreinterp/coreutils` with the app's own base utilities (see `internal/runtime` below). See proposal.md for the scope of this change.

Constraints that shape the skeleton:
- Wails v2 on Windows needs no cgo; on Linux it needs cgo with WebKitGTK and on macOS cgo with system frameworks. Only Windows exes are built in v1.
- The core must stay buildable and testable on Linux and macOS.

## Goals / Non-Goals

**Goals:**
- `wails build` produces working windows/amd64 and windows/arm64 exes from one Windows runner.
- Every embedded dependency has a real consumer, a unit test, and is exercised in the packaged exe through the smoke screen.
- Code that is not Wails wiring is testable with plain `go test` on all three OSes.

**Non-Goals:**
- Final APIs of `internal/source` and `internal/runtime`. They are deliberately minimal; feature changes reshape them.
- Streaming output, Stop, process-tree kill, `CREATE_NO_WINDOW`, run environment variables: these are `script-runtime` work.
- Settings storage of any kind.

## Decisions

### Project layout: thin root package, code in `internal/`
The Wails `vue-ts` template is the starting point (`main.go`, `app.go`, `wails.json`, `build/`, `frontend/`). The root package contains only Wails wiring: window options, embedding `frontend/dist`, and the bound `App` struct whose methods delegate to `internal/*`. Wails types never leave the root package.
- *Alternative: `cmd/chiron/` + `internal/`:* the usual Go layout, but Wails CLI expects `wails.json` and `main.go` at the project root; fighting it buys nothing.
- *Alternative: logic in bound methods:* faster to write, but needs Wails (and cgo on Linux/macOS) to test.

### `internal/` packages are created on first real code
This change creates `platform`, `source` and `runtime`. The target layout below is guidance for later changes, not empty packages. One package per capability, so a feature change mostly touches its own package:

```
internal/
  platform/    data folder, log, update notice        <- app-platform
  source/      course source, clone, course versions  <- course-source
  manifest/    course.yaml, activity.yaml parsing     <- course-manifest
  content/     activity files, Markdown preparation   <- activity-content
  quiz/        answer evaluation                      <- quiz
  runtime/     interpreter, process tree, output      <- script-runtime
  envcheck/    environment check and its hash         <- environment-check
  workspace/   learner folders, Reset, Full reset     <- learner-workspace
  progress/    state/*.json, atomic writes            <- progress
```

Platform-specific code goes into `_windows.go` / `_darwin.go` / `_linux.go` (or `_unix.go`) files, not `runtime.GOOS` branches, so a missing implementation fails the build on that OS.
- *Alternative: create all packages now with `doc.go`:* the tree looks complete, but empty packages invite guessing at boundaries before the code that defines them exists.

### `internal/platform`: data folder, log folder and log
Two separate locations, each following its OS convention, resolved in one build-tagged file per OS and created on startup if missing:

```
OS       DataDir()                              LogDir()
-------  -------------------------------------  -------------------------------------
Windows  %LOCALAPPDATA%\Chiron                  %LOCALAPPDATA%\Chiron\logs
macOS    ~/Library/Application Support/Chiron   ~/Library/Logs/Chiron
Linux    $XDG_DATA_HOME/chiron                  $XDG_STATE_HOME/chiron
         (default ~/.local/share/chiron)        (default ~/.local/state/chiron)
```

- Linux: the XDG Base Directory spec names logs as state data, so they go to `XDG_STATE_HOME`, not `XDG_DATA_HOME`. Both are resolved by hand; an unset or relative value falls back to the default, as the spec requires.
- macOS: `~/Library/Logs` is where Console.app looks for app logs.
- Windows has no per-user log folder; a `logs` subfolder of the data folder is the common convention.
- Log: `log/slog` text handler appending to `<LogDir>/chiron.log`. The future "Open log folder" button opens `LogDir()`. On startup, a log larger than 10 MB is renamed to `chiron.log.1` (replacing the old one); there is no other rotation.
- *Alternative: logs always inside the data folder:* one place to find everything, but against Linux and macOS conventions and invisible to their log tools.
- *Alternative: Windows Event Log:* needs an event source registered with admin rights; meant for services, not a portable exe.
- Wails' own logger (`options.App.Logger`) is adapted to the same `slog` logger, so webview and runtime errors land in one file.
- A panic in `main` is recovered, logged with its stack and re-raised.
- *Alternative: a logging library (zap, zerolog) or lumberjack rotation:* not needed for a low-volume log; script output never goes there.

### `internal/source`: minimal `go-git` consumer
One function: clone a repository URL into a given folder (no checkout) and return its tag names. Tests build a fixture repository with `go-git` itself in a temp folder, so tests need neither network nor system git.

### `internal/runtime`: minimal `mvdan.cc/sh` consumer with the app's own utilities
One function: run a script source in a given working folder and return stdout, stderr and the exit code. Shell builtins (`echo`, `printf`, `cd`, `test`/`[`, `export`, `exit`, ...) come from `mvdan.cc/sh` itself.

The guaranteed utilities `cat cp mkdir mv rm touch` are implemented in Go in `internal/runtime` and installed as an `interp.ExecHandlers` middleware on **every** OS, so they behave the same everywhere by construction; any other command falls through to the system. Rules:
- Flags are a fixed whitelist: `mkdir -p`, `cp -r`, `rm -r -f`; combined short flags (`-rf`) and `--` are accepted. An unknown flag prints a message to stderr and exits with code 1 instead of behaving subtly differently.
- Paths resolve against the script's current directory; globs are expanded by the interpreter before the utility runs.
- `cp -r src dst` copies into `dst` when it is an existing directory, as GNU `cp` does.
- `mv` falls back to copy-and-remove when a rename fails (e.g. across Windows volumes).
- `rm -f` clears the read-only attribute before removing (e.g. `.git` objects on Windows).
- `touch` creates a missing file and updates the times of an existing one.

`chmod`, `find`, `ls` and `xargs` are no longer guaranteed. Tests run the same scripts on all three OSes.
- *Alternative: `moreinterp/coreutils` (baseline):* u-root's `ls` on Windows prints full paths unless `-p` is given, while GNU `-p` means "append `/` to directories"; the module has no tagged releases. Each such difference makes one script behave differently per OS.
- *Alternative: native scripts per OS (PowerShell on Windows, sh elsewhere):* every script written twice; Windows PowerShell 5.1 reads BOM-less `.ps1` as ANSI and is subject to execution policy.
- *Alternative: patch `ls -p` on Windows:* fixes one symptom, not the class of problems.

### Frontend
- Vue 3 + TypeScript + Vite from the template; npm; Vitest; ESLint (flat config with `eslint-plugin-vue` and `typescript-eslint`) + Prettier.
- i18n: `vue-i18n` with `en` and `ru` message files. The initial locale is `ru` when `navigator.language` starts with `ru`, otherwise `en`. The choice is not persisted.
- Markdown: `markdown-it` with `html: false`. Shiki (`createHighlighterCore` with the Oniguruma engine) with two themes (light and dark via Shiki's dual-theme CSS variables). Before rendering, fence languages are collected and loaded on demand from Shiki's bundled grammars (Vite emits them as lazy chunks embedded in the exe); unknown languages render as plain text.
- *Alternative: `marked`:* comparable, but `markdown-it`'s token stream makes collecting fence languages before rendering straightforward.
- *Alternative: Shiki JS regex engine:* no WASM, but not all grammars are supported; Oniguruma is exact.

### Temporary smoke screen
The only view in the app is a smoke screen with three panels:
- Script: a textarea with a default script using guaranteed commands plus one external command (`hostname`), a Run button, and stdout/stderr/exit code shown after the script ends. The script runs in a fresh temp folder under `<data>/smoke/`.
- Git: a URL field (defaulting to a small public repository) and a Clone button that lists its tags. Clones go to `<data>/smoke/` and are removed before the next clone.
- Markdown: a sample with fences in several languages and one unknown language, plus a ru/en toggle for the UI strings.

Bound methods for it are named `Smoke*` and the view lives in `frontend/src/smoke/`, so removing it is one folder and a few methods. The external command shows whether a console window flashes; fixing it belongs to `script-runtime`.
- *Alternative: tests only:* cannot catch issues specific to the packaged exe (embedded assets, WASM in WebView2, the interpreter and child processes inside a GUI-subsystem exe, arm64).

### CI (GitHub Actions)
One workflow, `.github/workflows/ci-pr.yml`, on pull requests:

```
job            runner           steps
build          windows-latest   npm ci; wails build -platform windows/<arch> -webview2 download
                                -o chiron-windows-<arch>.exe for amd64 and arm64;
                                golangci-lint ./...; upload artifacts
go-test        windows, ubuntu, go test ./internal/...; golangci-lint ./internal/...
               macos (matrix)
frontend       ubuntu-latest    npm ci; lint; prettier --check; vue-tsc; vitest run
```

The root package is linted only in `build` (after the frontend is built, since `//go:embed all:frontend/dist` needs the folder). `go-test` covers `./internal/...` so Linux and macOS need no WebKitGTK or cgo setup, and build-tagged files are linted on their own OS. Tool versions (Go, Node, Wails CLI, golangci-lint) are pinned in the workflow; Go is also pinned by `go.mod`. `.gitattributes` forces LF line endings on checkout, since Windows runners use `core.autocrlf=true` and `gofmt` rejects CRLF.
- *Alternative: lint root package everywhere:* needs Wails' cgo dependencies on Linux runners for no benefit while only Windows is shipped.

### App name in project context
`openspec/config.yaml` states the app is named Chiron instead of asking for `<App>`.

## Risks / Trade-offs

- [The app's utilities diverge from GNU in edge cases] → strict flag whitelist, an error on any unknown flag, the same tests on three OSes.
- [The interpreter behaves differently inside a GUI-subsystem exe than in tests] → the smoke screen runs it in the real exe on both architectures.
- [Child processes flash a console window from the GUI exe] → visible on the smoke screen; fixed in the `script-runtime` change, not here.
- [Shiki grammars or the Oniguruma WASM fail to load from the Wails asset server] → verified on the smoke screen; fallback is the Shiki JS engine.
- [`navigator.language` in WebView2 may not match the Windows display language] → acceptable for the skeleton; a Go-side detection can replace it when settings arrive.
- [arm64 exe cannot run on CI runners] → arm64 is checked manually on an arm64 machine or VM before the change is considered done.
- [Smoke screen code lingers] → isolated in `frontend/src/smoke/` and `Smoke*` methods; the first UI feature change removes it.
