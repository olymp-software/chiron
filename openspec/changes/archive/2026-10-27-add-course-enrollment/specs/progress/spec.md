# Spec Delta

## MODIFIED Requirements

### Requirement: Strictly sequential progression
Progression SHALL be strictly sequential. Initially only the first activity SHALL be available. Every activity up to and including the current activity SHALL be available; every activity after it SHALL be locked.

#### Scenario: New course
- **WHEN** the learner enrolls in a course
- **THEN** only the first activity is available

#### Scenario: Locked activity
- **WHEN** the learner tries to open an activity after the current one
- **THEN** the app does not open it

### Requirement: Progress state
Progress SHALL consist of a pointer to the current activity and a set of completed activity slugs. There is no history of passes. Progress SHALL exist only for a course the learner is enrolled in; it SHALL be created empty at Enroll.

#### Scenario: Going back
- **WHEN** the learner reopens an earlier completed activity from the tree
- **THEN** the current activity does not change

#### Scenario: Course not enrolled
- **WHEN** a course is in the library and the learner has not enrolled in it
- **THEN** the course has no progress

### Requirement: Navigation through the tree
The learner SHALL be able to open any available activity through the course tree. Whenever entering a course opens the course screen, including on app start, the app SHALL open the current activity, not the one opened last.

#### Scenario: Restart after browsing back
- **WHEN** the learner opened an earlier activity, closed the app and started it again
- **THEN** the current activity is open
