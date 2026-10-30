# Spec Delta

## MODIFIED Requirements

### Requirement: Only Next moves progress
Every activity except the last SHALL have a "Next" button that marks the activity completed and opens the following one. What the learner did in the activity (quiz answers, script results) MUST NOT matter. There SHALL be no other completion rules in the format.

#### Scenario: Next without doing anything
- **WHEN** the learner opens a `coding` activity and immediately presses "Next"
- **THEN** the activity is completed and the next one becomes current

#### Scenario: Next on an activity with an error
- **WHEN** the current activity shows an activity error and the learner presses "Next"
- **THEN** the activity is completed and the next one becomes current

#### Scenario: Next on an earlier activity
- **WHEN** the learner reopens an earlier completed activity from the tree and presses "Next"
- **THEN** the following activity opens and the current activity does not change

### Requirement: Completing the course
On the last activity of the course, "Next" SHALL be replaced by "Finish course", which marks the activity completed and shows that the course is completed. The pointer SHALL stay on the last activity. A course SHALL be completed when its last activity in the selected version is completed. While the course is completed, its last activity SHALL show that the course is completed instead of "Finish course".

#### Scenario: Finish
- **WHEN** the learner presses "Finish course"
- **THEN** the app shows the course as completed

#### Scenario: Reopening a completed course
- **WHEN** the learner has finished the course, closes the app and starts it again
- **THEN** the last activity opens and shows that the course is completed

#### Scenario: Version with more activities
- **WHEN** a completed course switches to a version with activities after the former last one
- **THEN** the course is no longer completed and that activity shows "Next" again

### Requirement: Navigation through the tree
The learner SHALL be able to open any available activity through the course tree. The tree SHALL mark completed activities and the current activity, and show locked activities as unavailable. Whenever entering a course opens the course screen, including on app start, the app SHALL open the current activity, not the one opened last.

#### Scenario: Restart after browsing back
- **WHEN** the learner opened an earlier activity, closed the app and started it again
- **THEN** the current activity is open

#### Scenario: Marks in the tree
- **WHEN** the learner has completed the first two activities of a course and the third is current
- **THEN** the tree marks the first two as completed, the third as current, and shows every later activity as locked

#### Scenario: Clicking a locked activity
- **WHEN** the learner clicks a locked activity in the tree
- **THEN** nothing opens
