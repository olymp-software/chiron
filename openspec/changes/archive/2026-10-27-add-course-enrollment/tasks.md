# Tasks

## 1. Glossary

- [x] 1.1 Add Enrollment, Library, Course overview and Overview file to the domain glossary in `openspec/config.yaml` (with "Avoid" words: subscription, start / catalog / course page, course card / readme, about) and verify `openspec context --json` still loads without errors
- [x] 1.2 Update the Description entry (shown in the course overview when there is no overview file) and the Environment check entry (gates Enroll and Continue instead of "the course cannot start") and verify no other glossary entry contradicts the new terms

## 2. Spec consistency

- [x] 2.1 Search `openspec/specs/` for wording that assumes adding a course starts it or that the check lives on the course screen (`rg -n -i "start|course screen|open last|Check environment" openspec/specs`) and verify every hit is covered by a delta in this change or still holds
- [x] 2.2 Run `openspec validate add-course-enrollment --strict` and verify it passes

## Workflow follow-up

- Archive the change so the deltas are merged into the main specs.
- Library, course overview, enrollment storage and the entry guard are implemented in the feature changes that build those screens, against the archived specs.
