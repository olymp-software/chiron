# Proposal

## Why

A course assumes independent activities, each with its own learner folder. This does not fit work where every activity builds on the result of the previous ones: a program grown across activities, resources in an AWS account, a local Kubernetes cluster. Such work needs its own content type, because the app does not own the state the learner is changing and cannot reset or restore it per activity.

## What Changes

- Add a second content type, the workshop: a repository with `workshop.yaml` in its root instead of `course.yaml`, with its own schema versioning. Which file is present defines the content type; both or neither is an error.
- A workshop has sections and activities like a course and is added, versioned, enrolled in and progressed through the same way. "Next" stays unconditional.
- An activity of a workshop has no activity type and no questions: it has `theory_blocks`, and optional `solution` (Markdown shown on request), `check`, `controls` and `files`.
- A workshop has one workshop folder, shared by all its activities and possibly empty. `files` are copied into it without overwriting. There is no Reset or Full reset, because the state may live outside the folder.
- An activity's `check` is a built-in Check button that runs like the environment check: a modal with Cancel, no kept output, no remembered result. It gates nothing.
- Scripts of a workshop's activities get `TEMP_DIR`, an empty folder outside the workshop folder for one run, and `CONTENT_DIR`, the root of the course copy, for helpers shared between activities.
- `workshop.yaml` may declare `controls` that run from the course overview, for example to look for leftovers before removing a workshop.
- Leaving an activity with a running control asks for confirmation.
- **BREAKING** The published activity schema moves from `<schema-base>/activity/v<N>.json` to `<schema-base>/course-activity/v<N>.json`. Authors must update the `yaml-language-server` comment. Parsing is unaffected, because the app uses the embedded schema.

## Capabilities

### New Capabilities

- `workshop`: the workshop content type: `workshop.yaml`, workshop activities, the workshop folder, `solution`, the Check button and the script environment.

### Modified Capabilities

- `course-manifest`: the repository layout allows `workshop.yaml` as an alternative to `course.yaml`, and the published schema paths change.

## Impact

- `schemas/`: `activity/` moves to `course-activity/`; new `workshop/` and `workshop-activity/` schemas.
- `internal/manifest`: a root manifest that is either type, and parsing of workshop activities.
- `internal/apperr` and both i18n files: new error codes.
- `internal/library`, `internal/runtime` and `app.go`: the workshop folder, copying `files`, running controls and Check of an activity.
- `frontend/src/course` and `frontend/src/overview`: the workshop activity view, the output panel on the activity screen, and a Check modal reusable outside the course overview.
- `openspec/config.yaml` and `CLAUDE.md`: glossary and architecture notes.
- Controls, the output panel on the activity screen, opening the external editor and its setting are specified for courses but not implemented yet. This change builds what workshops need; see design.md.
