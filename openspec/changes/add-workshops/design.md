# Design

## Context

See proposal.md for motivation and specs/workshop/spec.md for behavior.

`manifest.ParseCourse` is called from `internal/library` wherever a course is opened, and `library`, `progress` and the tree only read sections and activities from its result. `manifest.ParseActivity` handles one schema with `type` and `data`. Environment check runs go through `library.StartCheck`, which remembers results by hash, and `CheckModal.vue` with `checkRun.ts`, both bound to the course-level check.

Not implemented today, although specified for courses: running controls, the output panel on the activity screen, `coding` and `quiz` rendering, learner folders, opening the external editor and its setting, relocating the root, removing a course, repository courses. The activity view renders theory blocks and "Next".

## Goals / Non-Goals

**Goals:**
- Add workshops with the smallest change to library, progress and the tree.
- Build controls, the output panel and the folder opener so the course work can reuse them.

**Non-Goals:**
- Controls, `coding` and `quiz` activities in courses.
- The editor choice in settings, `{dir}` commands, relocating the root.
- Repository sources: workshops get them when courses do.
- Questions, hints, snapshots, reference states, timeouts, learner input to scripts.

## Decisions

### One parsed manifest with a content type

`manifest.Parse(root)` replaces `ParseCourse`: it detects the file, validates against its schema and returns `Course` with a new `Type` field (`course` or `workshop`). `workshop.yaml` has the same fields, so sections, the tree, progress and the check keep working unchanged. `ManifestMissing` keeps its code with an updated message; a new `ManifestConflict` covers both files.

Alternative: a `Content` interface with separate `Course` and `Workshop` types. Rejected: nothing differs in these fields today, and the schemas are versioned separately, so the types can split when a schema diverges.

### workshop.yaml and activities have their own schemas

Files are `schemas/workshop/v1.json` and `schemas/workshop-activity/v1.json`. `workshop/v1.json` starts as a copy of the course schema. The file stays `activity.yaml`; the root manifest decides which schema reads it, so `ParseWorkshopActivity` is a separate function. `theory_blocks` definitions are copied, keeping each published schema a self-contained file, like `course/v1.json`.

```yaml
# activity.yaml of a workshop
version: 1
theory_blocks:
  - kind: markdown
    data:
      content: "# Rollback"
solution: |
  ```sh
  kubectl rollout undo deployment/web
  ```
check: check.sh
controls:
  - label: Start cluster
    script: up.sh
files:
  - path: manifests
```

The library returns a `Workshop` block on `Activity` with `solution`, `hasCheck`, `controls`, and `Type` stays empty for workshop activities. The frontend switches on it.

### One workshop folder, reused name

The folder is `<learner folder root>/<storage key>`, with the root `~/<App>/`, and `WORKSPACE_DIR` keeps its name, so scripts look the same as in a course. `workshop.yaml` declares nothing about the folder: the first activity's `files` provide any initial files, and a workshop may have no files at all. Files are merged one by one, because several activities may fill one folder.

Alternatives rejected: a folder chosen at Enroll needs rules for non-empty and vanished folders and can be added later without breaking the format. Per-activity service folders are replaced by `TEMP_DIR`.

Each file is written to a temporary name in the same folder, then renamed, so an interrupted copy leaves no partial file at its final path. Temporary names are removed on the next open of the activity.

### CONTENT_DIR names the root of the course copy

Scripts stay declared inside their activity folder, but shared helpers are reached through `CONTENT_DIR` instead of relative paths that depend on nesting. The name is neutral to the content type, so courses can get the same variable later. `WORKSHOP_DIR` was rejected as too close to `WORKSPACE_DIR`.

### TEMP_DIR belongs to the app

Each run gets `<data>/temp/<storage key>/<run id>`. The app deletes it after the run and removes all of `<data>/temp` on start. A script cannot clean up after Stop, so the app owns this. A failed delete is logged and left to the next start.

### Runs are per activity, keyed by slug

A new `library.Runs` holds the running scripts of the open activity, keyed by control index, plus at most one Check. `Sync`, `StopAll` and leaving the activity stop them. Exclusion is enforced in Go, so a double click cannot start a second run. Output uses the existing `runtime.Batcher` and Wails events named by run id, as the environment check does.

Each control keeps its own `OutputPanel` while the activity is open, selected by tabs, so parallel output is never mixed.

Leaving goes through a vue-router guard on the activity route: with a run in progress it asks for confirmation, and only then calls `StopActivityRuns`. "Next" uses the same path before `CompleteActivity`. The Wails side also stops runs on app shutdown through `StopAll`.

`runtime.Run` already kills what a script leaves behind when it exits, so "Processes end with their script" needs tests, not new code. `TEMP_DIR` is deleted after `Wait` returns.

### Workshop controls run in the check modal

`workshop.yaml` `controls` have `label` and `script` only: a modal runs one script at a time, so `concurrency` has no meaning there. They use the generalized `CheckModal.vue` and a `StartWorkshopControl` that remembers nothing. They appear in the overview and are not gated by the environment check, because they matter most when the learner is leaving.

### Tools on PATH are installed before the workshop

Scripts keep getting the environment the app started with. Tools a workshop needs are installed before starting it and verified by the environment check; an activity does not install them. After installing, the learner may have to restart the app, so the check output should say that.

### Check reuses the modal, not the check service

`library.StartCheck` remembers hashes and gates the course, which must not apply here, so Check gets `StartActivityCheck`, sharing `runtime.Start`. `CheckModal.vue` and `checkRun.ts` are generalized to take `start` and `cancel` functions; the title comes from a prop. The modal itself blocks control buttons, so no extra state is needed.

### External editor falls back to the file manager

There is no editor setting yet, and the spec says the button opens the file manager until an editor is chosen. This change implements exactly that path through a new `App.OpenFolder`. The setting will arrive with the course workspace work and needs no change here.

### Rename the activity schema in place

`schemas/activity/` moves to `schemas/course-activity/`, updating `$id`, `schemas.go` and `internal/manifest/schema.go`. Content and version stay the same. No redirect is kept at the old URL.

## Risks / Trade-offs

- [The workshop screen shares `CourseScreen.vue`] → Branch only on `Type` for the editor button and the finish label; keep workshop parts in a separate `WorkshopActivity.vue`.
- [This change builds controls for workshops first] → Keep `Runs` and the output panel free of workshop concepts, so course controls only add their view.
- [`files` reappear when an earlier activity is reopened after the learner deleted them] → Accepted, same as course workspace items.
- [Check reports stale state after a version switch] → Accepted: the learner chose the switch.
- [Old schema URL stops working] → Only editor completion is affected; note it in the release notes.

## Migration Plan

No stored data changes. Existing courses keep working because `course.yaml` parsing is unchanged. Rollback is reverting the change; workshops added meanwhile would show a course error.
