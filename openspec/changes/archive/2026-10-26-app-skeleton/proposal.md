# Proposal

## Why

The baseline specs and design fix the product contract and the stack, but there is no code yet. Every feature change needs a project that builds into a Windows exe, has the shared infrastructure (data folder, log, UI languages) and passes CI. The key embedded dependencies (`go-git`, `mvdan.cc/sh`, Shiki) also carry the biggest technical risk: they must work inside the packaged Windows exe and WebView2, not only in unit tests. This change builds that skeleton and retires that risk before feature work starts.

The app name is now chosen: **Chiron**.

## What Changes

- Wails v2 project with the `vue-ts` template: Go module `github.com/olymp-software/chiron`, frontend embedded into the exe.
- A thin Wails-bound facade in the root package; domain code lives under `internal/`.
- `internal/` packages are created only when their first real code arrives. This change creates three:
  - `internal/platform`: app data folder resolution per OS and the local log file.
  - `internal/source`: a minimal `go-git` consumer (open/clone a repository, list tags).
  - `internal/runtime`: a minimal `mvdan.cc/sh` consumer running a script with the app's own base utilities (`cat cp mkdir mv rm touch`) on every OS. They replace `moreinterp/coreutils` from the baseline design.
- The target layout of the rest of `internal/` is documented in design.md as guidance, not created as empty packages.
- Frontend: `vue-i18n` with Russian and English, defaulting to the system language; Markdown rendering with HTML disabled and Shiki highlighting.
- A temporary smoke screen in the app that runs a script, shows its output, clones a repository and renders highlighted Markdown, so the dependencies are verified in the real exe. It is removed by a later feature change.
- Tooling: npm, Vitest, ESLint + Prettier, golangci-lint.
- GitHub Actions CI: `wails build` for windows/amd64 and windows/arm64 on a Windows runner; `go test` on Windows, Linux and macOS; frontend lint and tests on Linux.
- Replace the "name not chosen" placeholder with Chiron in `openspec/config.yaml`.

Out of scope: any feature from the specs beyond the infrastructure above (course list, manifests, progress, controls, environment check, learner folders), the app update notice, persisting the language choice in settings, a supported-OS field in the course manifest (a separate change), the release workflow and SignPath signing.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `script-runtime`: "Guaranteed commands" narrows the guaranteed utilities to `cat cp mkdir mv rm touch` with a fixed set of flags, provided by the app on every OS instead of only on Windows; `chmod find ls xargs` are no longer guaranteed.
- `environment-check`: the "Working directory" scenario of "Check run environment" no longer relies on `ls`.

The skeleton also implements infrastructure for existing `app-platform` requirements (local log, UI languages) without changing them.

## Impact

- New code: Go module, Wails project files, `frontend/`, `.github/workflows/`.
- Dependencies: Wails v2, `github.com/go-git/go-git/v5`, `mvdan.cc/sh/v3`, Vue 3, `vue-i18n`, Shiki, a Markdown renderer. `mvdan.cc/sh/moreinterp` is not used.
- CI minutes on three OS runners (free for a public repository).
- `openspec/config.yaml` context: app name.
