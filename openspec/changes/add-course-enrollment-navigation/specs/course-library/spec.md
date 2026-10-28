# Spec Delta

## ADDED Requirements

### Requirement: Leaving the course screen
The course screen SHALL have a Library button that opens the library and an Overview button that opens the course overview of the course on the screen.

#### Scenario: To the library
- **WHEN** the learner presses Library on the course screen
- **THEN** the library opens

#### Scenario: To the course overview
- **WHEN** the learner presses Overview on the course screen of a local-folder course
- **THEN** the course overview of that course opens with Sync available

### Requirement: Back navigation
The app SHALL keep the history of screens and activities the learner opened. Back (the mouse back button or Alt+Left) SHALL return to the previous entry. A screen the app opened in place of the one the learner asked for SHALL replace that entry instead of adding one. The back button of the course overview SHALL act as Back and open the library when there is no previous entry.

#### Scenario: Back to the previous activity
- **WHEN** the learner opens activity A, then activity B from the course tree, and presses Back
- **THEN** activity A opens

#### Scenario: Back after switching courses
- **WHEN** the learner switches from course A to course B in the dropdown and presses Back
- **THEN** the activity of course A that was open before the switch opens

#### Scenario: Back after Enroll
- **WHEN** the learner enrolls from the course overview and presses Back on the course screen
- **THEN** the course overview opens

#### Scenario: Redirect is not an entry
- **WHEN** the learner picks a course in the dropdown whose `course.yaml` became invalid, so its course overview opens, and presses Back
- **THEN** the course screen that was open before the pick opens

#### Scenario: Overview opened on startup
- **WHEN** the app opened a course overview on startup and the learner presses its back button
- **THEN** the library opens

## MODIFIED Requirements

### Requirement: Entering a course respects the check
Entering an enrolled course from the dropdown or on app start SHALL open the course screen only if the course's `course.yaml` is valid and the course has no `check` or the environment check is satisfied; otherwise the app SHALL open that course's overview.

#### Scenario: Version switched, check not rerun, app restarted
- **WHEN** the learner switched to a version with a changed `check` file, closed the app without running the check and started it again
- **THEN** the course overview of that course opens with Continue unavailable

#### Scenario: Author edits the check
- **WHEN** the author edits the `check` file of an enrolled local-folder course and the learner presses Sync in the course overview
- **THEN** Continue is unavailable until the check exits with code 0

#### Scenario: Manifest broken by Sync
- **WHEN** a Sync left an enrolled local-folder course with an invalid `course.yaml` and the learner picks it in the dropdown
- **THEN** the course overview of that course opens with the course error

### Requirement: Startup screen
On startup the app SHALL enter the enrolled course entered last. If no course was entered yet, or the course entered last is no longer in the library or not enrolled, the app SHALL open the library.

#### Scenario: Restart
- **WHEN** the learner closes the app with an enrolled course open and starts it again
- **THEN** the app enters the same course

#### Scenario: No enrollments
- **WHEN** the learner has added courses but is enrolled in none and starts the app
- **THEN** the library opens

#### Scenario: Course entered last was removed
- **WHEN** the learner removed the course entered last, is still enrolled in another course and starts the app
- **THEN** the library opens
