# Spec Delta

## MODIFIED Requirements

### Requirement: Quiz data
A quiz's `data` SHALL contain `questions`: a list of at least one question, and `shuffle` (bool, optional, default `true`), which controls the question order. A question SHALL have `kind` (required), `prompt` (Markdown, optional), `data` (object whose shape depends on `kind`, required) and `explanation` (Markdown, optional). An empty `prompt` or `explanation` SHALL not be shown.

```yaml
data:
  shuffle: false
  questions:
    - kind: choice
      prompt: Which Kubernetes objects are namespaced?
      data:
        multiple: true
        shuffle: false
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

#### Scenario: shuffle on a question
- **WHEN** a question has `shuffle` next to `kind` instead of inside its `data`
- **THEN** opening the activity shows an activity error

### Requirement: choice kind
In v1 the only question kind SHALL be `choice`. Its `data` SHALL have `multiple` (bool, default `false`; `false` shows radio buttons, `true` shows checkboxes), `shuffle` (bool, default `true`), which controls the option order, and `options` (list of at least two). An option SHALL have `text` (Markdown, required), `correct` (bool, default `false`) and `feedback` (Markdown, optional).

#### Scenario: Single option
- **WHEN** a `choice` question has one option
- **THEN** opening the activity shows an activity error

#### Scenario: Order is preserved
- **WHEN** a question has `shuffle: false` in its `data` and the quiz is opened several times
- **THEN** that question's options always appear in the order written

#### Scenario: shuffle is not a boolean
- **WHEN** a `choice` question has `shuffle: "no"`
- **THEN** opening the activity shows an activity error

### Requirement: Screen switches keep the quiz
Switching between the theory screen and the quiz screen SHALL keep the quiz as it was: the order of questions and options, every checked question with its result, the selection of unchecked questions, the open question and the question map.

#### Scenario: Back from the theory
- **WHEN** the learner checks the first question, opens the second, presses "Theory" and then "Back to quiz"
- **THEN** the second question is open, the first one is still checked with the same result, and the map and counter are unchanged

#### Scenario: Order survives a screen switch
- **WHEN** the learner opens the quiz screen, presses "Theory" and then "Back to quiz"
- **THEN** the questions and the options of every question are in the same order as before

### Requirement: Question map
The quiz screen SHALL show a question map with one icon per question, in the order the questions are shown. Each icon SHALL show whether the question is unchecked, answered correctly or answered incorrectly; the open question SHALL be marked. Selecting an icon SHALL open its question, in any order.

#### Scenario: Map after answers
- **WHEN** the learner answers question 1 correctly and question 2 incorrectly in a quiz of four questions
- **THEN** the map shows question 1 as correct, question 2 as incorrect and questions 3 and 4 as unchecked

#### Scenario: Jump to a question
- **WHEN** the learner selects the icon of question 4 while question 1 is open
- **THEN** question 4 opens

### Requirement: Start over
The quiz screen SHALL have a "Start over" button, available at any time without confirmation. It SHALL clear every selection and result, draw a new order of questions and options, and open the first question on the quiz screen.

#### Scenario: Start over after answers
- **WHEN** the learner has checked three questions and presses "Start over"
- **THEN** every question is unchecked with no option selected, the first question is open and the counter shows 0 correct

## ADDED Requirements

### Requirement: Shuffled question order
When the quiz's `shuffle` is `true` or absent, the questions SHALL be shown in a random order. With `shuffle: false` they SHALL be shown in the order written. Question numbers such as "Question 1 of 5", "Previous" and "Next question" SHALL follow the shown order.

#### Scenario: Questions in written order
- **WHEN** the quiz has `shuffle: false` and is opened several times
- **THEN** the questions always appear in the order written

#### Scenario: Questions shuffled by default
- **WHEN** the quiz has no `shuffle` and is opened many times
- **THEN** the questions appear in more than one order

#### Scenario: Numbering follows the shown order
- **WHEN** the questions are shuffled and the third written question is shown first
- **THEN** it is shown as "Question 1 of N" and "Previous" is unavailable on it

### Requirement: Shuffled option order
When a `choice` question's `shuffle` is `true` or absent, its options SHALL be shown in a random order. With `shuffle: false` they SHALL be shown in the order written. The question's flag SHALL be independent of the quiz's `shuffle`. Correctness, highlighting and feedback SHALL stay bound to each option, not to its position.

#### Scenario: Options shuffled by default
- **WHEN** a question has no `shuffle` and the quiz is opened many times
- **THEN** its options appear in more than one order

#### Scenario: Flags are independent
- **WHEN** the quiz has `shuffle: false` and its questions have no `shuffle`
- **THEN** the questions appear in the order written and the options of each question are shuffled

#### Scenario: Feedback follows the option
- **WHEN** options are shuffled and the learner selects an option with `feedback` and checks the question
- **THEN** that option's own feedback is shown and its own correctness decides its mark

### Requirement: New order on every fresh start
The app SHALL draw a new random order of questions and options whenever the quiz starts fresh: each time the activity is opened and on "Start over". The order MUST NOT be saved. A drawn order MAY coincide with the written order or with the previous one.

#### Scenario: Reopening draws a new order
- **WHEN** the learner leaves a shuffled quiz activity and opens it again
- **THEN** the order of questions and options is drawn again, independent of the previous one
