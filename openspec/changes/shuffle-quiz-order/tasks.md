# Tasks

## 1. Schema and Go model

- [ ] 1.1 Add `shuffle` (boolean) to `quizData` and `choiceData` in `schemas/course-activity/v1.json`, with descriptions (question order / option order, default `true`); verify with `internal/manifest` tests that `shuffle: false` in both places parses, `shuffle: "no"` in a choice `data` and `shuffle` next to a question's `kind` give an activity error.
- [ ] 1.2 Add `Shuffle *bool` to `manifest.Quiz` and `manifest.Choice`; verify with `task go:test -- -run Activity` that an absent flag decodes to `nil` and `false` / `true` decode as written.
- [ ] 1.3 Add `Shuffle bool` to `library.Quiz` and `library.Question`, resolving an absent flag to `true` in `quiz()` in `internal/library/activity.go`; verify with a `internal/library` test for absent, `true` and `false` at both levels, and `task go:lint` / `task go:lint:internal` pass.

## 2. Frontend shuffling

- [ ] 2.1 Regenerate the Wails bindings (`wails generate` via the project's task or `task build`) so `library.Quiz` and `library.Question` in `frontend/wailsjs/go/models.ts` carry `shuffle`; verify `task frontend:check` type-checks.
- [ ] 2.2 Add `arrange(quiz, random = Math.random)` to `frontend/src/quiz/quiz.ts`: returns a copy with questions permuted when `quiz.shuffle` and each question's options permuted when `question.shuffle` (Fisher-Yates), leaving the input untouched; verify with `quiz.test.ts` using a deterministic generator: written order kept for `false` flags, flags independent, each option keeps its `correct` and `feedback`, `isCorrect` works on the arranged question.
- [ ] 2.3 In `QuizActivity.vue`, keep the arranged quiz in a ref built on creation and rebuilt in `startOver` together with fresh answers, and drive the map, `QuizQuestion`, numbering and navigation from it; verify with `QuizActivity.test.ts`: questions shown in arranged order with "Question 1 of N" on the first shown one, order unchanged after "Theory" / "Back to quiz", "Start over" draws a new order, a remounted component draws a new order, and `shuffle: false` quizzes match the written order. Existing quiz tests pass with `task frontend:check` (set `shuffle: false` in fixtures that assert written order).

## 3. Author guide and feature requests

- [ ] 3.1 Update the Quizzes section of `docs/authoring.md`: `shuffle` in the quiz `data` (question order) and in a `choice` question's `data` (option order), both default `true`, independent; a new order on every opening and on **Start over**; don't refer to positions in prompts, feedback or explanations, or set `shuffle: false`. Remove "(no shuffling: vary the position of correct options yourself)" and shuffling from the closing "ideas not yet specified" note; verify by reading the section against the `quiz` delta spec.
- [ ] 3.2 Remove the "Shuffling" entry from `openspec/feature-requests.md`; verify the file no longer mentions shuffling.

## 4. Manual check in the app

- [ ] 4.1 Create test course `tmp/courses/15-quiz-shuffle` with three quiz activities: one with defaults (4+ questions, 4+ options each), one with `shuffle: false` in the quiz `data` only, and one where every question has `shuffle: false` in its `data` and the last option is "All of the above". Run `task run`, add the course and check: the default quiz changes question and option order on reopening and on **Start over** but not on **Theory** / **Back to quiz**; the second keeps question order while options move; the third keeps "All of the above" last while questions move; feedback, highlighting and the counter stay correct after shuffling.
