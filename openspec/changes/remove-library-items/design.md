# Design

## Context

Everything the app stores for a library item is named by its storage key:

| Data | Location | Removal |
|---|---|---|
| Library entry and remembered check result (`checkHash`) | `<data>/library.json` | always |
| Course copy | `<data>/courses/<key>` | always |
| `TEMP_DIR`s of running scripts | `<data>/temp/<key>/...` | always |
| Progress (= enrollment) | `<data>/state/<key>.json` | unless "Keep progress" |
| Learner folders / workshop folder | `<learner root>/<key>` | unless "Keep my files" |
| Course entered last | settings `lastCourse` | cleared if it is the key |

Only local-folder courses exist today, so the course copy is never a git clone yet. Course learner folders are not implemented yet either, but they will live under `<learner root>/<key>` like the workshop folder, so one path covers both.

`Library` already has a per-storage-key lock (`lockCourse`), held by add and Sync and by every script start, plus `StopCheck` and `Runs.stopCourse`, which Sync uses to release the course copy before swapping it. `Recover` runs once on startup before anything else: it deletes the whole `<data>/temp` and the `.new-`/`.old-` leftovers of interrupted Syncs.

## Goals / Non-Goals

**Goals:**
- Removal is never half-done from the learner's point of view: the library item is either in the library or gone.
- Files that cannot be deleted are reported, not hidden, and app-owned ones are cleaned up later.

**Non-Goals:**
- Removing several library items at once.
- Cleaning up progress or learner folders of items that are no longer in the library (they may have been kept on purpose).
- Undo.

## Decisions

### 1. Entry first, files after

`Library.Remove(key, keepProgress, keepFiles)` does, under `lockCourse(key)`:

1. `StopCheck(key)` and `runs.stopCourse(key)`. They wait until the scripts have exited, so their working directories and `TEMP_DIR`s are released before deletion.
2. Remove the entry from `library.json` in one atomic `update`. This is the point of no return; if it fails, nothing else happens and the error is returned.
3. Delete the course copy, `<data>/temp/<key>`, the progress file unless kept, and `<learner root>/<key>` unless kept. Each failure is logged and its path is collected.

It returns the collected paths. Holding the course lock for the whole operation keeps a script start, a Sync, or a re-add of the same source (same storage key) from interleaving.

*Alternative:* delete files first and the entry last, so a failure leaves the item in the library to retry. Rejected: it leaves an item without a course copy or without progress in the library, which every other screen would then have to handle.

### 2. Partial failure is a result, not an error

`App.RemoveCourse(key, keepProgress, keepFiles)` returns `RemoveResult{Left []string, Error *apperr.Error}`. `Error` is only set when nothing was removed (entry not found, `library.json` unreadable or unwritable). `Left` holds absolute paths. Paths are data, not user-facing text, so no new `apperr` code is needed for the warning; the frontend owns the sentence around them.

Deleting a folder retries like `removeTemp` already does (`renameAttempts` × `renameDelay`, one attempt off Windows), because on Windows a program that has just exited, an antivirus or an indexer often holds a file for a moment. `removeTemp`'s loop becomes a shared helper that returns the error instead of only logging it.

`StopCheck` and `runs.stopCourse` do not take the course lock (they wait on the runs' `finished` channels, and a stopped check does not write `library.json`), so calling them while holding it cannot deadlock.

### 3. Progress deletion lives in `progress.Store`

New `Store.Delete(key)` removes `<key>.json` under the store's per-key lock, so it cannot race an `Update`. A missing file is not an error.

### 4. Startup cleanup of orphaned copies

`Recover` gets one more step: after repairing copies, it deletes every entry in `<data>/courses` whose name is not a storage key in `library.json`. That covers the existing `.new-`/`.old-` leftovers too, so the suffix check can be folded into it. `<data>/temp` is already deleted as a whole. Progress files and learner folders are not touched (see Non-Goals). If `library.json` cannot be read, `Recover` already skips copy repair, and it must skip this step as well, or it would delete every copy.

### 5. Course entered last

`App.RemoveCourse` clears `lastCourse` when it equals the removed key. `StartCourse` already ignores a key that is not enrolled, but a course removed with "Keep progress" and re-added would otherwise be entered on startup unexpectedly.

### 6. Frontend: the dialog reports leftovers before leaving

A new `overview/RemoveDialog.vue` (native `<dialog>`, like `library/ScriptWarning.vue`) shows the title of the library item, the two checkboxes, Cancel and Remove. On Remove it calls `RemoveCourse`:

- `Error` set: the dialog closes and the overview shows the error as it does for Sync and Enroll.
- `Left` empty: `router.replace('/library')`.
- `Left` not empty: the dialog switches to a warning listing the paths with a single OK button, and OK does `router.replace('/library')`.

Showing the warning inside the dialog avoids carrying state across navigation into the library. `replace` keeps the removed overview out of history (spec: Back after removal). While removal runs, the dialog buttons are disabled.

The Remove button is shown whenever the overview has loaded, including with a course error, and is disabled while Sync runs. The check and control modals are modal, so removal cannot start while one of them is open.

## Risks / Trade-offs

- [A progress file that cannot be deleted silently re-enrolls the learner if the same source is added again] → The path is in the warning; `Recover` cannot tell such a file from kept progress, so it is not cleaned automatically.
- [The workshop folder may hold state the app does not own, such as containers or cloud resources, that removing the folder does not clean up] → Out of scope; authors can provide a workshop control such as "Check leftovers" in the overview, which the learner can run before Remove.
- [`Recover` deleting by "not in library.json" could wipe all copies if it ever ran with a wrong library] → It runs only after a successful `List`; covered by a test with an unreadable `library.json`.
- [Deleting `<learner root>/<key>` follows the storage key, not anything the learner picked] → Storage keys contain only `[a-z0-9-]`, so the path cannot escape the learner root.
