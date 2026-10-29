## MODIFIED Requirements

### Requirement: Course overview contents
The course overview SHALL show the course title, the course source, the storage key, the overview text, the selected version with the version selector for a repository course or Sync for a local-folder course, the environment check controls if `check` is declared, and an Enroll or Continue button. If `course.yaml` is invalid, the overview SHALL show the course error and keep the version selector or Sync available.

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
