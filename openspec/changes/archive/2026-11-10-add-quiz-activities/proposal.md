# Proposal

## Why

`quiz` activities are parsed and validated, but the Learner never sees their questions: `library.Activity` drops the quiz data and the activity view renders only the theory blocks. A course author can write a quiz today and the Learner gets a page of theory with nothing to answer. This change makes `quiz` the second usable activity type after `theory`.

## What Changes

- Go passes the quiz questions of a `quiz` activity to the frontend with the activity: prompts, options with `correct` and `feedback`, explanations. Answers are evaluated in the frontend; they are not protected anyway.
- A `quiz` activity opens on a theory screen showing its theory blocks and a "Start quiz" button. The quiz screen shows one question at a time, as slides, with a question map on top; a "Theory" button goes back to the theory screen, whose button then reads "Back to quiz". Switching screens keeps every answer, the open question and the map.
- Each question is checked on its own with "Check". The answer is evaluated by exact match, the question locks, and the app shows whether it is correct, highlights the correct options and the selected incorrect ones, and shows the feedback of the selected options and the explanation.
- The question map has one icon per question showing unanswered, correct or incorrect, marks the open question and opens any question on click. Next to it a counter shows how many questions are answered correctly. "Previous" and "Next question" move between questions; there is no auto-advance and no summary slide.
- "Start over" clears every answer and opens the first question, at any time.
- Answers are still not saved: every time the activity is opened, the quiz starts fresh on the theory screen. Next and Finish course stay available on both screens; the quiz never affects progress.
- Not in this change: controls of course activities (still declared and validated, not shown), saving quiz answers, the `activity.yaml` format (unchanged; `theory_blocks` stays required for `quiz`).

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `quiz`: how questions are shown (theory screen, question slides, question map with a correct-answer counter), checking a question locks it and highlights the correct options, "Start over"; answers survive switching between the theory and quiz screens.
- `activity-content`: the `quiz` type shows its theory blocks on a separate theory screen before the questions instead of above them.

## Impact

- Code: `internal/library` adds the quiz to `Activity` (`activity.go`); Wails regenerates `frontend/wailsjs/go/models.ts`.
- Frontend: a new quiz component in `frontend/src/course/` (theory screen, quiz screen, question map, question slide), a pure evaluation function, `ActivityView` rendering it for `quiz`; i18n strings in `en.json` and `ru.json`; vitest tests.
- No changes to `activity.yaml`, JSON Schemas, `apperr` codes, progress or files in the data folder.
