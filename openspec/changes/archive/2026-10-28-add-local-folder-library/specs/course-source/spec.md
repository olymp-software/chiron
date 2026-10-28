# Spec Delta

## ADDED Requirements

### Requirement: Local-folder source normalization
The course source of a local-folder course SHALL be the absolute path of the folder with `.` and `..` resolved, symbolic links resolved and no trailing separator. On Windows two sources SHALL be the same course when they differ only in letter case; the source SHALL be stored as the learner's spelling of the first add.

#### Scenario: Same folder, different spelling
- **WHEN** on Windows the learner adds `c:\courses\go-basics\` while `C:\Courses\go-basics` is already in the library
- **THEN** the app treats it as the existing course, not a new one

#### Scenario: Folder behind a symbolic link
- **WHEN** the learner adds `D:\link-to-go`, a symbolic link to the already added `C:\Courses\go-basics`
- **THEN** the app treats it as the existing course

### Requirement: Storage key
Every course SHALL have a storage key derived only from its course source: a label from the last segment of the source, followed by `-` and a short hash of the source. It SHALL consist only of lowercase ASCII letters, digits and `-`. The app SHALL name everything it stores on disk for a course by the storage key, never by the course source. The library SHALL store the storage key next to the course source.

#### Scenario: Local-folder course
- **WHEN** the learner adds the folder `C:\Courses\Go Basics`
- **THEN** the storage key has the form `go-basics-<hash>` and contains no `:` or `\`

#### Scenario: Repository course
- **WHEN** the course source is `github.com/owner/repo`
- **THEN** the storage key has the form `repo-<hash>` and contains no `/`

#### Scenario: Stable across adds
- **WHEN** a course is added, removed and added again from the same source
- **THEN** it gets the same storage key

#### Scenario: Same last segment
- **WHEN** the learner adds `C:\a\course` and `C:\b\course`
- **THEN** the two courses have different storage keys

### Requirement: Course copy of a local-folder course
The app SHALL work from its own copy of a local-folder course, never from the folder itself. The copy SHALL contain the folder's files except `.git` and the files excluded by `.gitignore` files inside the folder, matched as git matches them. `.gitignore` files outside the folder and the author's global git settings SHALL NOT apply.

#### Scenario: Ignored build output
- **WHEN** the folder has `.gitignore` with `bin/` and a `basics/hello/bin/` folder
- **THEN** the course copy has no `basics/hello/bin/`

#### Scenario: Nested .gitignore with negation
- **WHEN** `basics/.gitignore` contains `*.log` and `!keep.log`
- **THEN** the course copy has `basics/keep.log` and no other `.log` files under `basics/`

#### Scenario: Git metadata
- **WHEN** the folder is a git repository
- **THEN** the course copy has no `.git`

#### Scenario: Folder removed after adding
- **WHEN** the author deletes or moves the folder after the course was added
- **THEN** the course still opens from its course copy

### Requirement: Sync
The course overview of a local-folder course SHALL show a Sync button in the place of the version selector. Sync SHALL stop the course's running scripts and replace the course copy with a fresh copy of the folder as a whole, so the copy is never left partially updated. The new copy MUST NOT be validated in advance. If the folder cannot be read, the course copy SHALL stay unchanged and the app SHALL show an error.

#### Scenario: Author edits the course
- **WHEN** the author edits `OVERVIEW.md` in the folder and the learner presses Sync
- **THEN** the course overview shows the edited text

#### Scenario: Synced manifest is broken
- **WHEN** the author breaks `course.yaml` and the learner presses Sync
- **THEN** the course copy holds the broken `course.yaml` and the course overview shows the course error

#### Scenario: Folder is gone
- **WHEN** the folder no longer exists and the learner presses Sync
- **THEN** the app shows an error and the course copy is unchanged

#### Scenario: Interrupted Sync
- **WHEN** the app is closed while Sync is copying
- **THEN** on the next start the course copy is either the previous one or the new one in full

## MODIFIED Requirements

### Requirement: Local-folder course
A course added by a local folder path SHALL have no tags; it includes the folder's uncommitted edits as of the last add or Sync. The app MUST NOT watch the folder. Edits SHALL be picked up only by Sync in the course overview. Repository courses have no Sync.

#### Scenario: Author edits an activity
- **WHEN** the author edits `activity.yaml` of an activity in the folder and the learner presses Sync
- **THEN** the course copy has the edited `activity.yaml`

#### Scenario: Edits without reload
- **WHEN** the author edits files of the folder without pressing Sync
- **THEN** the app keeps showing the previous state

#### Scenario: Folder is not a git repository
- **WHEN** the learner adds a local folder that is not a git repository
- **THEN** the course is added

#### Scenario: Uncommitted edits
- **WHEN** the folder is a git repository with uncommitted edits and the learner presses Sync
- **THEN** the course copy contains the uncommitted edits

### Requirement: Adding a course
A course SHALL be added by repository URL or from a local folder. There is no course catalog. On adding, the app SHALL warn once that the course will run scripts on the learner's computer with the learner's permissions. A course SHALL be added only if its `course.yaml` (for a repository course, at the version chosen when adding) is valid; otherwise the app SHALL show the course error and not add the course.

#### Scenario: Script warning
- **WHEN** the learner adds a course
- **THEN** the app shows the script warning and adds the course only after confirmation

#### Scenario: No repeated warning on version switch
- **WHEN** the learner switches an already added course to another version
- **THEN** no script warning is shown

#### Scenario: No repeated warning on Sync
- **WHEN** the learner presses Sync on an already added local-folder course
- **THEN** no script warning is shown

#### Scenario: Folder without a valid manifest
- **WHEN** the learner adds a folder that has no `course.yaml` or an invalid one
- **THEN** the app shows the course error and the course is not in the library

#### Scenario: Course already added
- **WHEN** the learner adds a source that is already in the library
- **THEN** no second entry is created and the course overview of the existing course opens
