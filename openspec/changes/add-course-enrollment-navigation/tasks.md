# Tasks

## 1. Error codes

- [ ] 1.1 Add `notEnrolled`, `progressReadFailed`, `progressWriteFailed` to `internal/apperr` and `All`, run `go generate ./internal/apperr`, add `errors.<code>` messages to `en.json` and `ru.json`; verify with `task go:test` and `task frontend:check`

## 2. Progress store

- [ ] 2.1 Create `internal/progress`: `state/<key>.json` (`version`, `current`, `completed`) read and written atomically, newer `version` read-only, per-key lock; `Exists`, `Load`, `Enroll` (no-op on an existing file); verify with tests for enroll, repeated enroll, missing file and newer format via `task go:test`
- [ ] 2.2 Implement `Resolve(order, progress)` per design (current kept, moved after the last completed one, first when nothing completed is left, last when the last completed one is last); verify with table tests via `task go:test`

## 3. Settings store

- [ ] 3.1 Create `internal/settings`: `settings.json` (`version`, `lastCourse`), read-modify-write under one mutex, atomic writes, newer `version` read-only, read errors treated as empty settings; verify with tests for missing, corrupt and newer files via `task go:test`

## 4. Library by storage key and enrollment

- [ ] 4.1 Switch `library` lookups from course source to storage key (`entry`, `Sync`, `Overview`) and update existing tests; verify with `task go:test`
- [ ] 4.2 Create the progress store in `library.Open` for `<data>/state`; add the `enrolled` flag to list items and `Overview` without changing the persisted `Entry`; verify with tests that library.json is unchanged and the flag follows the state file via `task go:test`
- [ ] 4.3 Add `Enroll(key)` (valid `course.yaml` required, current = first activity), `Enter(key)` (in library, enrolled, valid `course.yaml`, resolve and save the pointer, return the current slug; marked place for the environment check) and `Tree(key)` (title, current, sections with activity slugs and titles); verify with tests for an unenrolled course, an invalid manifest, a removed current activity and a repeated Enroll via `task go:test`

## 5. Bound API

- [ ] 5.1 In `app.go` and `main.go`: open the settings store; switch `CourseOverview` and `SyncCourse` to the key; add `EnrollCourse`, `EnterCourse` (stores `lastCourse`, logs a settings failure), `CourseTree`, `StartCourse` (returns `lastCourse` only if it is in the library and enrolled); verify with `task go:lint`, `task go:lint:internal` and `task go:test`

## 6. Router

- [ ] 6.1 Add `vue-router` to `frontend/package.json` and create `frontend/src/router.ts` with hash history and the route table from design, components receiving `key`/`slug` as props; `App.vue` renders `<RouterView>`; verify `task frontend:check` passes
- [ ] 6.2 Implement the `/` startup redirect via `StartCourse` and the global entering guard calling `EnterCourse` when the course key changes, redirecting to `/library/:key` on error and `/courses/:key` to the current activity; verify with vitest tests of the guard with the Wails bindings mocked (key change, same key, error, startup with and without `lastCourse`) via `task frontend:test`
- [ ] 6.3 Handle Back input (mouse button 3, Alt+ArrowLeft) with `preventDefault` and `router.back()`; verify with a vitest test dispatching both events via `task frontend:test`

## 7. Screens

- [ ] 7.1 `LibraryView`: navigate with `router.push('/library/<key>')` on click and after Add; show the enrolled mark; add i18n strings in `en.json` and `ru.json`; verify `task frontend:check`
- [ ] 7.2 `CourseOverview`: load by `key`; show Enroll or Continue (unavailable when the course error is shown); Enroll calls `EnrollCourse` then pushes `/courses/<key>`, Continue pushes `/courses/<key>`; back button uses `router.back()` or pushes `/library` when there is no previous entry; i18n strings; verify `task frontend:check`
- [ ] 7.3 Create `course/CourseScreen.vue` (header with Library, dropdown of enrolled courses pushing `/courses/<key>`, Overview; keyed by `key`), `course/CourseTree.vue` (sections and activity `RouterLink`s, open activity highlighted) and `course/ActivityView.vue` (activity title; unknown slug replaced with the current activity); i18n strings; verify `task frontend:check`

## 8. Integration

- [ ] 8.1 Run `task check` and verify it passes
- [ ] 8.2 Build with `task build`, run the exe and verify by hand: add a course, Enroll, open activities from the tree, Back with the mouse button and Alt+Left goes back exactly one entry, switch courses in the dropdown, Library and Overview buttons, Sync with a broken `course.yaml` then pick the course in the dropdown opens the overview, restart reopens the last course
