# Spec Delta

## ADDED Requirements

### Requirement: Stored course title
The library SHALL store the course title from `course.yaml` when the course is added. After a version switch or Sync, the stored title SHALL be replaced with the title from the new `course.yaml` if it is valid and kept otherwise. The library SHALL show the stored title without reading the course.

#### Scenario: Title changed by the author
- **WHEN** the author changes `title` in `course.yaml` of a local-folder course and the learner presses Sync
- **THEN** the library shows the new title

#### Scenario: Manifest broken after adding
- **WHEN** a Sync or version switch leaves the course with an invalid `course.yaml`
- **THEN** the library keeps showing the last valid title

## MODIFIED Requirements

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

### Requirement: Entering a course respects the check
Entering an enrolled course from the dropdown or on app start SHALL open the course screen only if the course has no `check` or the environment check is satisfied; otherwise the app SHALL open that course's overview.

#### Scenario: Version switched, check not rerun, app restarted
- **WHEN** the learner switched to a version with a changed `check` file, closed the app without running the check and started it again
- **THEN** the course overview of that course opens with Continue unavailable

#### Scenario: Author edits the check
- **WHEN** the author edits the `check` file of an enrolled local-folder course and the learner presses Sync in the course overview
- **THEN** Continue is unavailable until the check exits with code 0
