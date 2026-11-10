# Spec Delta

## ADDED Requirements

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
Switching between the theory screen and the quiz screen SHALL keep the quiz as it was: every checked question with its result, the selection of unchecked questions, the open question and the question map.

#### Scenario: Back from the theory
- **WHEN** the learner checks the first question, opens the second, presses "Theory" and then "Back to quiz"
- **THEN** the second question is open, the first one is still checked with the same result, and the map and counter are unchanged

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
After a question is checked, the app SHALL mark every correct option as correct, whether selected or not, and every selected incorrect option as incorrect.

#### Scenario: Wrong single choice
- **WHEN** the learner selects an incorrect option of a question with `multiple: false` and checks it
- **THEN** the selected option is marked incorrect and the correct option is marked correct

#### Scenario: Missed option
- **WHEN** two options are correct and the learner selects only one of them and checks
- **THEN** both correct options are marked correct and the answer is incorrect

### Requirement: Question map
The quiz screen SHALL show a question map with one icon per question, in order. Each icon SHALL show whether the question is unchecked, answered correctly or answered incorrectly; the open question SHALL be marked. Selecting an icon SHALL open its question, in any order.

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
The quiz screen SHALL have a "Start over" button, available at any time without confirmation. It SHALL clear every selection and result and open the first question on the quiz screen.

#### Scenario: Start over after answers
- **WHEN** the learner has checked three questions and presses "Start over"
- **THEN** every question is unchecked with no option selected, the first question is open and the counter shows 0 correct

## MODIFIED Requirements

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
