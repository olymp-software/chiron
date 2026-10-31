# Proposal

## Why

A Learner can enroll in a course and open its activities, but an activity shows only its title from `course.yaml` and nothing moves progress: `activity.yaml` is never read, there is no Next, and every activity is open. With activity content and sequential progression a theory-only course can be taken from the first activity to "Finish course", for the first time. Controls, quizzes and coding activities build on this screen.

## What Changes

- `activity.yaml` is read when the Learner opens an activity, validated against a new JSON Schema `schemas/activity/v1.json` (published for authors, like the course schema), then semantic checks. The schema covers the whole activity schema version 1: `type`, `theory_blocks`, `controls`, and the `data` of `quiz` and `coding`, because any later schema change would bump the activity schema version.
- An `activity.yaml` with a higher `version`, an unknown `type` or an unknown theory block `kind` is reported as "requires a newer app version" before schema validation.
- The activity view shows the theory blocks in order, `markdown` blocks rendered with the existing Markdown renderer. Images resolve relative to the activity folder and are not loaded if they leave it.
- A broken activity (missing folder or `activity.yaml`, invalid YAML, schema or semantic problems) shows an activity error with the problems and their YAML paths; the course tree stays usable.
- Next on every activity except the last, "Finish course" on the last. Next marks the activity completed, advances the current activity if it was the current one, and opens the following activity. Finish course marks the last activity completed and shows the course as completed. Both are available on an activity with an error.
- Activities after the current activity are locked: the course tree shows them as unavailable and Go refuses to open them. The tree marks completed activities and the current activity.
- The library marks a completed course as completed instead of enrolled. Whether a course is completed is stored in its progress, so the library still does not read course copies.
- The course overview shows that a course is completed and offers Review instead of Continue. The current activity becomes the first activity that is not completed, and a course is completed when every activity is: when a Sync or version switch adds activities anywhere in a completed course, it stops being completed and the first added activity becomes current. Completed activities after the current one stay available. This replaces the "new" badge rule: an activity inserted before the current one becomes current instead of being skipped.
- Not in this change: controls (declared and validated, not shown), quiz questions and their evaluation (validated, not shown), and the coding activity's workspace, learner folder, external editor button and output panel (validated, not shown).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `activity-content`: what an activity error shows; images in activity Markdown that leave the activity folder are not loaded.
- `progress`: Next and Finish course are available on an activity with an error; the course tree marks completed, current and locked activities, and a completed course is shown on its last activity.
- `course-library`: a library entry and the course overview show whether the course is completed; the overview offers Review instead of Continue for a completed course.
- `progress`: the current activity is the first activity that is not completed; completed activities stay available; a course is completed when every activity is completed; an inserted activity becomes current instead of getting a "new" badge; a removed current activity moves to the first activity that is not completed.

## Impact

- Code: `internal/manifest` gets `ParseActivity` (pre-checks, schema, semantic checks: correct option count, control script and workspace paths inside the activity folder); `schemas/` embeds `activity/v1.json`; `internal/library` gets reading an activity of an enrolled course with the lock check, completing an activity, and completion and lock state in the course tree; `app.go` binds `Activity` and `CompleteActivity`; `internal/progress` stores `courseCompleted`, recomputed whenever the progress is reconciled with `course.yaml` (Enter, Next / Finish course, Sync); the library list reads it with the enrollment.
- `internal/apperr`: new codes for activity errors (missing activity file, invalid `activity.yaml`, its YAML syntax and unsupported schema version, type or kind, locked activity, wrong correct option count), with `en.json`/`ru.json` messages.
- Frontend: `ActivityView` loads and renders the activity, shows the activity error, Next / Finish course and the completed course; `CourseTree` shows completed, current and locked activities; `LibraryView` shows Completed instead of Enrolled for a completed course; an activity image resolver confined to the activity folder; i18n strings.
- Data folder: `state/<key>.json` starts receiving `completed` slugs and gains a `courseCompleted` flag; the format version stays 1, since older files without the flag read as not completed.
