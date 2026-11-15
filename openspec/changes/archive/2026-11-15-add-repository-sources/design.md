# Design

## Context

See proposal.md for motivation. Current state that shapes the approach:

- `library.Entry` already has `Origin` (`local` | `repo`); only `local` is ever written. `AddLocal` builds the copy in `<data>/courses/<key>.new-<ts>`, parses the manifest and renames it in under `update`. `Sync` swaps copies and then refreshes title, content type and progress (`progress.Resolve`). `Recover` deletes everything in `<data>/courses` that is not a storage key of a library entry, so any `.new-` folder left by an interrupted add disappears on the next start.
- `source.Key` hashes `ComparisonForm(source)`, which is OS-specific (lowercase on Windows, exact elsewhere) because it was written for paths. Its comment says the algorithm must never change.
- `source.CloneTags` is a go-git spike that nothing calls.
- `internal/runtime` already owns a process tree (job object on Windows, process group on Unix) that kills every descendant of a script; it is unexported.
- `apperr` forbids user-facing text from Go; every learner-visible failure needs a code.

## Goals / Non-Goals

**Goals:**
- git never blocks waiting for terminal input, and its behavior does not depend on the learner's git settings except for authentication and transport (credential helpers, ssh config, proxies, CAs).
- No git failure leaves a half-added course or a clone with mixed versions behind.
- No network access except the three explicit actions: listing versions while adding, cloning, "Check for new versions".

**Non-Goals:**
- Clone progress in percent; the add dialog shows an indeterminate "Cloning…" with Cancel.
- Branches or commits as versions, pre-release tags.
- Git LFS, submodules, shallow clones, storing credentials.
- Detecting that the Author moved or deleted a tag.

## Decisions

### 1. New package `internal/git`: one way to run git

All git calls go through `git.Run(ctx, dir, args...) (stdout []byte, err error)`:

- `exec.LookPath("git")` first; missing → `apperr.GitNotFound`.
- Fixed `-c` options before every subcommand: `core.autocrlf=false`, `core.hooksPath=<empty folder under <data>>`, `filter.lfs.smudge=`, `filter.lfs.process=`, `filter.lfs.required=false`, `protocol.allow=never`, `protocol.https.allow=always`, `protocol.ssh.allow=always`, `advice.detachedHead=false`.
- Environment: the app's environment plus `GIT_TERMINAL_PROMPT=0`, `GIT_LFS_SKIP_SMUDGE=1`, `LC_ALL=C`, `LANGUAGE=C`, and `GIT_SSH_COMMAND` = the learner's effective ssh command (`GIT_SSH_COMMAND`, else `git config core.sshCommand`, else `ssh`) with `-o BatchMode=yes` appended. Appending, as go-getter does for `-i`, keeps the learner's own ssh options.
- System and global git config are **not** disabled (`GIT_CONFIG_NOSYSTEM` unset): Git for Windows registers Git Credential Manager in the system config, and proxies and CAs live there too. Only the options above are overridden.
- Started through the process tree from `internal/runtime`, which is exported for this (`runtime.ProcessTree` or a small `runtime.Command` helper), with `CREATE_NO_WINDOW` on Windows. Cancelling `ctx` kills git and its helpers (`git-remote-https`, `ssh`).
- Failure: stderr (English because of `LC_ALL=C`) is matched against a short list of stable git/ssh messages: `Authentication failed`, `could not read Username`, `terminal prompts disabled`, `Permission denied (publickey` → `RepoAccessDenied`; `Repository not found`, `does not appear to be a git repository`, HTTP 404 → `RepoNotFound`; `Host key verification failed` → `SshHostUnknown`; `Could not resolve host`, `Failed to connect`, `Connection timed out`, `Network is unreachable` → `RepoUnreachable`. Anything else → `GitFailed` with an `output` param (last lines of stderr, shown as a technical detail). Full output is always logged.

*Alternatives:* go-getter (pulls in S3/GCS/HTTP getters, no environment hardening, opaque errors, Terraform itself regrets its detectors); go-git (no credential helpers, so private repositories would need our own token storage and sign-in UI). Rejected in exploration.

The package is portable; only the process start differs per OS and that lives in `runtime`.

