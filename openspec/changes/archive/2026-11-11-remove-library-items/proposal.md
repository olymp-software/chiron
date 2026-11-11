# Proposal

## Why

The specs already promise that the learner can remove a course or workshop from the library (the only way to end an enrollment), but the app has no way to do it: added library items pile up, and a finished workshop keeps its workshop folder and course copy forever.

## What Changes

- Add a Remove button to the course overview of a course and of a workshop (only there, not in the library list).
- Add a confirmation dialog with two checkboxes, unchecked by default: "Keep progress" and "Keep my files".
- Removal stops the library item's running scripts, removes its library entry (with the remembered environment check result), and deletes its course copy, its `TEMP_DIR`s, and, unless kept, its progress and its learner folders or workshop folder.
- The library entry is removed first and atomically; files are deleted after it. If some files cannot be deleted (for example a folder open in an editor on Windows), the library item is still removed and the learner gets a warning listing the paths left on disk.
- After removal the library opens, and Back does not return to the removed library item's overview.
- On startup the app deletes course copies left without a library entry by an interrupted or partly failed removal.
- Clarify the "Removing a course" requirement: it applies to every library item, "clone" becomes "course copy", the workshop folder counts as the learner's files.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `library-source`: the "Removing a course" requirement gets where the learner starts removal, what opens after it, partial failure, and leftover cleanup on startup.

## Impact

- `internal/library`: new `Remove` (stops runs and the environment check, updates `library.json`, deletes files), `Recover` deletes orphaned course copies.
- `internal/progress`: new `Delete`.
- `app.go`: new bound method `RemoveCourse` returning the paths that could not be deleted; clears the course entered last if it was the removed one. Wails bindings regenerate.
- Frontend: `overview/CourseOverview.vue` gets the Remove button and a confirmation dialog; new i18n strings in `en.json` and `ru.json`.
- No changes to `library.json` or progress file formats.
