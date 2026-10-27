# Spec Delta

## MODIFIED Requirements

### Requirement: Only the learner changes the selected version
The learner SHALL choose the course version in the course overview, where tags are listed in descending version order. Switching SHALL work in both directions. The app MUST NOT change the selected version on its own.

#### Scenario: A newer version is released
- **WHEN** the course gets a tag newer than the selected version
- **THEN** the selected version stays the same and the course is marked in the library as not on the latest version

#### Scenario: Switching to an older version
- **WHEN** the learner picks a version older than the selected one in the course overview
- **THEN** the course overview shows the picked version

#### Scenario: No version selector on the course screen
- **WHEN** the learner is on the course screen
- **THEN** there is no control that changes the selected version

### Requirement: Version switch procedure
Before switching versions the app SHALL stop the course's running scripts. The target version MUST NOT be validated in advance. Learner folders SHALL be left untouched by the switch.

#### Scenario: Target version has a broken manifest
- **WHEN** the learner switches to a version whose `course.yaml` is invalid
- **THEN** the course overview shows a course error, and the learner can switch to another version from the same list

#### Scenario: A control is running
- **WHEN** the learner switches versions while a control script is running
- **THEN** the script is terminated before the switch

### Requirement: Removing a course
Removing a course SHALL by default delete its library entry, enrollment, clone, progress, remembered environment check result and learner folders. The confirmation dialog SHALL offer two independent checkboxes, unchecked by default: "Keep progress" (keeps the enrollment, progress and remembered check result) and "Keep my files". Before removal the app SHALL stop the course's running scripts.

#### Scenario: Remove everything
- **WHEN** the learner removes a course without checking either box
- **THEN** the library entry, enrollment, clone, progress and learner folders of the course are deleted

#### Scenario: Re-adding with kept data
- **WHEN** the learner removed a course with "Keep progress" and "Keep my files" checked and later adds the same source
- **THEN** the learner is enrolled again, the course overview shows Continue, and the previous progress and learner folders are picked up

#### Scenario: Re-adding without kept progress
- **WHEN** the learner removed a course without "Keep progress" and later adds the same source
- **THEN** the learner is not enrolled and the course overview shows Enroll

## REMOVED Requirements

### Requirement: Course library
**Reason**: The library and the startup behavior are now owned by the `course-library` capability, which also introduces enrollment: the app no longer opens the course open last but the enrolled course entered last.
**Migration**: See `course-library` requirements "Library lists added courses" and "Startup screen".
