# Design

## Context

The code is the skeleton from `2026-10-26-app-skeleton`: `internal/platform` (data and log folders, log), a minimal `internal/source` (`CloneTags`), a minimal `internal/runtime`, and a smoke screen bound through `Smoke*` methods on `App`. There is no persisted state, no `course.yaml` parsing and no routing in the frontend. The Markdown renderer (`frontend/src/markdown/markdown.ts`, markdown-it with `html: false` and Shiki) already exists. The baseline design fixed state as JSON files in the data folder written via temp file and rename. See proposal.md for scope; specs in this change for the storage key, course copy and Sync behavior.

## Goals / Non-Goals

**Goals:**
- A course copy layout and storage key that the repository-course change can reuse unchanged.
- `course.yaml` parsing whose single source of truth is the JSON Schema authors will use.
- Course files served to the webview without letting a course reach files outside its copy.

**Non-Goals:**
- Publishing the JSON Schema at `<schema-base>`; only its file in the repository and its embedding.
- `activity.yaml` and its schema; no activity is opened in this change.
- `state/<key>.json`; nothing writes progress yet.
- Opening links from the overview text in an external browser.

## Decisions

### Package layout
Following the skeleton's one-package-per-capability guidance:

```
internal/
  source/     NormalizeLocal, Key, CopyLocal (.gitignore-aware copy)   <- course-source
  manifest/   course.yaml: schema validation, Course struct, semantics <- course-manifest
  library/    library.json, Add, Sync, Overview; owns <data>/courses   <- course-library
  platform/   + WriteFileAtomic
schemas/
  course/v1.json   + schemas.go (go:embed)
```

`library` orchestrates; `source` and `manifest` know nothing about `library.json`. `App` binds thin methods that delegate to `library`.
- *Alternative: storage key in `library`:* it is derived purely from the course source and the repository-course change needs it for clone paths, so it belongs with the source.

### Data folder layout

```
<data>/library.json
<data>/courses/<key>/          course copy (later: clone of a repository course)
<data>/courses/<key>.new-*/    Sync in progress
<data>/courses/<key>.old-*/    previous copy during the swap
<data>/state/<key>.json        future, progress change
```

`library.json`:

```json
{
  "version": 1,
  "courses": [
    { "origin": "local", "source": "C:\\Courses\\Go Basics", "key": "go-basics-3f2a9c1e", "title": "Go Basics" }
  ]
}
```

- `version` is the file's format version. A newer app migrates older files on read. An app that finds a version higher than it supports does not write the file, so running an older exe after a newer one cannot drop data it does not understand.
- `origin` (`local` or `repo`) is explicit instead of guessed from the shape of `source`. It is not called `kind`, which the domain language reserves for block and question kinds.
- Other content types added later (e.g. labs) get their own top-level array and their own folder under `<data>/`, not a shared list with a type discriminator.
- *Alternative: one `entries` list with a content-type field:* one ordering and one uniqueness check for sources, but every reader filters by type and every type shares one entry shape.

Writes go through `platform.WriteFileAtomic` (temp file in the same folder, `fsync`, rename over the target); a mutex in `library` serializes all reads-modify-writes.

### Local-folder source normalization
`filepath.Abs` → `filepath.EvalSymlinks` → `filepath.Clean`; the folder must exist and be a directory. The comparison form is the result itself on Linux and macOS and `strings.ToLower` of it on Windows (build-tagged file). Two sources are the same course when their comparison forms are equal; the stored `source` is the spelling of the first add.
- *Alternative: case-insensitive on macOS too:* APFS is usually case-insensitive but not always; Linux and macOS are not shipped in v1, so they keep exact comparison until there is a reason.

### Storage key
`key = label + "-" + hex(sha256(comparison form))[:8]`.
- label: the last segment of the source (folder base name, or the last path segment of a repository URL), lowercased, every run of characters outside `[a-z0-9]` replaced by `-`, leading and trailing `-` trimmed, cut to 32 characters; `course` if empty.
- The comparison form is hashed, so different spellings of one course on Windows give one key.
- 32 hash bits make a collision in one learner's library practically impossible; if `Add` still finds the key taken by another source it fails with an error instead of sharing files.
- The algorithm is fixed by tests with literal expected keys; changing it would orphan stored data.
- *Alternative: full hash only:* safe but unreadable when matching folders by hand. *Random id:* cannot be recomputed when a course removed with "Keep progress" is added again.