Tests need repositories without a network. The runner takes an option, set only by tests, that also allows the `file` protocol; library tests create bare repositories with tags in a temp folder and add them by their `file://` URL through an internal entry point that skips `NormalizeRepo`. The learner-facing path never allows `file`.

### 2. Repository addresses in `internal/source`

`source.NormalizeRepo(addr) (string, error)`:

| Input | Course source |
|---|---|
| `https://Host/Path/` | `https://host/Path` |
| `ssh://User@Host:22/Path` | `ssh://User@host:22/Path` |
| `user@host:path` (scp-like, no `://`) | `ssh://user@host/path` |
| `github.com/o/r` | `https://github.com/o/r` |
| anything else | `RepoAddressInvalid` |

Only scheme and host are lowercased and a trailing `/` removed; the path stays as typed because some servers are case-sensitive and need the `.git` suffix. The scp-like detection follows Terraform's `detectSSH` (`^(?:([^@]+)@)?([^:]+):/?(.+)$`), but without its "user must be `git`" restriction, and only after a `://` check so Windows drive paths never reach it (local paths come from the folder picker anyway).

`ComparisonForm(source)` gets a branch: a source starting with `https://` or `ssh://` is a repository source and is compared OS-independently as lowercased with a trailing `.git` removed. Local paths keep the existing OS-specific rule, so existing storage keys do not change. `label()` drops a trailing `.git` for repository sources (`repo-<hash>`, not `repo-git-<hash>`).

### 3. Adding: list, then clone

Two bound methods instead of one, because the version choice and script warning sit between them:

1. `App.RemoteVersions(addr) VersionsResult{Source, Versions []string, Existing *library.Entry, Error}` → `Library.RemoteVersions`: normalize; if `findSame` finds an entry, return it (frontend opens its overview, no network); else `git ls-remote --tags --refs <source>`, keep `refs/tags/vX.Y.Z`, sort descending by semver; none → `NoVersions`.
2. `App.AddRepoCourse(source, version) CourseResult` → `Library.AddRepo(ctx, source, version)` under `lockCourse(key)`:
   - `git clone --no-checkout --no-tags <source> <courses>/<key>.new-<ts>`
   - fetch the `vX.Y.Z` tags with explicit non-forced refspecs (decision 5), so the clone has every version, not only the tags reachable from branches;
   - `git checkout --detach refs/tags/<version>`; tag gone → `VersionNotFound`;
   - `manifest.Parse`; invalid → delete the folder, return the course error;
   - rename into `<courses>/<key>` and append the entry inside `update`, exactly like `AddLocal`.
   Any failure or cancel deletes the `.new-` folder; an app exit leaves it to `Recover`.
3. `App.CancelAdd()` cancels the context of the running add. One add runs at a time (the dialog is modal), so the App keeps a single cancel func under a mutex.

### 4. Selected version and versions are read from the clone

Not stored in `library.json` (spec: "The app owns the course clone").

- Versions: `git for-each-ref refs/tags --format=%(refname:short)`, filtered by `^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`, sorted by semver.
- Selected: `git tag --points-at HEAD`, filtered the same way, highest wins; none → empty (overview shows no selected version, the learner can still pick one).
- `Overview` gets `Version string` and `Versions []string` for repository courses. `Course` (library list) gets `Outdated bool` = highest known version > selected. Two git calls per repository course per library load; a failure (no git) is logged and leaves `Outdated` false.

A shared helper in `internal/library` (or `internal/git`) parses and orders versions; the same comparison serves `ls-remote` output.

### 5. Fetching tags without ever moving one

"Check for new versions" (`App.CheckVersions(key) CheckVersionsResult{New []string, Error}` → `Library.CheckVersions`):

1. `git ls-remote --tags --refs origin` → remote `vX.Y.Z` tags.
2. Subtract the locally known tags.
3. `git fetch --no-tags origin refs/tags/vA:refs/tags/vA …` for the missing ones only, without `+`.

Existing tags are never named in a refspec, so they cannot move, whatever the git version and its clobber rules. This removes the reason for the minimum git version 2.20 discussed during exploration; no version check is needed. The same step fills the tags during add. It does not touch the worktree, so it does not take the course lock against running scripts, but it holds the lock to serialize with switch and Sync-like operations.

### 6. Version switch in place

`App.SwitchVersion(key, version)` → `Library.SwitchVersion` under `lockCourse(key)`:

