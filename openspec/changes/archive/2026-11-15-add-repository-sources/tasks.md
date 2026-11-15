# Tasks

## 1. Remove the go-git spike

- [x] 1.1 Delete `internal/source/source.go` (`CloneTags`) and `source_test.go`, move the package comment to another file of `internal/source`, run `go mod tidy`; verify go-git's transport/ssh/crypto dependencies are gone from `go.mod` (go-git stays for the `.gitignore` matcher of `CopyLocal`) and `task go:test` passes

## 2. Process tree for git

- [x] 2.1 Export the process tree of `internal/runtime` (design decision 1) so a plain `exec.Cmd` can be started in it with `CREATE_NO_WINDOW` on Windows and killed with all descendants on context cancel; keep script runs on it unchanged; verify existing runtime tests pass with `task go:test` and `task go:test:windows`, plus a new test that cancelling kills a child of the started process

## 3. Package internal/git

- [x] 3.1 Add `internal/git` with `Run(ctx, dir, args...)`: `LookPath` → `GitNotFound`, the fixed `-c` options, the environment (`GIT_TERMINAL_PROMPT=0`, `GIT_LFS_SKIP_SMUDGE=1`, `LC_ALL=C`, `LANGUAGE=C`), the empty hooks folder, and a test-only option allowing the `file` protocol; verify with tests that check the built command line and environment, and that a `file://` clone is refused without the option
- [x] 3.2 Build `GIT_SSH_COMMAND` from the learner's `GIT_SSH_COMMAND`, else `core.sshCommand`, else `ssh`, with `-o BatchMode=yes` appended; verify with table tests for all three sources
- [x] 3.3 Add the new `apperr` codes (`GitNotFound`, `GitFailed` with `output`, `RepoAddressInvalid`, `RepoAccessDenied`, `RepoNotFound`, `RepoUnreachable`, `SshHostUnknown`, `NoVersions`, `VersionNotFound`), list them in `All`, run `go generate ./internal/apperr`, add `errors.<code>` messages to `en.json` and `ru.json`; verify the apperr Go test and the frontend codes test pass
- [x] 3.4 Map git failures to codes by the stderr patterns of design decision 1, falling back to `GitFailed` with the last lines of stderr, and log the full output; verify with table tests on captured real git/ssh outputs (auth failed, terminal prompts disabled, publickey, repository not found, 404, host key verification, could not resolve host, connection timeout, unknown)
- [x] 3.5 Add version helpers: parse `vX.Y.Z` strictly (no leading zeros, no pre-release), sort descending, pick the highest; verify with tests including `v1.10.0` > `v1.2.0`, `v2.0.0-beta.1` and `release-3` ignored

## 4. Repository addresses and storage keys

