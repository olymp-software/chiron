# Design

## Context

Greenfield project, no code yet. See proposal.md for motivation and specs/ for the behavior contract. Constraints that shape every decision below:

- Learners are mostly on Windows, where bash, git and Unix utilities are usually absent. The app must not require installing any of them.
- One course script must behave the same on Linux, macOS and Windows.
- The app knows nothing about course subject matter; all domain logic lives in course scripts.

## Goals / Non-Goals

**Goals:**
- A self-contained Windows exe that can clone courses, render content and run course scripts with no external prerequisites besides WebView2.
- A core that does not close the door on Linux and macOS builds later.

**Non-Goals (v1):**
- Built-in code editor and the `workspace` item fields that come with it (`readonly`, `open`); `terminal` activity type.
- Quiz: shuffling, hints, question ids, kinds other than `choice`.
- Control `timeout`, interactive stdin, structured script results, course-level shared scripts and `COURSE_DIR`.
- Math formulas and diagrams in Markdown; activity type icons in the tree.
- A CLI course linter for CI; slug aliases; a `keep` list for Full reset.
- Installer, winget, self-update; Linux and macOS builds.
- Private repositories, Git LFS.

Each of these can be added later as a new schema version without breaking existing courses.

## Decisions

### Stack: Go + Wails v2 + Vue 3
The core is Go because both key embedded components (shell interpreter, git) are Go libraries. The UI is a webview because it renders Markdown and highlighted code. The frontend is embedded into the exe via `go:embed`. Wails v2 on Windows needs no cgo, so amd64 and arm64 builds come from one machine (`wails build -platform windows/arm64`). WebView2 uses Wails' `-webview2 download` mode.
- *Alternative:* Wails v3 — still beta; revisit after a stable release.

### Embedded shell interpreter `mvdan.cc/sh`
Scripts run in `mvdan.cc/sh/v3/interp`, not a system shell; on Windows the base utilities come from `moreinterp/coreutils`, pinned to a specific commit (it has no tags or stability promise).
- *System bash (Git for Windows):* full bash, but an external install on Windows and varying environments.
- *Arbitrary command per control (`python check.py`):* machine requirements would depend on each author.
- *Per-OS scripts (`.sh` + `.ps1`):* double maintenance, one always lags.
- *WSL or containers:* most predictable, far too heavy for learners.

Consequence: only bash builtins and `cat chmod cp find ls mkdir mv rm touch xargs` are guaranteed; anything else goes through the environment check.

### Embedded `go-git`
Cloning and tag checkout use `go-git`, not system git — same reasoning as the interpreter.
- *System git:* full compatibility, but a separate install with varying versions.
- *Tag archive via GitHub API:* simplest, but ties the app to GitHub while a course source is meant to be any git repository.

Consequence: a version switch is fetch tags + forced checkout of the tag, which matches "the app owns the clone"; the selected version is simply the checked-out tag. Git LFS is unsupported.

### State in per-course JSON files
`library.json` (course list, last opened course) and `state/<course source>.json` per course (current activity, completed slugs, environment check hash, format version) in the app data folder. Writes go to a temp file next to the target and are renamed over it.
- *Single JSON for everything:* a write error affects all courses; whole file rewritten on every change.
- *Event log (JSONL):* history and crash resilience, but unbounded growth and awkward manual edits. History can be added later without touching current state.
- *bbolt / SQLite:* binary file or cgo/size cost; no queries are needed.
- *Next to learner folders:* easy to delete together with a folder or Full reset.
- *Webview storage:* tied to the webview profile, awkward for the Go core.

### Data locations
- App data folder: `%LOCALAPPDATA%\<App>`; future `~/Library/Application Support/<App>` on macOS and `$XDG_DATA_HOME/<app>` (default `~/.local/share/<app>`, written by hand — Go's stdlib has no such helper) on Linux.
- Course clones inside it at paths mirroring the course source; log file inside it.
- Learner folders separately under `~/<App>/` (configurable).

### Stop kills immediately
On Stop (and on leaving the activity, version switch, app close) the whole process tree is killed: a Job Object on Windows, a process group on Linux/macOS. Child processes start with `CREATE_NO_WINDOW`.
- *Graceful where possible (SIGTERM then SIGKILL on Unix, kill on Windows):* behavior would differ per OS; authors debugging on Unix would rely on `trap` and cleanup would silently not happen for Windows learners.
- *Kill only the direct script process:* grandchildren (built binaries, servers) keep ports and files busy.

### Progress moved only by "Next"
Completion is the learner's decision, not a check result.
- *Completion by check (control flag `completes`, correct quiz answers):* needs a success definition per type and a story for activities without checks; learners get stuck on material they already know.
- *Explicit completion rule field:* duplicates what follows from type and controls and can contradict them.

### Output pipeline
stdout/stderr are read from pipes, decoded as UTF-8 with replacement, SGR converted to styles, `\r` handled as line rewrite, other escapes stripped. A ring buffer of ~10,000 lines; batches pushed to the UI every 50–100 ms.

### Markdown and highlighting
Markdown rendered with HTML disabled. Code highlighting via Shiki with all languages and two themes; unknown languages fall back to plain text. All languages add about 8.6 MB to the exe; grammars load lazily in the webview. A future editor on Monaco can reuse the same grammars and themes via `@shikijs/monaco`; CodeMirror would use its own grammars and colors would differ.

### i18n
All UI strings via `vue-i18n` (ru, en).

### Distribution and signing
Release exes are named `<app>-windows-amd64.exe` and `<app>-windows-arm64.exe`, published on GitHub Releases, signed through SignPath Foundation (free for open source). SignPath requires automated builds from a public repository, so releases are built only in CI.

### License
MIT. Dependencies are compatible: `mvdan.cc/sh` BSD-3-Clause, `go-git` Apache 2.0, Wails/Vue/Shiki MIT.

## Risks / Trade-offs

- [Only a small command set is guaranteed on Windows] → authors use the course's own language for text processing, or verify extra tools in the environment check.
- [`moreinterp/coreutils` is untagged and unstable] → pin to a commit; update deliberately.
- [No sandbox: a malicious course can do anything the learner can] → one explicit warning on adding a course; this is the same trust as running `make` in a cloned repository.
- [Signed exe still triggers SmartScreen until publisher reputation builds up] → accepted; reputation grows with downloads.
- [Strict YAML: an old app fails on any new field] → every schema change bumps `version`, so the failure is a clear "requires a newer app version" message.
- [No folder watching for local-folder courses] → explicit "Reload" button.

## Open Questions

- Final app name and `<schema-base>` URL — before the first public release.
