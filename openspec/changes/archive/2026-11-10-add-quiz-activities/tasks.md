# Tasks

## 1. Quiz data in Go

- [x] 1.1 Add `Quiz`, `Question` and `Option` with JSON tags to `internal/library/activity.go`, fill `Activity.Quiz` from `manifest.Quiz` for `type: quiz` only, and verify with a `library` test that a quiz activity returns every question, option, `correct`, `feedback`, `explanation` and `multiple`, and a theory activity has no quiz (`task go:test`)
- [x] 1.2 Regenerate `frontend/wailsjs/go/models.ts` (`wails generate module` / `task build`) and verify `library.Activity` has `quiz` with the new classes; `task go:lint` and `task go:lint:internal` pass

## 2. Evaluation

- [x] 2.1 Add a pure evaluation function in `frontend/src/quiz/` (exact match of selected and correct option sets) and verify with vitest: single correct, wrong single, partial selection on `multiple`, extra selection, exact `multiple` match (`task frontend:test`)

## 3. Quiz components

- [x] 3.1 Add `course/QuizQuestion.vue`: "Question N of M", prompt, radio or checkbox options by `multiple`, "Check" disabled without a selection, locked after checking with correct / incorrect marking, feedback of selected options and explanation, all text through `MarkdownView`; verify with vitest covering the "Checking a question", "Correct options are highlighted" and "Feedback after answering" scenarios
- [x] 3.2 Add `course/QuizMap.vue`: one icon per question for unchecked / correct / incorrect, the open question marked, selecting an icon emits it, and the "Correct: X of Y" counter; verify with vitest covering the "Question map" and "Correct answer counter" scenarios
- [x] 3.3 Add `course/QuizActivity.vue` owning the quiz state: theory screen with "Start quiz" / "Back to quiz", quiz screen with map, "Theory", "Start over", "Previous" / "Next question"; verify with vitest covering "Theory screen first", "One question at a time", "Start over" and "Screen switches keep the quiz" (answer question 1, open question 2, go to theory and back: same question open, same results)
- [x] 3.4 Add the quiz UI strings (`quiz.*`) to `frontend/src/i18n/en.json` and `ru.json`, using "Start over" rather than "Reset", and verify the i18n key parity test passes

## 4. Activity view

- [x] 4.1 Render `QuizActivity` in `ActivityView.vue` for an activity with `quiz` and no error, keyed per opened activity, with Next / Finish course still below it; verify with `ActivityView` vitest that a quiz opens on the theory screen, Next works on both screens, and opening another activity and returning shows a fresh quiz
- [x] 4.2 Run `task check` and verify everything passes

## 5. Manual check

- [x] 5.1 Add a quiz test course to `tmp/courses/NN-quiz` (single and multiple choice, options with feedback, a question without prompt, an explanation, Markdown and an image) and its checks to `tmp/courses/CHECKLIST.md`; verify them in the app via `task run`

## Workflow follow-up

- Archive the change after review; verify `openspec/specs/quiz/spec.md` and `activity-content/spec.md` contain the merged requirements.
