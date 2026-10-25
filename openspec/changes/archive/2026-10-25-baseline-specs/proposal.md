# Proposal

## Why

We need a local desktop app for taking interactive IT courses: programming, DevOps, tooling, or any other topic where the learner reads material, answers questions and does hands-on work on their own machine. Authors publish courses as git repositories. The app downloads a course, renders its content and runs the course's scripts on the learner's machine. The app knows nothing about the subject matter: the course decides what its files mean and how work is checked.

This change establishes the baseline product contract: the course format, the app's behavior, and what the app guarantees to authors. Later changes build on it.

## What Changes

- Course repository format: the course manifest `course.yaml`, the activity file `activity.yaml`, and strict parsing rules.
- Three activity types: `theory`, `quiz`, `coding`. The type defines only the UI mode, not the topic, language or checking method.
- Control script contract: embedded shell interpreter, guaranteed commands, environment, output handling, stopping, and mutual exclusion of controls.
- Course-level environment check.
- A learner folder per activity, Reset and Full reset, opening in an external editor.
- Progression model: strictly sequential; only the "Next" button moves progress.
- Course versions as git tags, chosen explicitly by the learner, with compatibility rules for version switches.
- Platform and distribution for v1: Windows amd64/arm64, a single exe, local storage, no telemetry.

## Capabilities

### New Capabilities

- `course-source`: course source and its normalization, course versions, local-folder courses, adding, removing and listing courses, the course clone.
- `course-manifest`: course repository layout, `course.yaml`, slugs, common file rules, schema versions, JSON Schema, unsupported courses.
- `activity-content`: `activity.yaml`, activity types, content and Markdown rules, controls as an activity field.
- `quiz`: questions, the `choice` kind, options, feedback, explanations, answer evaluation.
- `script-runtime`: interpreter, guaranteed commands, run environment, output handling, exit code, Stop, exclusion groups.
- `environment-check`: the course environment check script, remembering a successful check by hash, blocking course start.
- `learner-workspace`: learner folder, workspace files, Reset and Full reset, external editor, relocating folders.
- `progress`: current, completed and new activities, completed course, progress behavior on course version switches.
- `app-platform`: release format, app update notices, UI languages, network use and local log.

### Modified Capabilities

_(none)_

## Impact

- New project on Go + Wails v2 + Vue 3.
- Dependencies: `mvdan.cc/sh` (+ `moreinterp/coreutils`), `go-git`, Shiki, `vue-i18n`.
- External systems: git hosts (public https repositories only), GitHub Releases (app update check), SignPath Foundation (exe signing in CI).
- JSON Schemas for `course.yaml` and `activity.yaml` are published. The `<schema-base>` URL is chosen together with the final app name.
