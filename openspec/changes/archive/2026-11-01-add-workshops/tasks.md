# Tasks

## 1. Schemas and glossary

- [x] 1.1 Move `schemas/activity/` to `schemas/course-activity/`, update `$id`, `schemas/schemas.go` and `internal/manifest/schema.go`; verify with `task go:test`
- [x] 1.2 Update the course test fixtures and `tmp/courses` that use the old schema URL; verify `rg 'schemas/activity|/activity/v1' internal schemas frontend/src tmp/courses` finds nothing
- [x] 1.3 Add `schemas/workshop/v1.json` and `schemas/workshop-activity/v1.json` and embed them; verify with schema tests for valid files, `type`, `data` and an empty `solution`
- [x] 1.4 Add Workshop, Workshop folder, Solution and Activity check to the glossary in `openspec/config.yaml`; verify `openspec validate add-workshops --strict`

## 2. Manifest

- [x] 2.1 Replace `ParseCourse` with `manifest.Parse` that detects the root file and sets `Course.Type`; verify tests for course, workshop, both, neither and a newer `workshop.yaml` version
- [x] 2.2 Add the `ManifestConflict` code, run `go generate ./internal/apperr`, add `errors.manifestConflict` and update `errors.manifestMissing` in `en.json` and `ru.json`; verify the apperr and i18n drift tests with `task check`
- [x] 2.3 Add `ParseWorkshopActivity` with path checks for `check`, `controls` and `files`, and newer-version handling; verify tests for each activity error scenario in the workshop spec
- [x] 2.4 Parse `controls` of `workshop.yaml` with path checks from the workshop root; verify tests for a valid control, `../leftovers.sh` and `controls` rejected in `course.yaml`

## 3. Library and scripts

- [x] 3.1 Return the content type on `Entry`, `Overview` and the tree, and the workshop fields on `Activity`, choosing the parser by type; verify with `internal/library` tests
- [x] 3.2 Add the workshop folder and exact, atomic per-file copying of `files` on activity open; verify tests for the empty folder, a deleted folder, kept edits, a missing file, two activities merging one folder, identical binary content and a folder where a file is expected
- [x] 3.3 Add `TEMP_DIR` creation, deletion after a run and cleanup on start, and pass `CONTENT_DIR`; verify tests for an empty folder on every run, leftovers after a stopped run, a helper loaded through `CONTENT_DIR` and neither variable in the environment check
- [x] 3.4 Add `library.Runs` for controls with exclusion groups and Stop, stopped by leaving the activity, Sync and `StopAll`; verify tests for the default group, a separate group and stopping on Sync
- [x] 3.5 Add `StartActivityCheck` running beside controls with no remembered result; verify tests that it runs while a control runs, leaves `library.json` unchanged and that a background process is gone before the exit code is reported
- [x] 3.5a Add `StartWorkshopControl` for `workshop.yaml` controls with the workshop root as working directory; verify tests for the environment variables, no remembered result and running with an unsatisfied environment check
- [x] 3.6 Bind runs, Check and `OpenFolder` in `app.go` with `*Result` structs, and stop runs on shutdown; verify `task go:lint` and `task go:lint:internal`, and that `OpenFolder` opens the folder in the file manager on Windows via `task run`

## 4. Frontend

- [x] 4.1 Generalize `CheckModal.vue` and `checkRun.ts` to take start and cancel functions; verify the existing `checkRun` and `CourseOverview` tests still pass
- [x] 4.2 Add `WorkshopActivity.vue` with theory blocks, Solution, the Check button, controls and one output per control; verify vitest for hidden solution, no Check button without `check`, disabled controls in one group, switching outputs and output cleared on restart
- [x] 4.3 Add the leave confirmation for "Next", the course tree, Back, Library and the dropdown; verify vitest for confirmed, declined and nothing running
- [x] 4.4 Add the workshop mark in the library and overview, the folder path with a copy button and the editor button on the course screen, workshop controls in the overview and "Finish workshop", with `en` and `ru` strings; verify vitest and `task frontend:check`
- [x] 4.5 Split the library into "Courses" and "Workshops" sections instead of the workshop mark, and make the add button, folder dialog, empty text and script warning apply to both, with `en` and `ru` strings; add Library item to the glossary; verify vitest for both sections and a hidden empty one, `openspec validate add-workshops --strict` and `task frontend:check`

## 5. Documentation and integration

- [x] 5.1 Update the architecture section of `CLAUDE.md` for workshops and `Runs`; verify the commands and paths it names exist
- [x] 5.2 Add test workshops `tmp/workshops/01-workshop` (shared folder, `files`, `solution`, Check, a server control) and `02-both-manifests`, with checks in `tmp/workshops/CHECKLIST.md`; verify by walking the checklist in the app via `task run`
- [x] 5.3 Run `task check` and `task go:test:windows`; verify both pass

## Workflow follow-up

- Archive the change after review, and verify the workshop and course-manifest main specs.
