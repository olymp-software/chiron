# course-library Specification

## Purpose

Defines the library of added courses, the course overview from which the learner checks the environment and enrolls, enrollment itself, and which screen the app opens on startup, after adding a course and on entering a course.

## Requirements

### Requirement: Library lists added courses
The library SHALL list every added course, whether or not the learner is enrolled in it. Each entry SHALL show the stored course title, whether the learner is enrolled, and, for a repository course, whether the selected version is not the latest. Clicking an entry SHALL open its course overview.

#### Scenario: Course added but not enrolled
- **WHEN** the learner has added a course and has not enrolled in it
- **THEN** the course is listed in the library without the enrolled mark

#### Scenario: Not on the latest version
- **WHEN** a repository course has a tag newer than its selected version
- **THEN** its library entry is marked as not on the latest version

#### Scenario: Open from the library
- **WHEN** the learner clicks a course in the library
- **THEN** the course overview of that course opens

### Requirement: Adding a course does not enroll
Adding a course SHALL put it in the library without enrolling the learner in it. After a course is added, the app SHALL open its course overview.

#### Scenario: Course just added
- **WHEN** the learner adds a course and confirms the script warning
- **THEN** the course overview of the new course opens with an Enroll button and no progress exists for the course

### Requirement: Course overview contents
The course overview SHALL show the course title, the course source, the storage key, the overview text, the selected version with the version selector for a repository course or Sync for a local-folder course, the environment check controls if `check` is declared, and an Enroll or Continue button. If `course.yaml` is invalid, the overview SHALL show the course error and keep the version selector or Sync available.

#### Scenario: Course with a check
- **WHEN** the learner opens the overview of a course whose `course.yaml` declares `check`
- **THEN** the overview shows the "Check environment" button, the check status and the output panel

#### Scenario: Course without a check
- **WHEN** the learner opens the overview of a course without `check`
- **THEN** no environment check controls are shown

#### Scenario: Broken manifest
- **WHEN** the selected version has an invalid `course.yaml`
- **THEN** the overview shows the course error, Enroll and Continue are unavailable, and the learner can pick another version

#### Scenario: Broken manifest in a local-folder course
- **WHEN** a local-folder course has an invalid `course.yaml` in its course copy
- **THEN** the overview shows the course error, Enroll and Continue are unavailable, and Sync is available

#### Scenario: Matching files on disk
- **WHEN** the learner opens the overview of a course
- **THEN** the overview shows the course source and the storage key used to name its files on disk

### Requirement: Overview text source
The overview text SHALL be the content of the overview file `OVERVIEW.md` in the course root, matched case-insensitively. Without that file it SHALL be the course `description`. Without both, no overview text SHALL be shown. More than one matching file SHALL be a course error.

#### Scenario: Overview file present
- **WHEN** the course root has `OVERVIEW.md` and `course.yaml` has a `description`
- **THEN** the overview shows the rendered `OVERVIEW.md`

#### Scenario: Lowercase file name
- **WHEN** the course root has `overview.md`
- **THEN** the overview shows its rendered content on every operating system

#### Scenario: Only description
- **WHEN** the course root has no overview file and `course.yaml` has a `description`
- **THEN** the overview shows the `description` as plain text

#### Scenario: Neither
- **WHEN** the course has no overview file and no `description`
- **THEN** the overview shows no overview text and no error

#### Scenario: Two overview files
- **WHEN** the course root has both `OVERVIEW.md` and `overview.md`
- **THEN** the overview shows a course error

### Requirement: Overview file rendering
The overview file SHALL be rendered with the same Markdown dialect, code highlighting, callouts and HTML restriction as activity Markdown. Images SHALL be referenced only by relative paths to files inside the course; remote images SHALL NOT be loaded.

#### Scenario: Script tag in the overview file
- **WHEN** `OVERVIEW.md` contains `<script>alert(1)</script>`
- **THEN** the text is shown literally and nothing executes

#### Scenario: Relative image
- **WHEN** `OVERVIEW.md` contains `![setup](images/setup.png)` and the file exists in the course
- **THEN** the image is shown

### Requirement: Enroll
A course overview of a course the learner is not enrolled in SHALL show an Enroll button. Enroll SHALL be available only if the course has no `check` or the environment check is satisfied. Enroll SHALL create the enrollment with empty progress and open the course screen on the first activity.

#### Scenario: Enroll without a check
- **WHEN** the course has no `check` and the learner presses Enroll
- **THEN** the learner is enrolled and the course screen opens on the first activity

#### Scenario: Check not run
- **WHEN** the course declares `check` and the check has not succeeded for the current `check` file
- **THEN** Enroll is unavailable

#### Scenario: Check succeeded
- **WHEN** the course declares `check` and the learner runs the check and it exits with code 0
- **THEN** Enroll becomes available

### Requirement: Continue
A course overview of a course the learner is enrolled in SHALL show a Continue button instead of Enroll. Continue SHALL be available only if the course has no `check` or the environment check is satisfied, and SHALL open the course screen on the current activity.

#### Scenario: Continue an enrolled course
- **WHEN** the learner is enrolled, the check is satisfied and the learner presses Continue
- **THEN** the course screen opens on the current activity

#### Scenario: Check file changed
- **WHEN** the learner is enrolled and the selected version has a `check` file different from the one that last succeeded
- **THEN** Continue is unavailable until the check exits with code 0, and the enrollment and progress are kept

### Requirement: No unenroll
The app SHALL NOT offer a way to end an enrollment other than removing the course.

#### Scenario: Enrolled course in the overview
- **WHEN** the learner opens the overview of an enrolled course
- **THEN** there is no control that ends the enrollment

### Requirement: Enrolled courses on the course screen
The course screen SHALL have a dropdown listing the courses the learner is enrolled in. Picking a course SHALL enter that course.

#### Scenario: Switch between enrolled courses
- **WHEN** the learner is enrolled in two courses and picks the other one in the dropdown
- **THEN** the app enters the picked course

#### Scenario: Not enrolled course
- **WHEN** a course is in the library but the learner is not enrolled in it
- **THEN** it is not listed in the dropdown

### Requirement: Leaving the course screen
The course screen SHALL have a Library button that opens the library and an Overview button that opens the course overview of the course on the screen.

#### Scenario: To the library
- **WHEN** the learner presses Library on the course screen
- **THEN** the library opens

#### Scenario: To the course overview
- **WHEN** the learner presses Overview on the course screen of a local-folder course
- **THEN** the course overview of that course opens with Sync available

### Requirement: Back navigation
The app SHALL keep the history of screens and activities the learner opened. Back (the mouse back button or Alt+Left) SHALL return to the previous entry. A screen the app opened in place of the one the learner asked for SHALL replace that entry instead of adding one. The Library button of the course overview SHALL open the library, wherever the overview was opened from.

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

#### Scenario: Library from the course overview
- **WHEN** the learner opens the course overview with Overview on the course screen and presses its Library button
- **THEN** the library opens

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

### Requirement: Stored course title
The library SHALL store the course title from `course.yaml` when the course is added. After a version switch or Sync, the stored title SHALL be replaced with the title from the new `course.yaml` if it is valid and kept otherwise. The library SHALL show the stored title without reading the course.

#### Scenario: Title changed by the author
- **WHEN** the author changes `title` in `course.yaml` of a local-folder course and the learner presses Sync
- **THEN** the library shows the new title

#### Scenario: Manifest broken after adding
- **WHEN** a Sync or version switch leaves the course with an invalid `course.yaml`
- **THEN** the library keeps showing the last valid title