### Copying a local folder
`source.CopyLocal(src, dst)` walks `src` with `filepath.WalkDir`:
- `.git` is skipped at any depth.
- Ignore rules come from go-git's `plumbing/format/gitignore`: `ReadPatterns` over an `osfs` of `src` collects every `.gitignore` inside the folder with its domain, and `NewMatcher(...).Match(pathParts, isDir)` decides per entry. Global and system patterns and `.git/info/exclude` are not loaded. An ignored directory is skipped without descending.
- Regular files are copied with their mode bits; symbolic links are skipped and logged, so nothing outside the folder ends up in the copy.
- *Alternative: own `.gitignore` matcher:* git's rules (anchoring, `**`, negation, directory-only patterns, nesting) are easy to get subtly wrong; go-git is already a dependency.
- *Alternative: `git ls-files`-like listing via go-git index:* would exclude untracked new files the author has not added yet, which surprises an author mid-edit, and needs the folder to be a repository.

### Add and Sync share one swap
Both build the new copy in `<data>/courses/<key>.new-<random>/`, then:
- **Add**: parse `course.yaml` from the new copy; if invalid, delete it and return the course error. Otherwise rename it to `<key>/` and append the entry with the title to `library.json`. A source already in the library returns the existing entry without copying.
- **Sync**: stop the course's scripts (no-op until the course screen exists), rename `<key>/` to `<key>.old-<random>/`, rename `.new-*` to `<key>/`, delete `.old-*`. No validation before the swap. Afterwards, if `course.yaml` is valid, its title replaces the stored one. If copying fails (folder gone, unreadable file), `.new-*` is deleted and `<key>/` is untouched.
- **Startup recovery**: for each library entry, if `<key>/` is missing and a `.old-*` exists, the newest `.old-*` is renamed back; all remaining `.new-*` and `.old-*` folders are deleted.
- Renames retry briefly on Windows, where antivirus or indexer handles can make a directory rename fail transiently.
- *Alternative: copy in place over `<key>/`:* an interrupted Sync leaves a mixed copy, against the spec.
- *Alternative: versioned subfolders with a pointer file:* no directory rename on the hot path, but every reader must resolve the pointer and old versions need garbage collection; not worth it for a single-learner app.

### `course.yaml` parsing
```
read bytes -> yaml.v3 decode to any -> read `version`
   version > 1 or not an int  -> "requires a newer app version" (no further parsing)
-> marshal to JSON -> validate against embedded schemas/course/v1.json (santhosh-tekuri/jsonschema/v6)
-> yaml.v3 decode into manifest.Course
-> semantic checks: unique section slugs, course-wide unique activity slugs,
   `check` path relative, `/`-separated, no `..`, file exists in the course copy
```
- The schema expresses types, required fields, `additionalProperties: false`, the slug pattern and non-empty lists, so unknown fields like `autor` are reported by the schema with their location.
- Converting through JSON gives the validator the same data model as authors' editors.
- All problems are collected into one course error as a list of coded problems with their YAML path (e.g. `sections[0].activities[1].slug`); see "Errors are codes, translated in the frontend".
- Activity folders are not touched (lazy reading).
- The schema's `$id` is provisional (`https://github.com/olymp-software/chiron/schemas/course/v1.json`); the schema is registered with the compiler from the embedded bytes, so no network access ever happens.
- *Alternative: strict decode only (`KnownFields`):* the published schema would become a second, drifting source of truth.
- *Alternative: `goccy/go-yaml`:* better error positions, but yaml.v3 (`go.yaml.in/yaml/v3`) is the common choice and positions are not required by the specs.

### Errors are codes, translated in the frontend
Every error the app shows is known in advance and has an explicit code. Go never returns user-facing text; bound methods return errors as data and the frontend renders them through `vue-i18n` with an `en` and a `ru` message per code.

```json
{ "code": "manifestInvalid",
  "problems": [
    { "code": "unknownField", "path": "sections[0].activities[1]", "params": { "field": "autor" } },
    { "code": "duplicateActivitySlug", "path": "sections[1].activities[0].slug", "params": { "slug": "intro" } }
  ] }
```

