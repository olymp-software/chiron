# Proposal

## Why

Today adding a course implicitly starts it: the course opens, the environment check gates its activities on the course screen, and the course description is a one-line plain-text string. A learner cannot keep courses in the library without taking them, and an author has no place to explain what a course covers and what the learner must install before starting. Fixing the model now, before the library and course screen are built, lets the implementation rely on it.

## What Changes

- Introduce **enrollment**: the learner's decision to take a course from the library. Adding a course no longer starts it; the learner enrolls explicitly. There is no unenroll; enrollment ends only when the course is removed.
- Introduce the **library** as the list of all added courses and the **course overview**, a window opened from the library that shows the course's overview text, the environment check, the version selector and an Enroll / Continue button.
- Add `OVERVIEW.md` support: a Markdown file in the course root, found case-insensitively. The overview shows `OVERVIEW.md` if present, otherwise `description`, otherwise nothing. `course.yaml` is unchanged, so the course schema stays at version 1.
- **BREAKING** (spec-level): the environment check gates Enroll and Continue in the course overview instead of gating activities on the course screen. The "Check environment" button moves to the course overview.
- **BREAKING** (spec-level): the selected version is changed only in the course overview; the library only marks courses that are not on the latest version.
- Entering a course (enrolled-courses dropdown on the course screen, app start, Reload of a local-folder course) opens the course screen only if the check is satisfied for the current `check` hash; otherwise it opens the course overview.
- On startup the app opens the last entered enrolled course, or the library if the learner has no enrollments. After adding a course the app opens its course overview.
- Removing a course also removes its enrollment; "Keep progress" keeps the enrollment and the remembered check result.
- Update the domain glossary in `openspec/config.yaml` with Enrollment, Library, Course overview and Overview file, and adjust Description and Environment check.

## Capabilities

### New Capabilities
- `course-library`: the library, the course overview, enrollment, the enrolled-courses dropdown, and which screen the app opens on startup, after adding a course and on entering a course.

### Modified Capabilities
- `course-manifest`: repository layout gains the optional `OVERVIEW.md`; the README rule and the `description` field are clarified.
- `course-source`: version selection moves to the course overview; the "Course library" requirement moves to `course-library`; adding and removing a course account for enrollment.
- `environment-check`: the check gates Enroll and Continue in the course overview instead of activities on the course screen.
- `progress`: progress exists only for an enrolled course and starts at Enroll; entering a course respects the environment check.

## Impact

- Specs: `course-manifest`, `course-source`, `environment-check`, `progress`, new `course-library`.
- Glossary in `openspec/config.yaml`.
- Code: nothing implemented yet conflicts. Future work on the library, course screen and environment check (`internal/source`, the Vue frontend, persisted library and progress state) builds on these specs. The overview reuses the existing Markdown renderer in `frontend/src/markdown/markdown.ts`.
- Course authors: optional `OVERVIEW.md`; no `course.yaml` schema change. Older app versions ignore the file.
