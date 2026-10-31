# Spec Delta

## MODIFIED Requirements

### Requirement: Library lists added courses
The library SHALL list every added course, whether or not the learner is enrolled in it. Each entry SHALL show the stored course title, whether the learner is enrolled or has completed the course, and, for a repository course, whether the selected version is not the latest. Clicking an entry SHALL open its course overview.

#### Scenario: Course added but not enrolled
- **WHEN** the learner has added a course and has not enrolled in it
- **THEN** the course is listed in the library without the enrolled mark

#### Scenario: Completed course
- **WHEN** the learner has pressed "Finish course" in a course and opens the library
- **THEN** the course is marked as completed instead of enrolled

#### Scenario: Not on the latest version
- **WHEN** a repository course has a tag newer than its selected version
- **THEN** its library entry is marked as not on the latest version

#### Scenario: Open from the library
- **WHEN** the learner clicks a course in the library
- **THEN** the course overview of that course opens

### Requirement: Course overview contents
The course overview SHALL show the course title, the course source, the storage key, the overview text, the selected version with the version selector for a repository course or Sync for a local-folder course, the environment check controls if `check` is declared, whether the course is completed, and an Enroll, Continue or Review button. If `course.yaml` is invalid, the overview SHALL show the course error and keep the version selector or Sync available.

#### Scenario: Course with a check
- **WHEN** the learner opens the overview of a course whose `course.yaml` declares `check`
- **THEN** the overview shows the "Check environment" button and the check status, and the check output is shown only in the check modal

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

#### Scenario: Completed course
- **WHEN** the learner opens the overview of a completed course
- **THEN** the overview shows that the course is completed

#### Scenario: Sync adds activities to a completed course
- **WHEN** the learner presses Sync in the overview of a completed local-folder course whose new `course.yaml` has a new activity anywhere in the course
- **THEN** the overview no longer shows the course as completed and shows Continue instead of Review

### Requirement: Continue
A course overview of a course the learner is enrolled in SHALL show a Continue button instead of Enroll; for a completed course the button SHALL be Review instead of Continue. Continue and Review SHALL be available only if the course has no `check` or the environment check is satisfied, and SHALL open the course screen on the current activity.

#### Scenario: Continue an enrolled course
- **WHEN** the learner is enrolled, the check is satisfied and the learner presses Continue
- **THEN** the course screen opens on the current activity

#### Scenario: Review a completed course
- **WHEN** the course is completed, the check is satisfied and the learner presses Review
- **THEN** the course screen opens on the last activity, which shows that the course is completed

#### Scenario: Check file changed
- **WHEN** the learner is enrolled and the selected version has a `check` file different from the one that last succeeded
- **THEN** Continue is unavailable until the check exits with code 0, and the enrollment and progress are kept
