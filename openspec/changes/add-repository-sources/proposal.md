# Proposal

## Why

Authors publish courses as git repositories, but the app can only add a course from a local folder, so an Author has no way to hand a course to Learners. The specs promise repository courses with course versions, yet they rely on a built-in git library and allow only public https repositories, which rules out internal company courses in private repositories.

## What Changes

- Add a course or workshop from a git repository: the Learner enters an address, the app lists the repository's `vX.Y.Z` tags (latest preselected), shows the script warning, clones the repository and adds the course if `course.yaml` (or `workshop.yaml`) at the chosen version is valid. A long clone can be cancelled.
- **BREAKING (spec)**: the app uses the git installed on the Learner's machine instead of a built-in git library. Authentication is left to git: credential helpers (Git Credential Manager) for https, ssh-agent or keys for ssh. Private repositories work; without git, repository courses cannot be added, local-folder courses still work.
- Accepted addresses: `https://…`, `ssh://…`, scp-like `user@host:path` and the shorthand `github.com/owner/repo`. An https and an ssh address of one repository are different courses.
- **BREAKING (spec)**: the course source of a repository course keeps its scheme (`https://github.com/owner/repo`), instead of the scheme-less form the current spec describes. No repository course exists yet, so no stored data changes.
- git runs non-interactively: no terminal prompts, ssh in batch mode, the Learner's line-ending conversion and hooks disabled, Git LFS disabled, no submodules.
- The course copy is a full clone. The selected version is the `vX.Y.Z` tag at HEAD (the highest one when several point at it). Switching versions in the course overview stops the course's scripts and checks out the tag, discarding any local changes; an interrupted switch is repaired on startup.
- **BREAKING (spec)**: tags are no longer fetched when a course opens. The course overview gets a "Check for new versions" button; nothing else touches the network after adding. Tags already known to the app are never replaced, even if the Author moves a tag.
- The library marks a repository course whose selected version is not the latest tag known locally.
- Remove the go-git spike (`source.CloneTags`) and the go-git dependency.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `library-source`: repository addresses and their normalization, system git and its non-interactive environment, adding a repository course, the selected version, version switch and its repair, "Check for new versions" replacing the fetch on open, immutable known tags, no LFS and submodules; the "public https only" requirement is removed.
- `library`: the library entry's "not on the latest version" mark comes from locally known tags; the course overview of a repository course shows "Check for new versions" next to the version selector.
- `workshop`: "Adding names neither content type" covers the "Add from repository" button and its dialog.

## Impact

- `internal/source`: repository address parsing and normalization, comparison form and storage key label for repository sources; go-git spike removed.
- New git runner (package placement in design): runs the system git with a fixed environment, cancellable, kills the whole process tree, maps failures to `apperr` codes.
- `internal/library`: `AddRepo`, remote version listing, `SwitchVersion`, `CheckVersions`, selected version and versions in the overview, "not latest" in the library list, `Recover` repairs repository course copies.
- `internal/apperr`: new codes (git not found, invalid address, access denied, repository unreachable, no versions, version not found, git failed), messages in `en.json` and `ru.json`.
- `app.go`: new bound methods for listing remote versions, adding, cancelling an add, switching versions and checking for new versions; Wails bindings regenerate.
- Frontend: the library gets "Add from repository" with an address/version/warning dialog; the course overview gets the version selector and "Check for new versions".
- `go.mod`: go-git and its dependencies removed.
- No change to the `library.json` format (`origin: "repo"` already exists) or to progress files.
- `docs/authoring.md`: publishing from a git repository (versions as `vX.Y.Z` tags, never move a released tag, no Git LFS or submodules, private repositories), "Repository courses and course versions" leaves "Not yet available".
