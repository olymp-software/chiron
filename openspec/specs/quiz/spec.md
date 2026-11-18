# quiz Specification

## Purpose

Defines the structure of quiz questions, the `choice` question kind, how the learner's answer is evaluated, and what feedback the learner sees.

## Requirements

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
After the learner checks a question, the app SHALL show the feedback of each selected option that has one, and the question's explanation if present.

#### Scenario: Wrong option with feedback
- **WHEN** the learner selects an option with `feedback` and checks the question
- **THEN** that feedback is shown after checking

#### Scenario: Feedback of an unselected option
- **WHEN** an option with `feedback` is not selected and the question is checked
- **THEN** its feedback is not shown

### Requirement: Quiz result does not affect progress
The quiz result SHALL be feedback for the learner only and MUST NOT affect progress. Learner answers MUST NOT be saved: each time the activity is opened the quiz starts fresh.

#### Scenario: Wrong answers and Next
- **WHEN** the learner answers every question incorrectly and presses "Next"
- **THEN** the activity becomes completed

#### Scenario: Reopening the quiz
- **WHEN** the learner answers a quiz, leaves and returns to it
- **THEN** the theory screen is shown, and after "Start quiz" no question is checked and no options are selected

### Requirement: Answers are not protected
Correct options SHALL be marked in plain text in the course. Protection against peeking is not provided.

#### Scenario: Learner reads the file
- **WHEN** the learner opens `activity.yaml` in the course clone
- **THEN** the correct options are visible

### Requirement: Theory screen first
A quiz activity SHALL open on its theory screen: the theory blocks in order and a "Start quiz" button that opens the quiz screen. The quiz screen SHALL have a "Theory" button that opens the theory screen again; once the quiz has been started, the theory screen's button SHALL read "Back to quiz". Next or Finish course SHALL be available on both screens.

#### Scenario: Opening a quiz
- **WHEN** the learner opens a quiz activity
- **THEN** the theory blocks and "Start quiz" are shown and no question is shown

#### Scenario: Starting the quiz
- **WHEN** the learner presses "Start quiz"
- **THEN** the quiz screen opens on the first question

#### Scenario: Next on the theory screen
- **WHEN** the learner presses Next on the theory screen without starting the quiz
- **THEN** the activity becomes completed

### Requirement: Screen switches keep the quiz
Switching between the theory screen and the quiz screen SHALL keep the quiz as it was: the order of questions and options, every checked question with its result, the selection of unchecked questions, the open question and the question map.

#### Scenario: Back from the theory
- **WHEN** the learner checks the first question, opens the second, presses "Theory" and then "Back to quiz"
- **THEN** the second question is open, the first one is still checked with the same result, and the map and counter are unchanged

#### Scenario: Order survives a screen switch
- **WHEN** the learner opens the quiz screen, presses "Theory" and then "Back to quiz"
- **THEN** the questions and the options of every question are in the same order as before

### Requirement: One question at a time
The quiz screen SHALL show one question at a time with its number and the total, such as "Question 3 of 5". "Previous" and "Next question" SHALL open the neighbouring question; "Previous" SHALL be unavailable on the first question and "Next question" on the last. Checking a question MUST NOT open another question.

#### Scenario: Last question
- **WHEN** the last question is open
- **THEN** "Next question" is unavailable and there is no summary screen after it

#### Scenario: No auto-advance
- **WHEN** the learner checks a question
- **THEN** the same question stays open with its result

### Requirement: Checking a question
Each question SHALL have a "Check" button, unavailable while no option is selected. Checking SHALL evaluate the selection, lock the question so its selection can no longer change, and show whether the answer is correct. A checked question SHALL stay locked until "Start over".

#### Scenario: Check without selection
- **WHEN** no option of the open question is selected
- **THEN** "Check" is unavailable

#### Scenario: Changing a checked answer
- **WHEN** the learner tries to select another option of a checked question
- **THEN** the selection does not change

### Requirement: Correct options are highlighted
After a question is checked, the app SHALL highlight every correct option, whether selected or not, and SHALL put a verdict icon only on selected options: correct or incorrect. A correct option the learner did not select SHALL be highlighted without an icon. Each mark SHALL also be available as text to screen readers.

#### Scenario: Wrong single choice
- **WHEN** the learner selects an incorrect option of a question with `multiple: false` and checks it
- **THEN** the selected option has the incorrect icon and no highlight, and the correct option is highlighted without an icon

#### Scenario: Missed option
- **WHEN** two options are correct and the learner selects only one of them and checks
- **THEN** both correct options are highlighted, only the selected one has the correct icon, and the answer is incorrect

### Requirement: Question map
The quiz screen SHALL show a question map with one icon per question, in the order the questions are shown. Each icon SHALL show whether the question is unchecked, answered correctly or answered incorrectly; the open question SHALL be marked. Selecting an icon SHALL open its question, in any order.

#### Scenario: Map after answers
- **WHEN** the learner answers question 1 correctly and question 2 incorrectly in a quiz of four questions
- **THEN** the map shows question 1 as correct, question 2 as incorrect and questions 3 and 4 as unchecked

#### Scenario: Jump to a question
- **WHEN** the learner selects the icon of question 4 while question 1 is open
- **THEN** question 4 opens

### Requirement: Correct answer counter
Next to the question map the app SHALL show how many questions are answered correctly out of the total, such as "Correct: 3 of 5".

#### Scenario: Counter after a wrong answer
- **WHEN** the learner answers one question incorrectly in a fresh quiz
- **THEN** the counter shows 0 correct

### Requirement: Start over
The quiz screen SHALL have a "Start over" button, available at any time without confirmation. It SHALL clear every selection and result, draw a new order of questions and options, and open the first question on the quiz screen.

#### Scenario: Start over after answers
- **WHEN** the learner has checked three questions and presses "Start over"
- **THEN** every question is unchecked with no option selected, the first question is open and the counter shows 0 correct

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
