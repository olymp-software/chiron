# Design

## Context

See proposal.md - Why. Today every activity type renders the top-level `content` string above its type-specific UI. Quiz questions already use a `kind` + `data` shape for a polymorphic item. No code parses `activity.yaml` yet and no activity JSON Schema is published.

## Goals / Non-Goals

**Goals:**
- A polymorphic block list that new kinds can join without changing existing kinds or the top-level shape.
- One block shape for authors to learn, consistent with quiz questions.

**Non-Goals:**
- Block kinds other than `markdown`.
- Blocks inside quiz text fields.
- Placing blocks anywhere other than the top of the activity.

## Decisions

### `theory_blocks` stays at the top level
All three activity types render it, which matches the existing rule that common fields live at the top level and type-specific ones under `data`.
- Alternative: under `data` per type. Rejected: the same field would be repeated in every type's schema, and `theory` would need a `data` field (today forbidden) only for this.

### Name `theory_blocks`
Says what the blocks are for; `content` moves into the `markdown` block's `data`, where it names the Markdown text itself.
- Alternatives: `blocks` (unclear what they hold), `body`, `theory` (clashes with `type: theory`).
- It is the first multi-word field; multi-word fields use snake_case from now on.

### Block shape `kind` + `data`
Mirrors quiz questions (`kind` + kind-specific `data`). Future kinds have unknown schemas; nesting them under `data` keeps the block's own fields (`kind`, and any later common ones) apart from kind-specific ones.
- Alternative: flat (`kind: markdown`, `content: ...`). Shorter to write, but common and kind-specific fields would mix, and it would differ from questions.
- `kind`, not `type`: `type` is reserved for activities.

### Unknown kind uses the existing unsupported-course rule
"Requires a newer app version" instead of a generic activity error or skipping the block. Skipping would silently hide part of the theory. A new kind is a schema change and bumps the activity schema version, so an app that supports that version knows the kind.

### Redefine activity schema v1 in place
No app release and no course uses `content`, so there is no v1 to keep supporting. A v2 would force the app to parse both shapes forever.

## Risks / Trade-offs

- [More YAML per activity for the common single-Markdown case] → Accepted for a uniform shape; JSON Schema completion helps authors.
- [`theory_blocks` on a `coding` or `quiz` activity reads oddly when the text is a task statement rather than theory] → Accepted; the name stays as decided.
