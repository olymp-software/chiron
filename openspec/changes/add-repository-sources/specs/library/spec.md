## MODIFIED Requirements

### Requirement: Library lists added courses
The library SHALL list every added course, whether or not the learner is enrolled in it. Each entry SHALL show the stored course title, whether the learner is enrolled or has completed the course, and, for a repository course, whether the selected version is not the latest. Whether a version is the latest SHALL be decided from the tags known locally, without contacting the repository. Clicking an entry SHALL open its course overview.

#### Scenario: Course added but not enrolled
- **WHEN** the learner has added a course and has not enrolled in it
- **THEN** the course is listed in the library without the enrolled mark

#### Scenario: Completed course
- **WHEN** the learner has pressed "Finish course" in a course and opens the library
- **THEN** the course is marked as completed instead of enrolled

#### Scenario: Not on the latest version
- **WHEN** a repository course has a locally known tag newer than its selected version
- **THEN** its library entry is marked as not on the latest version

#### Scenario: Newer tag not yet fetched
- **WHEN** the Author has published a newer tag and the learner has not pressed "Check for new versions" since
- **THEN** the library entry is not marked as not on the latest version

#### Scenario: Open from the library
- **WHEN** the learner clicks a course in the library
- **THEN** the course overview of that course opens

### Requirement: Course overview contents
The course overview SHALL show the course title, the course source, the storage key, the overview text, the selected version with the version selector and "Check for new versions" for a repository course or Sync for a local-folder course, the environment check controls if `check` is declared, whether the course is completed, and an Enroll, Continue or Review button. If `course.yaml` is invalid, the overview SHALL show the course error and keep the version selector, "Check for new versions" or Sync available.

#### Scenario: Course with a check
- **WHEN** the learner opens the overview of a course whose `course.yaml` declares `check`
- **THEN** the overview shows the "Check environment" button and the check status, and the check output is shown only in the check modal

#### Scenario: Course without a check
- **WHEN** the learner opens the overview of a course without `check`
- **THEN** no environment check controls are shown

#### Scenario: Repository course
- **WHEN** the learner opens the overview of a repository course
- **THEN** the overview shows the selected version, the version selector with the locally known versions and the "Check for new versions" button, and no Sync button

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

#### Scenario: Version switch adds activities to a completed course
- **WHEN** the learner switches a completed repository course to a version with a new activity
- **THEN** the overview no longer shows the course as completed and shows Continue instead of Review
