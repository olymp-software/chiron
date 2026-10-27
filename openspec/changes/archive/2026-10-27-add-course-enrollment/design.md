# Design

## Context

The app is at the skeleton stage: `internal/source` only clones a repository and lists tags, and the frontend has a smoke screen plus a Markdown renderer (`frontend/src/markdown/markdown.ts`). The library, the course screen, progress storage and the environment check are not implemented yet, so this change reshapes the specs before code depends on them. See proposal.md for motivation.

## Goals / Non-Goals

**Goals:**
- One place, the course overview, where the learner reads about a course, checks the environment, picks a version and enrolls or continues.
- A single rule for when the course screen may open, covering the dropdown, app start and Reload.

**Non-Goals:**
- Visual layout of the library and the course overview.
- Persisted file formats for the library and progress; they are decided when those features are implemented.
- A course catalog or search.

## Decisions

### Overview file found by name, not declared in course.yaml
`OVERVIEW.md` in the course root is picked up by convention, like README on GitHub. Alternatives: a path field in `course.yaml` (needs a course schema bump to v2 for little gain) or making `description` Markdown (loses the short plain-text summary). The convention keeps `course.yaml` at schema version 1, and older app versions simply ignore the file.

### Case-insensitive lookup, ambiguity is an error
The app lists the course root and matches the name case-insensitively, so a course authored on Windows with `overview.md` behaves the same on Linux. Two matching files can exist only on case-sensitive file systems; the app reports a course error instead of picking one silently, consistent with strict parsing elsewhere.

### Enrollment is the existence of progress
A course is enrolled exactly when it has a progress record; Enroll creates an empty one. This needs no separate flag and makes "Keep progress" on removal keep the enrollment automatically. Alternative: an `enrolled` flag in the library entry, which could disagree with progress after removal and re-adding.

### Check status derived from the stored hash
The remembered successful check (SHA-256 of the `check` file) already exists in the environment-check spec. "Check required" is derived by comparing it with the current file's hash on opening the overview and on entering a course, so no extra "check pending" state is persisted and the requirement survives restarts by construction.

### Entering a course is one guarded operation
The dropdown, app start and Reload of a local-folder course all go through the same entry: if the check is not satisfied, the app opens the course overview instead of the course screen. Version switching lives only in the overview, so the common hash-change case already lands there; the guard covers edits on disk and Reload.

## Risks / Trade-offs

- [The learner cannot see a course's activities before enrolling] → The overview text is the author's place to describe the course; a read-only course tree in the overview can be added later without changing enrollment.
- [Without unenroll, the dropdown can only shrink by removing a course] → Acceptable for v1; unenroll can be added later as an explicit operation if needed.
- [`OVERVIEW.md` is not shown on the GitHub repository page as prominently as README] → Authors can link it from README; the app deliberately ignores README because it targets GitHub readers.
