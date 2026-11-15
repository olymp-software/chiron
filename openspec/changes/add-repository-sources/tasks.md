# Tasks

## 1. Remove go-git

- [ ] 1.1 Delete `internal/source/source.go` (`CloneTags`) and `source_test.go`, move the package comment to another file of `internal/source`, run `go mod tidy`; verify go-git is gone from `go.mod` and `task go:test` passes

## 2. Process tree for git

- [ ] 2.1 Export the process tree of `internal/runtime` (design decision 1) so a plain `exec.Cmd` can be started in it with `CREATE_NO_WINDOW` on Windows and killed with all descendants on context cancel; keep script runs on it unchanged; verify existing runtime tests pass with `task go:test` and `task go:test:windows`, plus a new test that cancelling kills a child of the started process

## 3. Package internal/git

- [ ] 3.1 Add `internal/git` with `Run(ctx, dir, args...)`: `LookPath` → `GitNotFound`, the fixed `-c` options, the environment (`GIT_TERMINAL_PROMPT=0`, `GIT_LFS_SKIP_SMUDGE=1`, `LC_ALL=C`, `LANGUAGE=C`), the empty hooks folder, and a test-only option allowing the `file` protocol; verify with tests that check the built command line and environment, and that a `file://` clone is refused without the option
- [ ] 3.2 Build `GIT_SSH_COMMAND` from the learner's `GIT_SSH_COMMAND`, else `core.sshCommand`, else `ssh`, with `-o BatchMode=yes` appended; verify with table tests for all three sources
- [ ] 3.3 Add the new `apperr` codes (`GitNotFound`, `GitFailed` with `output`, `RepoAddressInvalid`, `RepoAccessDenied`, `RepoNotFound`, `RepoUnreachable`, `SshHostUnknown`, `NoVersions`, `VersionNotFound`), list them in `All`, run `go generate ./internal/apperr`, add `errors.<code>` messages to `en.json` and `ru.json`; verify the apperr Go test and the frontend codes test pass
- [ ] 3.4 Map git failures to codes by the stderr patterns of design decision 1, falling back to `GitFailed` with the last lines of stderr, and log the full output; verify with table tests on captured real git/ssh outputs (auth failed, terminal prompts disabled, publickey, repository not found, 404, host key verification, could not resolve host, connection timeout, unknown)
- [ ] 3.5 Add version helpers: parse `vX.Y.Z` strictly (no leading zeros, no pre-release), sort descending, pick the highest; verify with tests including `v1.10.0` > `v1.2.0`, `v2.0.0-beta.1` and `release-3` ignored

## 4. Repository addresses and storage keys

