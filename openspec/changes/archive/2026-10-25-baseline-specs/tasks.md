# Tasks

This change establishes the baseline contract only. Building the app itself is split into follow-up changes, one per capability or vertical slice.

## 1. Project context

- [x] 1.1 Fill `context` in `openspec/config.yaml`: product summary, stack, v1 target, artifact language and the domain language (terms with "avoid" lists); verify `openspec context --json` succeeds and the file parses as YAML

## 2. Spec validation

- [x] 2.1 Run `openspec validate baseline-specs --strict` and fix every error and warning in the delta specs; verify the command exits 0
- [x] 2.2 Cross-check terms across all nine specs (learner, learner folder, control, exclusion group, Reset, Full reset, current/completed/new activity) and verify each term is used with one meaning everywhere

## Workflow follow-up

- Archive the change (`/opsx:archive`) so the nine specs land in `openspec/specs/`.
- Verify with `openspec list --specs` that all nine capabilities are present.
