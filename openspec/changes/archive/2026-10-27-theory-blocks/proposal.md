# Proposal

## Why

An activity's main text is a single top-level Markdown string, `content`. Some things the app may later want to show (interactive or specially rendered material) are hard or impossible to express in Markdown, and a single string leaves no room to add them. Replacing the string with a list of typed blocks lets the app render each block by its kind and add new kinds later without reshaping `activity.yaml` again. Nothing is released yet, so the format can change now at no cost.

## What Changes

- **BREAKING** Remove the top-level `content` field from `activity.yaml`.
- Add a required top-level `theory_blocks` field: a list of at least one theory block, shown in order, for every activity type (`theory`, `quiz`, `coding`).
- A theory block has `kind` (required) and `data` (object whose shape depends on `kind`, required), mirroring quiz questions.
- The only block kind is `markdown`; its `data` has `content` (Markdown, required, non-empty).
- An unknown block `kind` is handled by the existing unsupported-course rule: the activity shows "requires a newer app version". Adding a new block kind bumps the activity schema version.
- Quiz text fields (`prompt`, `explanation`, option `text`, `feedback`) stay plain Markdown strings.
- The activity schema version stays 1: v1 is redefined in place, since no app or course using the old `content` field has been released.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `activity-content`: `activity.yaml` fields replace `content` with `theory_blocks`; new requirements for theory blocks and the `markdown` block kind; activity type requirements, Markdown dialect and Markdown-related scenarios refer to theory blocks instead of `content`.
- `course-manifest`: the unsupported-courses requirement gains a scenario for an unknown theory block kind.
- `course-library`: overview file rendering refers to activity Markdown instead of activity content.

## Impact

- Specs only; no code parses `activity.yaml` yet.
- The published activity JSON Schema v1 (when it is written) must describe `theory_blocks` instead of `content`.
- The domain glossary in `openspec/config.yaml` defines "Content" as an activity's main Markdown text; it needs a "Theory block" / "Block kind" entry instead.
