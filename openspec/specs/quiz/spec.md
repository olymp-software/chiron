# quiz Specification

## Purpose

Defines the structure of quiz questions, the `choice` question kind, how the learner's answer is evaluated, and what feedback the learner sees.

## Requirements

### Requirement: Quiz data
A quiz's `data` SHALL contain `questions`: a list of at least one question, shown in order. A question SHALL have `kind` (required), `prompt` (Markdown, optional), `data` (object whose shape depends on `kind`, required) and `explanation` (Markdown, optional). An empty `prompt` or `explanation` SHALL not be shown.

```yaml
data:
  questions:
    - kind: choice
      prompt: Which Kubernetes objects are namespaced?
      data:
        multiple: true
        options:
          - text: "`Pod`"
            correct: true
          - text: "`Node`"
            feedback: Nodes are cluster-scoped.
      explanation: Namespaced objects live inside a namespace.
```

#### Scenario: No questions
- **WHEN** `questions` is an empty list
- **THEN** opening the activity shows an activity error

#### Scenario: Question without prompt
- **WHEN** a question has no `prompt`
- **THEN** only its options are shown

### Requirement: choice kind
In v1 the only question kind SHALL be `choice`. Its `data` SHALL have `multiple` (bool, default `false`; `false` shows radio buttons, `true` shows checkboxes) and `options` (list of at least two, shown in order without shuffling). An option SHALL have `text` (Markdown, required), `correct` (bool, default `false`) and `feedback` (Markdown, optional).

#### Scenario: Single option
- **WHEN** a `choice` question has one option
- **THEN** opening the activity shows an activity error

#### Scenario: Order is preserved
- **WHEN** the quiz is opened several times
- **THEN** options always appear in the order written

### Requirement: Correct option count
With `multiple: false` exactly one option SHALL be correct. With `multiple: true` at least one option SHALL be correct; a single correct option is allowed so that the input type (radio buttons or checkboxes) does not hint at the answer.

#### Scenario: Radio with two correct options
- **WHEN** `multiple` is `false` and two options are `correct: true`
- **THEN** opening the activity shows an activity error

#### Scenario: Checkboxes with one correct option
- **WHEN** `multiple` is `true` and exactly one option is correct
- **THEN** the question is valid

### Requirement: Exact-match evaluation
An answer SHALL count as correct only when the selected set of options equals the set of correct options. There is no partial credit.

#### Scenario: Partial selection
- **WHEN** two options are correct and the learner selects only one of them
- **THEN** the answer is incorrect

### Requirement: Feedback after answering
After the learner answers, the app SHALL show the feedback of each selected option that has one, and the question's explanation if present.

#### Scenario: Wrong option with feedback
- **WHEN** the learner selects an option with `feedback`
- **THEN** that feedback is shown after answering

### Requirement: Quiz result does not affect progress
The quiz result SHALL be feedback for the learner only and MUST NOT affect progress. Learner answers MUST NOT be saved: each time the activity is opened the quiz starts fresh.

#### Scenario: Wrong answers and Next
- **WHEN** the learner answers every question incorrectly and presses "Next"
- **THEN** the activity becomes completed

#### Scenario: Reopening the quiz
- **WHEN** the learner answers a quiz, leaves and returns to it
- **THEN** no options are selected

### Requirement: Answers are not protected
Correct options SHALL be marked in plain text in the course. Protection against peeking is not provided.

#### Scenario: Learner reads the file
- **WHEN** the learner opens `activity.yaml` in the course clone
- **THEN** the correct options are visible