1. `StopCheck(key)`, `runs.stopCourse(key)`.
2. `git checkout --force --detach refs/tags/<version>`, then `git clean -ffdx`.
3. Refresh title, content type and progress like `Sync` does; that part of `Sync` becomes a shared helper.

In place rather than a copy-and-swap like Sync: git already writes only changed files, a full copy of a repository with history would be slow, and the repair below makes an interruption harmless.

### 7. Repair on startup

`Recover`, after the existing steps, runs `git reset --hard -q` and `git clean -ffdx` in every repository course copy that exists. A clone interrupted mid-checkout either still has the old HEAD (back to the old version) or the new one (finished), in both cases with exactly that commit's files. No git → logged and skipped; the course still opens.

### 8. Removal and Windows read-only files

`Remove` and `Recover` already delete course copies with `os.RemoveAll`. git writes pack files and loose objects read-only, which `DeleteFile` refuses on Windows. Go's `os.Remove` on Windows already clears the read-only attribute and retries, so no helper of our own is needed; a test that removes a real clone after `git gc` pins this on Windows.

### 9. Frontend

- `library/LibraryView.vue`: "Add from folder" (existing flow) and "Add from repository". Like the folder flow, the dialog names neither courses nor workshops (workshop spec: "Adding names neither content type").
- New `library/AddRepoDialog.vue` (native `<dialog>`): address input → Next (`RemoteVersions`; `Existing` → `router.push` to its overview) → version `<select>` (highest preselected) with the script warning text and Add → "Cloning…" with Cancel (`CancelAdd`) → on success `router.push('/library/<key>')`, on error back to the step that failed with `ErrorView`. The script warning text is shared with `ScriptWarning.vue` through i18n keys.
- `overview/CourseOverview.vue` for `origin === 'repo'`: version `<select>` (change → `SwitchVersion`, then reload), "Check for new versions" (result: "Found: v1.3.0" or "No new versions"), both disabled while either runs or while enrolling; no Sync.
- `LibraryView` shows a "newer version available" mark from `Outdated`.
- New i18n strings in `en.json` and `ru.json`, including messages for the new error codes.

### 10. Removing the go-git spike

`internal/source/source.go` (`CloneTags`) and its test are deleted, `go mod tidy` drops the dependencies only the spike needed (transport, ssh, crypto). The package comment moves to another file of `internal/source`.

go-git itself stays: `source.CopyLocal` uses `plumbing/format/gitignore` to skip files excluded by `.gitignore` when copying a local folder, and that package pulls in only go-billy and gcfg. *Alternatives:* standalone gitignore libraries are either unmaintained for years (sabhiram, denormal, monochromegane) or new with almost no users (git-pkgs/gitignore, bep/gogitignore); go-git's package is maintained and widely copied (e.g. by fluxcd).

## Risks / Trade-offs

- [stderr classification breaks on a future git or ssh wording change] → `LC_ALL=C` fixes the language; unknown messages fall back to `GitFailed` with the raw output, so the learner still sees what happened. Unit tests pin the patterns with captured real outputs.
- [Git Credential Manager may still try a console prompt in some configurations and fail instead of showing a window] → `GIT_TERMINAL_PROMPT=0` makes it fail fast as `RepoAccessDenied`; checked manually on Windows with GCM.
- [Appending `-o BatchMode=yes` to a learner's custom ssh command that is not OpenSSH (e.g. plink)] → plink rejects the option and git fails with `GitFailed` showing the output; acceptable for v1, noted in the docs.
- [`Recover` now runs two git processes per repository course on every start] → They are fast on a clean worktree; if startup time becomes visible, only courses whose switch did not finish could be marked in `library.json` later.
- [Two git calls per repository course on every library load] → Fine for tens of courses; can be cached in memory per process later without spec changes.
- [The clone contains `.git`, which course files could reference] → The image handler serves only images inside the copy, and nothing reads `.git` as course content; manifest paths into `.git` are not specially rejected.
- [Courses that rely on LFS silently show pointer files] → Documented for Authors; the clone is the same on every machine.

## Migration Plan

None: no repository course exists in any `library.json`, the file format is unchanged and local-folder storage keys keep their algorithm. Rollback is reverting the change; a repository entry added in between would be shown by an older build as a course it cannot sync, which only matters to development builds.
