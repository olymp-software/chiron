# Design

## Context

- `internal/manifest` decodes quiz `data` into `Quiz` / `Question` / `Choice`; `internal/library` flattens each `choice` question (`Multiple`, `Options`) into `library.Question` for the frontend.
- The whole quiz state lives in `QuizActivity.vue`: `answers` holds selected option indexes per question, and indexes are positions in `quiz.questions` and `question.options`. A new `QuizActivity` is mounted for every opened activity; "Start over" resets `answers` in place. Switching screens keeps the component, so the quiz state survives.
- `QuizQuestion.vue`, `QuizMap.vue` and `quiz.ts` (`isCorrect`) all work with those indexes.
- The manifest spec requires a schema version bump for any schema change. Nothing is released yet, so `schemas/course-activity/v1.json` is changed in place (decided with the project owner); the bump rule applies from the first release.

## Goals / Non-Goals

**Goals:**
- Shuffle questions and options in the frontend, at the moments the quiz already starts fresh.
- Keep answer evaluation, highlighting, feedback and the map untouched by working on an already reordered quiz.

**Non-Goals:**
- Pinning single options in place (`fixed` on an option): rejected; an Author who needs a fixed position turns shuffling off for the question.
- A seed, a remembered order, or a guarantee that the order differs from the written or previous one.
- `shuffle` for workshops: workshop activities have no quizzes.

## Decisions

### `shuffle` lives in kind-specific `data`

The question-level flag sits in the `choice` question's `data`, next to `multiple`, not next to `kind`. It controls options, which belong to the kind; a future kind without options (e.g. `true_false`) would not have it. The quiz-level flag sits in the quiz `data` and controls only the question order. The two are independent: no inheritance, so no precedence rule to explain.

### Defaults resolved in Go, plain booleans to the frontend

`manifest.Quiz.Shuffle` and `manifest.Choice.Shuffle` are `*bool` so that "absent" can mean `true`. `library.Quiz` and `library.Question` get a plain `Shuffle bool` with the default already applied, so the frontend never deals with `undefined`. Alternative: an inverted `NoShuffle` / `keep_order` field to get a zero-value default; rejected because the YAML name the Authors agreed on is `shuffle: false`.

### Shuffle by reordering a copy of the quiz in the frontend

A pure function in `quiz/quiz.ts`, e.g. `arrange(quiz, random = Math.random): library.Quiz`, returns a copy with the questions and each question's options permuted (Fisher-Yates) according to the flags. `QuizActivity` keeps the arranged quiz in a ref: it is computed when the component is created (activity opened) and again in `startOver`, together with fresh answers. Everything downstream (`answers` indexes, `isCorrect`, `QuizQuestion`, `QuizMap`, numbering, "Previous" / "Next question") keeps working on positions in the arranged quiz, and since each option object carries its own `correct` and `feedback`, they follow the option.

Alternatives:
- *Shuffle in Go when the activity is read:* the order would change only when the activity is fetched, so "Start over" would need another call or a frontend shuffle anyway.
- *Keep the written quiz and store permutations next to it:* every index lookup would need mapping, touching `isCorrect`, `QuizQuestion` and `QuizMap` for no gain.

`random` is a parameter so tests pass a deterministic generator.

## Risks / Trade-offs

- [Prompts, feedback or explanations that refer to positions ("option B", "the previous question") become wrong] → the Author guide tells Authors not to refer to positions, or to set `shuffle: false`.
- [A random order may equal the written one, so a test asserting "different order" is flaky] → frontend tests use an injected generator; spec scenarios only require "more than one order" over many openings.
- [Changing v1 in place breaks any unreleased draft course that relied on the written order] → acceptable before the first release; such quizzes add `shuffle: false`.