- [ ] 4.1 Add `source.NormalizeRepo` per design decision 2 (https, ssh://, scp-like, `github.com/` shorthand; `http://`, `git://`, `file://` and others rejected with `RepoAddressInvalid`); verify with table tests for every row of the design table and the spec scenarios
- [ ] 4.2 Extend `ComparisonForm` with the OS-independent repository rule and make `label()` drop a trailing `.git` for repository sources; verify with tests: `https://GitHub.com/Owner/Repo.git/` and `github.com/owner/repo` give the same key on every OS, https vs ssh give different keys, `repo-<hash>` form, and existing local-folder key tests unchanged (`task go:test` and `task go:test:windows`)

## 5. Library: adding a repository course

- [ ] 5.1 Add a test helper that creates a bare repository with commits and `vX.Y.Z` tags (including two tags on one commit and a tag not reachable from any branch) in a temp folder using the test-only `file` option
- [ ] 5.2 Implement `Library.RemoteVersions(addr)`: normalize, return an existing entry without running git, else `ls-remote --tags --refs`, filter and sort, `NoVersions` when empty; verify with tests: existing course returns its entry and runs no git, versions sorted, no-tags repository, missing repository maps to `RepoNotFound`
- [ ] 5.3 Implement the tag fetch with explicit non-forced refspecs for missing tags only (design decision 5) and use it from add; verify with a test that a tag moved in the remote keeps its old commit locally after a fetch, and a new tag arrives
- [ ] 5.4 Implement `Library.AddRepo(ctx, source, version)` per design decision 3: clone `--no-checkout --no-tags` into `<key>.new-<ts>`, fetch tags, checkout the tag detached, parse the manifest, rename and append the entry with `origin: "repo"`; verify with tests: added with the right title and type (course and workshop), invalid manifest at that version leaves no folder and no entry, unknown version → `VersionNotFound`, cancelled context leaves no folder and no entry, a `.new-` leftover is removed by `Recover`, `core.autocrlf=true` in a test global config still gives LF files

## 6. Library: versions, switch, check, repair

- [ ] 6.1 Read versions and the selected version from the clone (design decision 4) and add `Version`/`Versions` to `Overview` and `Outdated` to `Course`; verify with tests: highest of two tags on HEAD is selected, `Outdated` true only when a newer local tag exists, no git → `Outdated` false and overview still loads, local-folder courses unaffected
- [ ] 6.2 Extract the title/type/progress refresh from `Sync` into a helper and implement `Library.SwitchVersion(key, version)` per design decision 6; verify with tests: running control and environment check stopped first, local edits and untracked files gone, broken manifest at the target leaves the overview with a course error and allows switching back, new activity un-completes a completed course, title updated only from a valid manifest, `Sync` tests still pass
- [ ] 6.3 Implement `Library.CheckVersions(key)` returning the new versions; verify with tests: new tag found and selected version unchanged, nothing new returns an empty list, unreachable remote returns an error and leaves tags unchanged, moved tag not updated
- [ ] 6.4 Add the repair of repository course copies to `Recover` (design decision 7); verify with tests: a clone with modified and untracked files is clean after `Recover`, no git → logged and other recovery steps still run
- [ ] 6.5 Make deleting a course copy clear read-only attributes on Windows (design decision 8); verify with `task go:test:windows -- -run Remove` on a test that removes a repository course whose clone has packed objects

## 7. App bindings

- [ ] 7.1 Add `RemoteVersions`, `AddRepoCourse`, `CancelAdd` (single cancel func under a mutex), `SwitchVersion` and `CheckVersions` with their `*Result` types to `app.go`; verify `task go:lint` passes and the Wails bindings regenerate with these methods and the new `Overview`/`Course` fields in `frontend/wailsjs`

## 8. Frontend

- [ ] 8.1 Keep "Add from folder" and add "Add from repository" to the library, and add `library/AddRepoDialog.vue` per design decision 9 (address → versions with warning → cloning with Cancel), sharing the script warning text; add `addRepo.*` strings to `en.json` and `ru.json` that name neither courses nor workshops; verify with vitest: both buttons present, existing course opens its overview without calling `AddRepoCourse`, highest version preselected, Cancel calls `CancelAdd`, errors shown at the failing step, success navigates to the overview
- [ ] 8.2 Show the "newer version available" mark in `LibraryView.vue` from `Outdated`; verify with a vitest case
- [ ] 8.3 Add the version selector and "Check for new versions" to `CourseOverview.vue` for repository courses (no Sync), disabled while switching, checking or enrolling, kept with a course error; show found versions or "No new versions"; verify with `CourseOverview.test.ts` cases: repository vs local-folder controls, `SwitchVersion` called and overview reloaded, check result messages, error shown, controls available with a broken manifest
- [ ] 8.4 Run `task frontend:check` and verify it passes

## 9. Author guide

- [ ] 9.1 Update `docs/authoring.md`: a section on publishing a course or workshop from a git repository (accepted addresses, a course version is a `vX.Y.Z` tag and other tags are ignored, the Learner picks the version and finds new ones only with "Check for new versions", never move or delete a released tag but release a new version, the clone holds files exactly as committed so line endings are not converted, Git LFS and submodules are not supported, private repositories work through the Learner's git authentication); mention "Add from repository" next to "Add from folder" in the getting-started part; remove "Repository courses and course versions" from "Not yet available"; verify the guide describes only behavior implemented by this change and agrees with the delta specs

## 10. Integration

- [ ] 10.1 Create two GitHub repositories for manual checks, a public and a private one, holding a copy of `tmp/courses/01-full` with tags `v1.0.0`, `v1.1.0` (a new activity and changed title) and `v1.2.0` (broken `course.yaml`), and add their checks to `tmp/courses/CHECKLIST.md`: add public by https and shorthand, add private by https through Git Credential Manager and by ssh with an agent, ssh with an unknown host and with a locked key fail fast, add then cancel during clone, switch versions both ways including the broken one, "Check for new versions" before and after pushing `v1.3.0`, moved tag not picked up, outdated mark in the library, remove a repository course; plus a third public repository holding a copy of `tmp/workshops/01-workshop` with tags `v1.0.0` and `v1.1.0`: added under "Workshops", switching versions leaves the workshop folder untouched and keeps progress, removal deletes the workshop folder unless "Keep my files"; add these checks to `tmp/workshops/CHECKLIST.md`; verify by running them in `task run` on Windows
- [ ] 10.2 Run `task check` and verify it passes