- [x] 4.1 Add `source.NormalizeRepo` per design decision 2 (https, ssh://, scp-like, `github.com/` shorthand; `http://`, `git://`, `file://` and others rejected with `RepoAddressInvalid`); verify with table tests for every row of the design table and the spec scenarios
- [x] 4.2 Extend `ComparisonForm` with the OS-independent repository rule and make `label()` drop a trailing `.git` for repository sources; verify with tests: `https://GitHub.com/Owner/Repo.git/` and `github.com/owner/repo` give the same key on every OS, https vs ssh give different keys, `repo-<hash>` form, and existing local-folder key tests unchanged (`task go:test` and `task go:test:windows`)

## 5. Library: adding a repository course

- [x] 5.1 Add a test helper that creates a bare repository with commits and `vX.Y.Z` tags (including two tags on one commit and a tag not reachable from any branch) in a temp folder using the test-only `file` option
- [x] 5.2 Implement `Library.RemoteVersions(addr)`: normalize, return an existing entry without running git, else `ls-remote --tags --refs`, filter and sort, `NoVersions` when empty; verify with tests: existing course returns its entry and runs no git, versions sorted, no-tags repository, missing repository maps to `RepoNotFound`
- [x] 5.3 Implement the tag fetch with explicit non-forced refspecs for missing tags only (design decision 5) and use it from add; verify with a test that a tag moved in the remote keeps its old commit locally after a fetch, and a new tag arrives
- [x] 5.4 Implement `Library.AddRepo(ctx, source, version)` per design decision 3: clone `--no-checkout --no-tags` into `<key>.new-<ts>`, fetch tags, checkout the tag detached, parse the manifest, rename and append the entry with `origin: "repo"`; verify with tests: added with the right title and type (course and workshop), invalid manifest at that version leaves no folder and no entry, unknown version → `VersionNotFound`, cancelled context leaves no folder and no entry, a `.new-` leftover is removed by `Recover`, `core.autocrlf=true` in a test global config still gives LF files

## 6. Library: versions, switch, check, repair

- [x] 6.1 Read versions and the selected version from the clone (design decision 4) and add `Version`/`Versions` to `Overview` and `Outdated` to `Course`; verify with tests: highest of two tags on HEAD is selected, `Outdated` true only when a newer local tag exists, no git → `Outdated` false and overview still loads, local-folder courses unaffected
- [x] 6.2 Extract the title/type/progress refresh from `Sync` into a helper and implement `Library.SwitchVersion(key, version)` per design decision 6; verify with tests: running control and environment check stopped first, local edits and untracked files gone, broken manifest at the target leaves the overview with a course error and allows switching back, new activity un-completes a completed course, title updated only from a valid manifest, `Sync` tests still pass
- [x] 6.3 Implement `Library.CheckVersions(key)` returning the new versions; verify with tests: new tag found and selected version unchanged, nothing new returns an empty list, unreachable remote returns an error and leaves tags unchanged, moved tag not updated
- [x] 6.4 Add the repair of repository course copies to `Recover` (design decision 7); verify with tests: a clone with modified and untracked files is clean after `Recover`, no git → logged and other recovery steps still run
- [x] 6.5 Make deleting a course copy clear read-only attributes on Windows (design decision 8; Go's `os.Remove` already does, so no code change); verify with `task go:test:windows -- -run Remove` on a test that removes a repository course whose clone has packed objects

## 7. App bindings

- [x] 7.1 Add `RemoteVersions`, `AddRepoCourse`, `CancelAdd` (single cancel func under a mutex), `SwitchVersion` and `CheckVersions` with their `*Result` types to `app.go`; verify `task go:lint` passes and the Wails bindings regenerate with these methods and the new `Overview`/`Course` fields in `frontend/wailsjs`

## 8. Frontend

- [x] 8.1 Keep "Add from folder" and add "Add from repository" to the library, and add `library/AddRepoDialog.vue` per design decision 9 (address → versions with warning → cloning with Cancel), sharing the script warning text; add `addRepo.*` strings to `en.json` and `ru.json` that name neither courses nor workshops; verify with vitest: both buttons present, existing course opens its overview without calling `AddRepoCourse`, highest version preselected, Cancel calls `CancelAdd`, errors shown at the failing step, success navigates to the overview
- [x] 8.2 Show the "newer version available" mark in `LibraryView.vue` from `Outdated`; verify with a vitest case
- [x] 8.3 Add the version selector and "Check for new versions" to `CourseOverview.vue` for repository courses (no Sync), disabled while switching, checking or enrolling, kept with a course error; show found versions or "No new versions"; verify with `CourseOverview.test.ts` cases: repository vs local-folder controls, `SwitchVersion` called and overview reloaded, check result messages, error shown, controls available with a broken manifest
- [x] 8.4 Run `task frontend:check` and verify it passes

## 9. Author guide

- [x] 9.1 Update `docs/authoring.md`: a section on publishing a course or workshop from a git repository (accepted addresses, a course version is a `vX.Y.Z` tag and other tags are ignored, the Learner picks the version and finds new ones only with "Check for new versions", never move or delete a released tag but release a new version, the clone holds files exactly as committed so line endings are not converted, Git LFS and submodules are not supported, private repositories work through the Learner's git authentication); mention "Add from repository" next to "Add from folder" in the getting-started part; remove "Repository courses and course versions" from "Not yet available"; verify the guide describes only behavior implemented by this change and agrees with the delta specs

## 10. Integration

- [x] 10.1 Create three GitHub repositories for manual checks: a public and a private course holding a copy of `tmp/courses/01-full` with tags `v1.0.0`, `v1.1.0` (a new activity and changed title) and `v1.2.0` (broken `course.yaml`), and a public workshop holding a copy of `tmp/workshops/01-workshop` with tags `v1.0.0` and `v1.1.0`; verify by running these checks in `task run` on Windows:
  - public course added by https and by the `github.com/owner/repo` shorthand (the second opens the existing course)
  - private course added by https through Git Credential Manager and by ssh with an agent
  - ssh with an unknown host and with a locked key without an agent fail fast with an error
  - add, then Cancel during the clone: nothing in the library, no folder in `<data>/courses`
  - switch versions both ways, including to the broken `v1.2.0` (course error shown, switching back works)
  - "Check for new versions" before and after pushing `v1.3.0`; the selected version stays
  - a tag moved in the repository is not picked up by "Check for new versions"
  - the library marks the course as having a newer version, and stops after switching to it
  - removing a repository course deletes its clone
  - the workshop is listed under "Workshops"; switching its version leaves the workshop folder untouched and keeps progress; removal deletes the workshop folder unless "Keep my files"
- [x] 10.2 Run `task check` and verify it passes
