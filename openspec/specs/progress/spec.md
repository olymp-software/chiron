# progress Specification

## Purpose

Defines how the learner moves through a course, what counts as a completed activity and a completed course, and how progress behaves when the course version changes.

## Requirements

### Requirement: Strictly sequential progression
Progression SHALL be strictly sequential. The current activity SHALL be the first activity in course order that is not completed, or the last activity when every activity is completed. Every completed activity and the current activity SHALL be available; every other activity SHALL be locked. Initially only the first activity SHALL be available.

#### Scenario: New course
- **WHEN** the learner enrolls in a course
- **THEN** only the first activity is available

#### Scenario: Locked activity
- **WHEN** the learner tries to open an activity after the current one that is not completed
- **THEN** the app does not open it

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

### Requirement: Progress state
Progress SHALL consist of a pointer to the current activity and a set of completed activity slugs. There is no history of passes. Progress SHALL exist only for a course the learner is enrolled in; it SHALL be created empty at Enroll.

#### Scenario: Going back
- **WHEN** the learner reopens an earlier completed activity from the tree
- **THEN** the current activity does not change

#### Scenario: Course not enrolled
- **WHEN** a course is in the library and the learner has not enrolled in it
- **THEN** the course has no progress

### Requirement: Completing the course
On the last activity of the course, "Next" SHALL be replaced by "Finish course", which marks the activity completed and shows that the course is completed. A course SHALL be completed when every activity in the selected version is completed; the pointer then stays on the last activity. While the course is completed, its last activity SHALL show that the course is completed instead of "Finish course".

#### Scenario: Finish
- **WHEN** the learner presses "Finish course" and every other activity is completed
- **THEN** the app shows the course as completed

#### Scenario: Reopening a completed course
- **WHEN** the learner has finished the course, closes the app and starts it again
- **THEN** the last activity opens and shows that the course is completed

#### Scenario: Version with more activities
- **WHEN** a completed course switches to a version with activities after the former last one
- **THEN** the course is no longer completed, the first added activity becomes current, and the former last activity shows "Next" again

#### Scenario: Sync with more activities
- **WHEN** the learner presses Sync for a completed local-folder course whose folder has a new activity anywhere in the course
- **THEN** the course is no longer completed and the new activity becomes current

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

### Requirement: New activities after a version switch
An activity that a version switch or Sync adds before the current activity SHALL be not completed and SHALL become the current activity, since it is the first activity that is not completed. Completed activities after it SHALL stay available. A completed course with such an activity SHALL no longer be completed.

#### Scenario: Activity inserted behind the learner
- **WHEN** a new version inserts an activity before the current one
- **THEN** the inserted activity becomes current and the completed activities after it stay available

#### Scenario: Activity inserted into a completed course
- **WHEN** the learner presses Sync for a completed local-folder course whose folder has a new activity in the middle of the course
- **THEN** the course is no longer completed, the new activity becomes current, and the overview shows Continue instead of Review

#### Scenario: Activity added ahead
- **WHEN** a new version adds an activity after the current one
- **THEN** it opens in its turn

### Requirement: Removed activities
If the current activity is removed in the selected version, the pointer SHALL move to the first activity in the new order that is not completed, or the last activity when every activity is completed. A removed activity's learner folder SHALL simply stop being used.

#### Scenario: Current activity removed
- **WHEN** the current activity no longer exists in the new version
- **THEN** the current activity becomes the first activity in the new order that is not completed

### Requirement: Content changes apply immediately
Changes to content, quiz questions or scripts in the selected version SHALL be visible to the learner immediately.

#### Scenario: Fixed typo
- **WHEN** a new version fixes a typo in an activity's content and the learner switches to it
- **THEN** the learner sees the fixed text
