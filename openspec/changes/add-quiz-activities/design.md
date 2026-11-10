# Design

## Context

`manifest.ParseActivity` already decodes and checks quiz data into `manifest.Quiz` (at least one question, at least two options, correct option count). `library.Activity` copies only `Type` and the theory blocks, so the quiz never reaches the frontend. `ActivityView` renders theory blocks for every non-workshop activity and the workshop part through `WorkshopActivity`. `ActivityView` is reused across activities: it reloads on a `slug` change instead of remounting.

## Goals / Non-Goals

**Goals:**
- Quiz state that lives exactly as long as the opened activity and survives switching between the theory and quiz screens.
- Evaluation as a pure function covered by unit tests.

**Non-Goals:**
- Any change to the `activity.yaml` format, schemas or semantic checks.
- New Go bindings: the quiz arrives with the existing `Activity` call.

## Decisions

### Quiz data travels with `Activity`
`library.Activity` gets `Quiz *Quiz` (`json:"quiz,omitempty"`), set only for `type: quiz`. Its own structs mirror the manifest ones (`Question{Kind, Prompt, Explanation, Multiple, Options}`, `Option{Text, Correct, Feedback}`) with JSON tags, as `TheoryBlock` does, so the manifest's YAML structs do not become the frontend contract. `Multiple` moves up from `Choice` into the question because `choice` is the only kind.

Alternative: a separate `CheckAnswer` binding evaluated in Go. Rejected: answers are not protected by spec, nothing is saved, and a round trip per check buys nothing.

### Evaluation in the frontend
A pure function (e.g. `quiz/evaluate.ts`) takes a question and the selected option indexes and returns whether the answer is correct: the selected set equals the set of correct options. Highlighting is derived from the question alone after checking: correct options are marked correct, selected incorrect ones incorrect.

### One state object owned by the quiz component
A new `course/QuizActivity.vue` owns the whole quiz state:

```
QuizActivity  (state: screen, started, open index, per question: selected[], checked, correct)
  +-- theory screen: theory blocks + [Start quiz | Back to quiz]
  +-- quiz screen
        +-- QuizMap      (props: results, open index; emits open)
        +-- QuizQuestion (props: question, selection, checked; emits select, check)
```

`QuizMap` and `QuizQuestion` are stateless. Switching screens only changes `screen`, so the state survives even when the screens are toggled with `v-if`. "Start over" replaces the per-question state with a fresh one and sets the open index to 0.

Alternative: state inside the question slides component. Rejected: unmounting it on the switch to the theory screen loses the answers, which the spec forbids.

### Fresh quiz per opened activity
`ActivityView` renders `QuizActivity` with `:key="activity.slug"` (or the load counter) when `activity.quiz` is set and there is no activity error. A new activity, or reopening the same one, mounts a new component and therefore a fresh quiz on the theory screen. No store outlives the component.

### Next stays in `ActivityView`
The Next / Finish course footer stays where it is, below whichever screen is shown, so progress keeps a single code path for every activity type.

### Markdown in quiz texts
Prompts, option texts, feedback and explanations go through the existing `MarkdownView` with the activity image resolver, so they follow the same dialect and image rules as theory blocks.

## Risks / Trade-offs

- [Option text rendered as block Markdown wraps a paragraph inside a label] → style the paragraph margins away in the option; no separate inline renderer.
- [Leaving the activity loses answers, which may surprise the learner] → accepted for this change; saving answers is a later change, the state shape is local to `QuizActivity` so it can be lifted into a store then.
- [Sync while a quiz is open changes questions under the learner] → the current activity view is not reloaded by Sync today; the next open shows the new quiz fresh, so no stale indexes are reused.
