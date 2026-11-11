# Tasks

## 1. Backend: progress and file deletion

- [ ] 1.1 Add `progress.Store.Delete(key)` under the per-key lock, a missing file is not an error; verify with tests in `internal/progress` (deletes, missing file, `Exists` false after) via `task go:test -- -run Delete`
- [ ] 1.2 Extract `removeTemp`'s retry loop in `internal/library` into a helper that returns the error, keep `removeTemp` using it; verify existing tests pass with `task go:test`

## 2. Backend: Library.Remove

- [ ] 2.1 Implement `Library.Remove(key, keepProgress, keepFiles) ([]string, error)` per design decision 1: under `lockCourse`, stop the environment check and `Runs` of the key, remove the entry atomically, then delete the course copy, `<data>/temp/<key>`, progress unless kept and `<learner root>/<key>` unless kept, collecting the paths that could not be deleted; verify with tests: remove everything, keep progress, keep files, unknown key returns `CourseNotFound` and changes nothing, `checkHash` is gone after re-adding
- [ ] 2.2 Test that a running workshop control and a running environment check are stopped before deletion and that `TEMP_DIR` is gone; verify with `task go:test -- -run Remove`
- [ ] 2.3 Test re-adding the same source after removal with "Keep progress" (enrolled, progress picked up) and without (not enrolled); verify with `task go:test -- -run Remove`
- [ ] 2.4 Run the removal tests on Windows with a folder held open (a test that keeps a file open in the workshop folder) and check the path is returned in the left list and the entry is gone; verify with `task go:test:windows -- -run Remove`

## 3. Backend: Recover cleanup

- [ ] 3.1 Make `Recover` delete every entry in `<data>/courses` that is not the storage key of a library entry (folding in the `.new-`/`.old-` check), only after `List` succeeded; verify with tests: orphan copy deleted, copies of library items kept, Sync leftovers still deleted, unreadable `library.json` deletes nothing, kept progress file and learner folder are untouched

## 4. App binding

- [ ] 4.1 Add `RemoveResult{Left []string, Error *apperr.Error}` and `App.RemoveCourse(key, keepProgress, keepFiles)` in `app.go`, clearing `lastCourse` when it is the removed key; verify `task go:lint` passes and Wails bindings regenerate with `RemoveCourse` in `frontend/wailsjs`

## 5. Frontend

- [ ] 5.1 Add `overview/RemoveDialog.vue`: native `<dialog>` with the item title, "Keep progress" and "Keep my files" checkboxes unchecked, Cancel and Remove, buttons disabled while removing, and a leftovers state listing paths with an OK button; add `remove.*` strings to `en.json` and `ru.json` using "library item"-neutral wording (course and workshop variants where the text differs); verify with a vitest test for both states and the default checkbox values
- [ ] 5.2 Add the Remove button to `overview/CourseOverview.vue`, shown whenever the overview is loaded (also with a course error), disabled while syncing; on success `router.replace('/library')`, on `Error` show it like Sync/Enroll errors; verify with `CourseOverview.test.ts` cases: button present for course, workshop and broken manifest, `RemoveCourse` called with the checkbox values, `replace` used, error shown, leftovers shown before navigating
- [ ] 5.3 Run `task frontend:check` and verify it passes

## 6. Integration

- [ ] 6.1 Add a test workshop `tmp/workshops/03-remove` (a control that writes into the workshop folder and a long-running control) and checks in `tmp/workshops/CHECKLIST.md` and `tmp/courses/CHECKLIST.md` for: remove everything, keep progress + keep files then re-add, remove with a terminal open in the workshop folder (warning shown, copy gone after restart), Back after removal, removing the course entered last then restarting opens the library; verify by running them in `task run`
- [ ] 6.2 Run `task check` and verify it passes
