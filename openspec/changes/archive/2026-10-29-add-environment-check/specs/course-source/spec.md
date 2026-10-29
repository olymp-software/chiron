## MODIFIED Requirements

### Requirement: Removing a course
Removing a course SHALL by default delete its library entry, enrollment, clone, progress, remembered environment check result and learner folders. The confirmation dialog SHALL offer two independent checkboxes, unchecked by default: "Keep progress" (keeps the enrollment and progress) and "Keep my files". The remembered check result SHALL always be deleted, because the learner's machine may change between removing and re-adding the course. Before removal the app SHALL stop the course's running scripts.

#### Scenario: Remove everything
- **WHEN** the learner removes a course without checking either box
- **THEN** the library entry, enrollment, clone, progress and learner folders of the course are deleted

#### Scenario: Re-adding with kept data
- **WHEN** the learner removed a course with "Keep progress" and "Keep my files" checked and later adds the same source
- **THEN** the learner is enrolled again, the course overview shows Continue, and the previous progress and learner folders are picked up

#### Scenario: Re-adding without kept progress
- **WHEN** the learner removed a course without "Keep progress" and later adds the same source
- **THEN** the learner is not enrolled and the course overview shows Enroll

#### Scenario: Re-adding a course with a check
- **WHEN** the learner removed a course whose check was satisfied, with "Keep progress" checked, and later adds the same source
- **THEN** the check is not satisfied and Continue is unavailable until the check exits with code 0
