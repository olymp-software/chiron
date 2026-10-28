# Tasks

## 1. Glossary

- [ ] 1.1 Update `openspec/config.yaml`: Local-folder course (works from a course copy refreshed by Sync, no tags), Course overview (version selector or Sync); add Course copy (app-owned folder a course is read and run from; a clone or a copy of a local folder; avoid: cache, mirror), Storage key (file-system-safe name derived from the course source; avoid: course id), Sync (refreshing a local-folder course's copy from its folder; avoid: reload, refresh); verify no other glossary entry still says "used as is" or mentions Reload

## 2. Shared foundations

- [ ] 2.1 Add `platform.WriteFileAtomic` (temp file in the target folder, fsync, rename over the target) and verify with `go test ./internal/platform/...` that the target holds either old or new content and no temp files remain after success or a failed write
- [ ] 2.2 Add the coded error type (code, path, params, nested problems) and the code constants per design, with a generator that writes the code list as a TypeScript union for the frontend; verify with a Go test that the generated file is up to date with the constants

## 3. Course source: normalization, storage key, copy

- [ ] 3.1 Add `source.NormalizeLocal` (Abs, EvalSymlinks, Clean, must be a directory) and the comparison form (lowercase on Windows, build-tagged); verify with tests for trailing separator, `..`, a symlinked folder, a file instead of a folder, and on Windows a case-only difference
- [ ] 3.2 Add `source.Key` per design (label rules, `-`, 8 hex of SHA-256 of the comparison form); verify with tests asserting literal keys for `C:\Courses\Go Basics`-style and `github.com/owner/repo` sources, an empty label falling back to `course`, a long label cut to 32 characters, and two folders with the same base name getting different keys
- [ ] 3.3 Add `source.CopyLocal` with go-git `gitignore` matching (only `.gitignore` files inside the folder), `.git` skipped, ignored directories not descended, symlinks skipped and logged; verify with tests for the spec scenarios: `bin/` ignored, nested `*.log` with `!keep.log`, `.git` absent, a non-git folder copied, global ignore patterns not applied

## 4. Course manifest

- [ ] 4.1 Add `schemas/course/v1.json` (types, required fields, `additionalProperties: false`, slug pattern, non-empty `sections` and `activities`) and `schemas/schemas.go` embedding it; add `santhosh-tekuri/jsonschema/v6` and `go.yaml.in/yaml/v3`; verify the schema compiles in a test
- [ ] 4.2 Add `manifest.ParseCourse(root)`: read `course.yaml`, `version` check before validation ("requires a newer app version"), JSON Schema validation, decode into `Course`, semantic checks (unique section slugs, course-wide unique activity slugs, `check` relative, no `..`, exists), all reported as coded problems with YAML paths; map every JSON Schema keyword used by the schema to a code; verify with tests asserting the code, path and params for each course-manifest scenario that applies to `course.yaml` (missing file, empty section, invalid slug, duplicate activity slug across sections, unknown field `autor` with `field: autor`, `check: ../outside.sh`, `version: 2`), a valid example, and a test that fails when the schema uses a keyword without a mapping
- [ ] 4.3 Add overview text lookup (`OVERVIEW.md` matched case-insensitively in the root, two matches a course error, fallback to `description`, neither gives no text); verify with tests for the overview-text scenarios of `course-library`

## 5. Library

- [ ] 5.1 Add `internal/library` with `library.json` (format version 1, `courses` array of `origin`, `source`, `key`, `title`) loaded and saved via `WriteFileAtomic` under a mutex, refusing to write a file with a higher format version; verify round-trip, a missing file read as an empty library, and a `version: 2` file left unchanged with an error in tests
- [ ] 5.2 Implement `Add(path)`: normalize, return the existing entry for a known source, fail on a key taken by another source, copy into `courses/<key>.new-*`, parse `course.yaml`, on error delete the copy and return the course error, otherwise rename to `courses/<key>` and save the entry with the title; verify with tests for a valid folder, a folder without a valid manifest (not in the library, no leftover folder), a duplicate add in another spelling, and the key-collision error
- [ ] 5.3 Implement `Sync(source)` with the `.new-*`/`.old-*` swap, no validation before the swap, title updated only when the new `course.yaml` is valid, copy untouched when the folder is gone; verify with tests for the Sync scenarios (edited overview, broken manifest keeps the old title, folder gone leaves the copy unchanged, a file written into the copy disappears and the author's folder is unchanged)
- [ ] 5.4 Implement startup recovery (restore the newest `.old-*` when `courses/<key>` is missing, delete leftover `.new-*`/`.old-*`, delete `<data>/smoke/`) with Windows rename retries; verify with tests that simulate an interrupted swap at each step and end with a complete old or new copy
- [ ] 5.5 Implement `List()` (stored titles, no course reads) and `Overview(source)` (title, origin, source, key, overview text or course error, computed from the copy on each call); verify with tests including a broken manifest returning the course error with the stored title

## 6. Course files for the webview

- [ ] 6.1 Add the asset handler for `GET /course-files/<key>/<path>` using `os.OpenRoot`, known keys only, image extensions only with content types; verify with `httptest` tests for a served PNG, `..` and absolute paths, a symlink escaping the copy, an unknown key and a non-image file all rejected
- [ ] 6.2 Add an image resolver option to `renderMarkdown` (relative `src` rewritten to `/course-files/<key>/...` relative to the course root; scheme, `//` and leading `/` rendered without `src`); verify with Vitest for a relative image, a nested relative path, an `https://` image and a `<script>` tag shown literally

## 7. App wiring and screens

- [ ] 7.1 Replace `Smoke*` methods in `app.go` with `ListCourses`, `PickFolder` (native folder dialog), `AddLocalCourse`, `SyncCourse`, `CourseOverview`; create the library on startup with recovery; plug the asset handler into `main.go`; regenerate `frontend/wailsjs`; verify `wails build` succeeds and `golangci-lint ./...` is clean
- [ ] 7.2 Add `en` and `ru` messages `errors.<code>` for every code and an error component rendering a coded error with its problems and paths; verify with a Vitest test that both locales have a message for every code in the generated code list
- [ ] 7.3 Remove `frontend/src/smoke/` and its i18n strings; add the view switch in `App.vue`, the library screen (list with titles, Add folder, empty state) and the add flow (folder picker → script warning → add → overview, coded error shown on failure) with `en` and `ru` strings; verify with `npm run lint`, `npm run typecheck`, `npm test`
- [ ] 7.4 Add the course overview screen: title, source, storage key, Sync with busy state and error, overview Markdown with images or plain-text description, course error; no Enroll, Continue or environment check controls; back to the library; verify the same frontend checks pass

## 8. Integration check

- [ ] 8.1 In the built Windows exe, add a sample local-folder course with `OVERVIEW.md`, an image, a `.gitignore`d folder and a `.git` folder; verify the library shows its title, the overview shows the source, key, text and image, `<data>/courses/<key>/` lacks the ignored folder and `.git`, an edit plus Sync is shown, a broken `course.yaml` plus Sync shows the course error with Sync still available, re-adding the same folder opens the existing course, and the state survives an app restart

## Workflow follow-up

- Archive the change after review; the archive applies the spec deltas to `course-source`, `course-library`, `learner-workspace` and `script-runtime`.
