# progress Specification

## Purpose

Defines how the learner moves through a course, what counts as a completed activity and a completed course, and how progress behaves when the course version changes.

## Requirements

### Requirement: Strictly sequential progression
Progression SHALL be strictly sequential. Initially only the first activity SHALL be available. Every activity up to and including the current activity SHALL be available; every activity after it SHALL be locked.

#### Scenario: New course
- **WHEN** the learner starts a course
- **THEN** only the first activity is available

#### Scenario: Locked activity
- **WHEN** the learner tries to open an activity after the current one
- **THEN** the app does not open it

### Requirement: Only Next moves progress
Every activity except the last SHALL have a "Next" button that marks the activity completed and opens the following one. What the learner did in the activity (quiz answers, script results) MUST NOT matter. There SHALL be no other completion rules in the format.

#### Scenario: Next without doing anything
- **WHEN** the learner opens a `coding` activity and immediately presses "Next"
- **THEN** the activity is completed and the next one becomes current

### Requirement: Progress state
Progress SHALL consist of a pointer to the current activity and a set of completed activity slugs. There is no history of passes.

#### Scenario: Going back
- **WHEN** the learner reopens an earlier completed activity from the tree
- **THEN** the current activity does not change

### Requirement: Completing the course
On the last activity of the course, "Next" SHALL be replaced by "Finish course", which marks the activity completed and shows that the course is completed. The pointer SHALL stay on the last activity. A course SHALL be completed when its last activity in the selected version is completed.

#### Scenario: Finish
- **WHEN** the learner presses "Finish course"
- **THEN** the app shows the course as completed

#### Scenario: Version with more activities
- **WHEN** a completed course switches to a version with activities after the former last one
- **THEN** the course is no longer completed and that activity shows "Next" again

### Requirement: Navigation through the tree
The learner SHALL be able to open any available activity through the course tree. On entering a course, including on app start, the app SHALL open the current activity, not the one opened last.

#### Scenario: Restart after browsing back
- **WHEN** the learner opened an earlier activity, closed the app and started it again
- **THEN** the current activity is open

### Requirement: New activities after a version switch
An activity that, after a version switch, lies before the current activity but is not completed SHALL be available and marked with a "new" badge. The learner SHALL continue from where they were.

#### Scenario: Activity inserted behind the learner
- **WHEN** a new version inserts an activity before the current one
- **THEN** it is available with a "new" badge and the current activity is unchanged

#### Scenario: Activity added ahead
- **WHEN** a new version adds an activity after the current one
- **THEN** it opens in its turn

### Requirement: Removed activities
If the current activity is removed in the selected version, the pointer SHALL move to the activity following the last completed one in the new order. A removed activity's learner folder SHALL simply stop being used.

#### Scenario: Current activity removed
- **WHEN** the current activity no longer exists in the new version
- **THEN** the current activity becomes the one after the last completed activity in the new order

### Requirement: Content changes apply immediately
Changes to content, quiz questions or scripts in the selected version SHALL be visible to the learner immediately.

#### Scenario: Fixed typo
- **WHEN** a new version fixes a typo in an activity's content and the learner switches to it
- **THEN** the learner sees the fixed text
