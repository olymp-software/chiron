# Tasks

## 1. Glossary

- [x] 1.1 In `openspec/config.yaml` replace the "Content" term with "Theory block" (an item of `theory_blocks`; every activity has at least one) and add "Block kind" (the form of a theory block, e.g. `markdown`; avoid "block type"); verify no other glossary entry still describes activity text as "content"

## 2. Verification

- [x] 2.1 Run `openspec validate theory-blocks --strict` and verify it passes
- [x] 2.2 Search `openspec/specs` and this change's delta specs for a top-level `content` field of `activity.yaml` with `rg -n "content" openspec/specs openspec/changes/theory-blocks/specs` and verify every remaining hit is a `markdown` block's `content` or generic English

## Workflow follow-up

- Archive the change to merge the delta specs into `openspec/specs`.
