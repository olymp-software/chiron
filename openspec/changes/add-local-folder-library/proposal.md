# Proposal

## Why

The app has specs for the whole product but code only for the skeleton: there is no course manifest parsing, no persisted state and no screen other than the smoke screen. The library with local-folder courses is the first product slice: it needs no git, tags or network, and it is the path authors use to build the test courses every later feature relies on. Building it now also fixes two problems in the specs before code depends on them: a course source containing `:` or `/` cannot be used as a file or folder name on disk, and reading a local-folder course straight from the author's folder makes it behave differently from a repository course.

## What Changes

- Introduce the **storage key**: a file-system-safe name `<label>-<hash>` derived from the course source. State files, course copies and learner folders are named by it instead of mirroring the course source. The library stores it next to the source so a person can match files on disk to courses; the course overview shows both.
- Define normalization of a local-folder course source: absolute, cleaned, symlinks resolved; compared case-insensitively on Windows.
- **BREAKING** (spec-level): a local-folder course is no longer used as is. The app copies the folder into its own course copy, honoring `.gitignore` files inside the folder and skipping `.git`, and works only from that copy. The learner refreshes it with **Sync** in the course overview, in the place of the version selector of a repository course. Sync replaces the copy atomically, does not validate it in advance, and leaves the copy untouched if the folder cannot be read.
- **BREAKING** (spec-level): the "Reload" button on the course screen is removed.
- Adding a course requires a valid `course.yaml`; otherwise the course is not added. The course title is stored in the library when a course is added and updated whenever the course copy changes and `course.yaml` is valid.
- Implement the library and the course overview for local-folder courses: add from a folder with the script warning, list courses, open the overview with the title, source, storage key, Sync, the overview text with relative images, and course errors. Implement `course.yaml` parsing validated by an embedded JSON Schema.
- Not in this change: repository courses, removing a course, Enroll / Continue, the environment check UI, the course screen. The course overview shows no Enroll, Continue or environment check controls until those are implemented.
- Remove the temporary smoke screen.
- Update the domain glossary in `openspec/config.yaml`: Local-folder course and Course overview change; add Course copy, Storage key and Sync.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `course-source`: storage key, local-folder source normalization, the course copy for local-folder courses with Sync instead of Reload, a valid `course.yaml` required to add a course, the course title stored in the library.
- `course-library`: the course overview shows the course source, storage key and Sync; library titles come from the stored title; entering a course no longer happens by Reload.
- `learner-workspace`: learner folders are placed under the storage key instead of mirroring the course source.
- `script-runtime`: the course copy of a local-folder course is also overwritten, on Sync.

## Impact

- Specs: `course-source`, `course-library`, `learner-workspace`, `script-runtime`. Glossary in `openspec/config.yaml`.
- Code: new `internal/manifest` (course.yaml, JSON Schema), `internal/library` (library.json, storage key, adding, Sync); `internal/source` gains the local-folder copy; `app.go` drops the `Smoke*` methods and binds library methods; a Wails asset handler serves course files for overview images; the Vue frontend replaces `frontend/src/smoke/` with the library and course overview screens.
- New dependencies: `github.com/santhosh-tekuri/jsonschema/v6`, a YAML parser; `.gitignore` matching from `go-git` (already a dependency).
- Course authors: a local-folder course must be synced after edits; files ignored by `.gitignore` are not visible to the app, as they would not be in a published repository.
