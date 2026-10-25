# course-source Specification

## Purpose

Defines where the app gets a course from, how it tells courses apart, how the learner picks a course version, and how courses are added to and removed from the library.

## Requirements

### Requirement: Course source identifies a course
A course SHALL be identified by its course source: the normalized repository URL, or for a local-folder course, the folder path. The course manifest has no id field. Different sources SHALL be different courses with separate progress.

#### Scenario: Repository URL normalization
- **WHEN** the learner adds a course from `https://GitHub.com/Owner/Repo.git/`
- **THEN** the course source is `github.com/owner/repo`: scheme, trailing `.git` and trailing slash are dropped and the result is lowercased

#### Scenario: Same course, different spelling
- **WHEN** the learner adds `https://github.com/owner/repo` while `github.com/owner/repo` is already in the library
- **THEN** the app treats it as the existing course, not a new one

#### Scenario: A fork is a different course
- **WHEN** the learner adds `github.com/other/repo`, a fork of the already added `github.com/owner/repo`
- **THEN** it is a separate course with empty progress

### Requirement: Public https repositories only
The app SHALL support courses from public git repositories over https without authentication. A system git installation MUST NOT be required.

#### Scenario: Repository requires authentication
- **WHEN** the learner adds a course whose repository requires authentication
- **THEN** the course is not added and the app shows an error

#### Scenario: No git on the machine
- **WHEN** git is not installed on the learner's machine
- **THEN** adding and switching versions of repository courses still works

### Requirement: Course version is a vX.Y.Z git tag
A course version SHALL be a git tag of the form `vX.Y.Z`. A repository without such tags MUST NOT be added as a course.

#### Scenario: Repository without tags
- **WHEN** the learner adds a repository that has no `vX.Y.Z` tags
- **THEN** the course is not added and the app shows an error

#### Scenario: Version choice when adding
- **WHEN** the learner adds a repository with tags `v1.0.0`, `v1.2.0`, `v1.10.0`
- **THEN** the app lists the versions in descending order with `v1.10.0` preselected

### Requirement: Only the learner changes the selected version
The learner SHALL choose the course version in the course library, where tags are listed in descending version order. Switching SHALL work in both directions. The app MUST NOT change the selected version on its own.

#### Scenario: A newer version is released
- **WHEN** the course gets a tag newer than the selected version
- **THEN** the selected version stays the same and the course is marked in the library as not on the latest version

#### Scenario: Switching to an older version
- **WHEN** the learner picks a version older than the selected one
- **THEN** the course opens on the picked version

### Requirement: Tags are fetched when a course opens
When a repository course opens, the app SHALL fetch its tag list. Without network access the fetch SHALL be skipped and the course SHALL work from the local clone.

#### Scenario: Offline
- **WHEN** the learner opens a repository course without network access
- **THEN** the course opens on the selected version from the clone, with no error

### Requirement: The app owns the course clone
The selected version SHALL be the tag the clone is checked out at; it is not stored separately. On a version switch the app SHALL check out the chosen tag and discard any local changes in the clone.

#### Scenario: Local changes in the clone
- **WHEN** the clone has local changes and the learner switches versions
- **THEN** the clone is at the chosen tag and the local changes are gone

### Requirement: Version switch procedure
Before switching versions the app SHALL stop the course's running scripts. The target version MUST NOT be validated in advance. Learner folders SHALL be left untouched by the switch.

#### Scenario: Target version has a broken manifest
- **WHEN** the learner switches to a version whose `course.yaml` is invalid
- **THEN** a course error is shown instead of the course tree, and the learner can switch to another version from the same list

#### Scenario: A control is running
- **WHEN** the learner switches versions while a control script is running
- **THEN** the script is terminated before the switch

### Requirement: Local-folder course
A course added by a local folder path SHALL be used as is: no tags, uncommitted edits included. The app MUST NOT watch the folder. Edits SHALL be picked up via a "Reload" button on the course screen that re-reads `course.yaml` and the open activity. Repository courses have no such button.

#### Scenario: Author edits an activity
- **WHEN** the author edits `activity.yaml` of the open activity in a local-folder course and presses "Reload"
- **THEN** the app shows the edited activity

#### Scenario: Edits without reload
- **WHEN** the author edits files of a local-folder course without pressing "Reload"
- **THEN** the app keeps showing the previous state

#### Scenario: Folder is not a git repository
- **WHEN** the learner adds a local folder that is not a git repository
- **THEN** the course is added

### Requirement: Adding a course
A course SHALL be added by repository URL or from a local folder. There is no course catalog. On adding, the app SHALL warn once that the course will run scripts on the learner's computer with the learner's permissions.

#### Scenario: Script warning
- **WHEN** the learner adds a course
- **THEN** the app shows the script warning and adds the course only after confirmation

#### Scenario: No repeated warning on version switch
- **WHEN** the learner switches an already added course to another version
- **THEN** no script warning is shown

### Requirement: Course library
Added courses SHALL be listed in the app's course library. On startup the app SHALL open the course that was open last.

#### Scenario: Restart
- **WHEN** the learner closes the app with a course open and starts it again
- **THEN** the same course opens

### Requirement: Removing a course
Removing a course SHALL by default delete its library entry, clone, progress and learner folders. The confirmation dialog SHALL offer two independent checkboxes, unchecked by default: "Keep progress" and "Keep my files". Before removal the app SHALL stop the course's running scripts.

#### Scenario: Remove everything
- **WHEN** the learner removes a course without checking either box
- **THEN** the library entry, clone, progress and learner folders of the course are deleted

#### Scenario: Re-adding with kept data
- **WHEN** the learner removed a course with "Keep progress" and "Keep my files" checked and later adds the same source
- **THEN** the previous progress and learner folders are picked up
