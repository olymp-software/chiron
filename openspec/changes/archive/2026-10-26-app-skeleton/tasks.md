# Tasks

## 1. Project context

- [x] 1.1 Replace the "final app name is not chosen yet" sentence in `openspec/config.yaml` with "The app is named Chiron"; verify `openspec context --json` succeeds

## 2. Wails project

- [x] 2.1 Generate the project from the Wails v2 `vue-ts` template with module `github.com/olymp-software/chiron` and name `chiron`; set window title "Chiron" in `main.go` and product info in `wails.json`; verify `wails build -platform windows/amd64` produces an exe that opens a window
- [x] 2.2 Pin tool versions: Go in `go.mod` (`go` + `toolchain`), Node in `.nvmrc`, Wails in `go.mod`; extend `.gitignore` with `build/bin/`, `frontend/dist/`, `frontend/node_modules/`; verify `git status` shows no generated files after a build
- [x] 2.3 Reduce the root package to Wails wiring: `main.go` (options, `//go:embed all:frontend/dist`) and `app.go` (bound `App` struct, no logic); verify `wails dev` starts and the template's demo binding is removed

## 3. Frontend tooling

- [x] 3.1 Add ESLint (flat config, `eslint-plugin-vue`, `typescript-eslint`), Prettier, Vitest and `vue-tsc` with npm scripts `lint`, `format:check`, `typecheck`, `test`; verify all four pass on the template code
- [x] 3.2 Add `vue-i18n` with `src/i18n/en.json` and `src/i18n/ru.json` and locale selection from `navigator.language` (`ru*` → `ru`, otherwise `en`); add a Vitest test for the selection; verify `npm test` passes

## 4. Go tooling

- [x] 4.1 Add `.golangci.yml` (default linters plus `gofmt`, `goimports`, `errorlint`); verify `golangci-lint run ./...` passes after a frontend build

## 5. `internal/platform`

- [x] 5.1 Implement `DataDir()` and `LogDir()` in `dirs_windows.go`, `dirs_darwin.go`, `dirs_linux.go` per the table in design.md (Linux: `XDG_DATA_HOME` / `XDG_STATE_HOME`, unset or relative values fall back to `~/.local/share/chiron` / `~/.local/state/chiron`); verify with tests run on each OS, including set, unset and relative `XDG_*` cases on Linux
- [x] 5.2 Implement the log: `slog` text handler appending to `<LogDir>/chiron.log`, rename to `chiron.log.1` when over 10 MB at startup; verify with tests for creation, appending and the size rollover
- [x] 5.3 Wire platform into `main.go`: create both folders, open the log, adapt Wails' `logger.Logger` to `slog`, recover and log a panic with its stack; verify that starting the exe creates `%LOCALAPPDATA%\Chiron\logs\chiron.log` with a startup line

## 6. `internal/source`

- [x] 6.1 Add `github.com/go-git/go-git/v5` and implement clone-without-checkout plus tag listing; verify with a test that builds a fixture repository with tags `v1.0.0` and `v1.1.0` via go-git in a temp folder and gets both tags back

## 7. `internal/runtime`

- [x] 7.1 Add `mvdan.cc/sh/v3`; implement running a script in a working folder returning stdout, stderr and exit code, with the app's utilities in the exec handler chain on every OS; verify with tests on all OSes: a script using `mkdir touch cat echo` produces the expected stdout, `echo x >&2` lands in stderr, `exit 3` yields exit code 3, and `go.mod` does not require `mvdan.cc/sh/moreinterp`
- [x] 7.2 Implement the utilities `cat cp mkdir mv rm touch` in `internal/runtime` per design.md (flag whitelist `mkdir -p`, `cp -r`, `rm -r -f`, combined short flags, `--`; paths relative to the script's directory); verify with tests on all OSes: `mkdir -p` creates nested folders, `cp -r` into an existing folder copies inside it, `rm -rf` removes a tree with a read-only file, `touch` creates a new file and updates an existing one's time, `mv` renames, an unknown flag prints a message and exits with code 1, `--` allows a file named `-x`

## 8. Markdown and highlighting

- [x] 8.1 Add `markdown-it` (`html: false`) and Shiki core with the Oniguruma engine and light/dark dual themes; load fence languages on demand before rendering, unknown languages as plain text; verify with Vitest tests: raw HTML is escaped, a `go` fence gets highlighted tokens, an unknown fence renders as plain `<pre><code>`
- [x] 8.2 Verify the production build emits grammars as separate lazy chunks: `npm run build` output contains per-language chunk files and the main bundle does not include them

## 9. Smoke screen

- [x] 9.1 Add bound methods `SmokeRunScript` and `SmokeListTags` in `app.go` delegating to `internal/runtime` and `internal/source`, using fresh folders under `<data>/smoke/` (cleared before each run/clone); verify `wails generate module` produces TS bindings that type-check
- [x] 9.2 Add `frontend/src/smoke/` with the Script, Git and Markdown panels and a ru/en toggle, all strings via `vue-i18n`; make it the app's only view; verify `npm run lint`, `typecheck` and `test` pass

## 10. CI

- [x] 10.1 Add `.github/workflows/ci-pr.yml` running on pull requests, with jobs `build` (windows-latest: wails build amd64 and arm64 with `-webview2 download -o chiron-windows-{amd64,arm64}.exe`, `golangci-lint ./...`, upload artifacts), `go-test` (matrix windows/ubuntu/macos: `go test ./internal/...`, `golangci-lint ./internal/...`), `frontend` (ubuntu: `npm ci`, lint, format check, typecheck, test); pin all tool versions; verify all jobs pass on a pushed branch

## 11. Integration check

- [x] 11.1 Run the CI-built amd64 exe on Windows: the window opens, the Script panel shows output and exit code for the default script, the Git panel lists tags of the default repository, the Markdown panel shows highlighted code in both themes, the ru/en toggle switches strings, and `%LOCALAPPDATA%\Chiron\logs\chiron.log` has entries; record whether `hostname` flashes a console window
  - Result: checked under x64 emulation on Windows 11 arm64, no native x64 machine: the exe starts, the Script and Git panels work.
- [x] 11.2 Run the CI-built arm64 exe on a Windows arm64 machine or VM with the same checks as 11.1
  - Result: all checks pass. Running the default script flashes a console window for `hostname`; fixing it (`CREATE_NO_WINDOW`) belongs to the `script-runtime` change.