- Codes live in one Go file as typed constants; the frontend message keys are `errors.<code>` with `params` as named interpolation values.
- Schema problems: each JSON Schema keyword the course schema uses (`required`, `additionalProperties`, `type`, `pattern`, `minItems`, ...) maps to one code, taking its params from the validator's error kind. A test walks `schemas/course/v1.json` and fails if it uses a keyword without a mapping, so a schema change cannot produce an unmapped error.
- Semantic and app errors have their own codes: `manifestMissing`, `unsupportedSchemaVersion`, `duplicateSectionSlug`, `duplicateActivitySlug`, `pathEscapesCourse`, `pathNotFound`, `overviewFileAmbiguous`, `folderNotFound`, `notAFolder`, `storageKeyTaken`, `libraryNewerFormat`, `copyFailed`, ...
- Unexpected I/O failures (permissions, disk full) use a small set of codes such as `copyFailed` with the failing path as a param; the underlying Go error goes to the log, not to the UI.
- A Go test lists all codes and a frontend test checks that `en.json` and `ru.json` have a message for each, so a new code without translations fails CI. The code list is exported to the frontend as a generated TypeScript union so `t()` keys are type-checked.
- Wails rejects a bound call with a string on `error`; the bound methods therefore return a result struct carrying either the value or the coded error instead of a Go `error`.
- *Alternative: English messages from Go:* simplest, but the learner sees a broken course too and the UI must be available in Russian.
- *Alternative: translating in Go (`x/text/message`):* the UI language lives in the frontend, and translations would be split across two stacks.

### Course overview data
`library.Overview(source)` returns title, origin, source, key, and either the overview text or a course error. The overview file is searched in the copy root by `strings.EqualFold(name, "OVERVIEW.md")`; more than one match is a course error. Without it the `description` is returned as plain text. The overview is computed on every open, from the copy, not cached.

### Course files for the webview
Wails `assetserver.Options.Handler` handles requests the embedded assets do not. It serves `GET /course-files/<key>/<path>`:
- `<key>` must belong to a library entry; `<path>` is opened through `os.OpenRoot(<data>/courses/<key>)`, which rejects `..`, absolute paths and symlinks escaping the root.
- Only image extensions (`png`, `jpg`, `jpeg`, `gif`, `webp`, `svg`) are served, with their content type; anything else is 404. SVG is shown through `<img>`, where its scripts do not run.

In the frontend, `renderMarkdown` gains an optional image resolver. A markdown-it `image` rule override passes each `src` to it: relative paths become `/course-files/<key>/<path relative to the course root>`; anything with a scheme, `//` or a leading `/` gets no `src` and renders its alt text, so remote images are never requested.
- *Alternative: inline images as data URLs from Go:* no handler, but large overview images bloat each call and the same mechanism is needed for theory blocks.

### Frontend screens
No router: `App.vue` holds the current view (`library` or `overview` of a source) in a ref. New folders `frontend/src/library/` (library list, Add folder) and `frontend/src/overview/` (course overview). Add: Go opens the native folder picker (`runtime.OpenDirectoryDialog`, root package only) → the frontend shows the script warning → `AddLocalCourse(path)` → the overview of the returned course. The smoke screen and the `Smoke*` methods are removed; `source.CloneTags` stays for the repository-course change.

## Risks / Trade-offs

- [Directory rename fails on Windows while a file in the copy is open, e.g. an image being served] → short retries; on final failure the old copy stays and Sync reports an error.
- [Large author folders (ignored `node_modules` aside) make add and Sync slow] → acceptable for v1; the copy runs off the UI thread and the Sync button shows progress state.
- [`.gitignore` semantics of go-git differ from git in an edge case] → tests with the spec scenarios (nested files, negation, directory patterns); differences are reported upstream rather than patched locally.
- [Symlinks in an author's folder are silently dropped] → logged; authors publish via git, where symlinks are rare in course content.
- [Storage key algorithm cannot change later] → literal-value tests; any future change needs an explicit migration.
- [Smoke screen no longer exercises git and scripts in the packaged exe] → the library exercises file copying, the asset handler and Markdown in the exe; git and scripts get their real UI in the next changes.

## Migration Plan

No existing data: the app has never written `library.json` or course copies. Leftover `<data>/smoke/` folders from the skeleton are deleted on startup.

## Open Questions

- Final `<schema-base>` URL and `$id` of the published schemas; decided by the schema publishing change, without changing parsing.
