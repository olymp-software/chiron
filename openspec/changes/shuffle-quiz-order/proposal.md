# Proposal

## Why

Quiz options and questions are always shown in the order written. Authors have to vary the position of the correct option by hand, and a Learner retaking a quiz remembers answers by position instead of by content.

## What Changes

- Quiz questions and the options of each `choice` question are shown in a random order by default.
- A new random order is drawn whenever the quiz starts fresh: each time the activity is opened and on "Start over". Switching between the theory screen and the quiz screen keeps the order.
- `shuffle: false` in the quiz `data` keeps the question order as written; `shuffle: false` in a `choice` question's `data` keeps that question's option order as written. The two flags are independent.
- Question numbers, the question map and "Previous" / "Next question" follow the shown order.
- The activity JSON Schema gains the two `shuffle` fields in place, at schema version 1: nothing is released yet.
- **BREAKING** (unreleased): quizzes that relied on the written option order (e.g. "All of the above" last) must set `shuffle: false` on that question.
- The Author guide describes shuffling and drops it from "Not yet available"; the shuffling entry is removed from `openspec/feature-requests.md`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `quiz`: questions and options are shuffled by default, with `shuffle: false` at quiz and question level; when a new order is drawn; numbering and the map follow the shown order.

## Impact

- `schemas/course-activity/v1.json`: `shuffle` in `quizData` and `choiceData`.
- `internal/manifest` (quiz types) and `internal/library` (quiz passed to the frontend), with tests.
- Frontend: `frontend/src/quiz/quiz.ts`, `course/QuizActivity.vue`, `course/QuizQuestion.vue` and their tests.
- `docs/authoring.md`, `openspec/feature-requests.md`.
