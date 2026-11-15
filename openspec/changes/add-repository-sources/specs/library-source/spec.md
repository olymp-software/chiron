## MODIFIED Requirements

### Requirement: Course source identifies a course
A course SHALL be identified by its course source: the normalized repository address, or for a local-folder course, the folder path. The course manifest has no id field. Different sources SHALL be different courses with separate progress.

#### Scenario: Repository URL normalization
- **WHEN** the learner adds a course from `HTTPS://GitHub.com/Owner/Repo.git/`
- **THEN** the course source is `https://github.com/Owner/Repo.git`: scheme and host are lowercased and the trailing slash is dropped

#### Scenario: Same course, different spelling
- **WHEN** the learner adds `https://GitHub.com/Owner/Repo.git/` while `https://github.com/owner/repo` is already in the library
- **THEN** the app treats it as the existing course, not a new one

#### Scenario: Shorthand of an added course
- **WHEN** the learner adds `github.com/owner/repo` while `https://github.com/owner/repo` is already in the library
- **THEN** the app treats it as the existing course

#### Scenario: A fork is a different course
- **WHEN** the learner adds `github.com/other/repo`, a fork of the already added `github.com/owner/repo`
- **THEN** it is a separate course with empty progress

#### Scenario: https and ssh are different courses
- **WHEN** the learner adds `git@github.com:owner/repo.git` while `https://github.com/owner/repo` is already in the library
- **THEN** it is a separate course with empty progress

### Requirement: Course version is a vX.Y.Z git tag
A course version SHALL be a git tag of the form `vX.Y.Z`, where X, Y and Z are non-negative integers. Other tags, including pre-release tags such as `v2.0.0-beta.1`, SHALL be ignored. A repository without `vX.Y.Z` tags MUST NOT be added as a course.

#### Scenario: Repository without tags
- **WHEN** the learner adds a repository that has no `vX.Y.Z` tags
- **THEN** the course is not added and the app shows an error

#### Scenario: Version choice when adding
- **WHEN** the learner adds a repository with tags `v1.0.0`, `v1.2.0`, `v1.10.0`
- **THEN** the app lists the versions in descending order with `v1.10.0` preselected

#### Scenario: Pre-release and other tags
- **WHEN** the repository has tags `v1.0.0`, `v2.0.0-beta.1` and `release-3`
- **THEN** only `v1.0.0` is listed

### Requirement: Only the learner changes the selected version
The learner SHALL choose the course version in the course overview, where the locally known tags are listed in descending version order. Switching SHALL work in both directions. The app MUST NOT change the selected version on its own.

#### Scenario: A newer version is released
- **WHEN** "Check for new versions" finds a tag newer than the selected version
- **THEN** the selected version stays the same, the new tag is listed in the version selector and the course is marked in the library as not on the latest version

#### Scenario: Switching to an older version
- **WHEN** the learner picks a version older than the selected one in the course overview
- **THEN** the course overview shows the picked version

#### Scenario: No version selector on the course screen
- **WHEN** the learner is on the course screen
- **THEN** there is no control that changes the selected version

### Requirement: The app owns the course clone
The course copy of a repository course SHALL be a full clone of the repository. The selected version SHALL be the `vX.Y.Z` tag the clone is checked out at; it is not stored separately. When several `vX.Y.Z` tags point at the checked-out commit, the highest one SHALL be the selected version. On a version switch the app SHALL check out the chosen tag and discard any local changes and untracked files in the clone.

#### Scenario: Local changes in the clone
- **WHEN** the clone has local changes and the learner switches versions
- **THEN** the clone is at the chosen tag and the local changes are gone

#### Scenario: Two tags on one commit
- **WHEN** `v1.0.0` and `v1.0.1` point at the same commit and the learner picks `v1.0.0`
- **THEN** the course overview shows `v1.0.1` as the selected version and the course is not marked as not on the latest version because of `v1.0.1`

#### Scenario: Switching offline
- **WHEN** the learner switches to another locally known version without network access
- **THEN** the switch succeeds

### Requirement: Version switch procedure
Before switching versions the app SHALL stop the course's running scripts. The target version MUST NOT be validated in advance. Learner folders SHALL be left untouched by the switch.

#### Scenario: Target version has a broken manifest
- **WHEN** the learner switches to a version whose `course.yaml` is invalid
- **THEN** the course overview shows a course error, and the learner can switch to another version from the same list

#### Scenario: A control is running
- **WHEN** the learner switches versions while a control script is running
- **THEN** the script is terminated before the switch

#### Scenario: Interrupted switch
- **WHEN** the app is closed while a version switch is checking out files and the learner starts it again
- **THEN** the clone holds exactly the files of the version it shows as selected

### Requirement: Adding a course
A course SHALL be added by repository address or from a local folder. There is no course catalog. On adding, the app SHALL warn once that the course will run scripts on the learner's computer with the learner's permissions. A course SHALL be added only if its `course.yaml` (for a repository course, at the version chosen when adding) is valid; otherwise the app SHALL show the course error and not add the course.

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

#### Scenario: Repository version without a valid manifest
- **WHEN** the learner adds a repository and the chosen version has no `course.yaml` or an invalid one
- **THEN** the app shows the course error, the course is not in the library and nothing of the clone is left on disk

#### Scenario: Course already added
- **WHEN** the learner adds a source that is already in the library
- **THEN** no second entry is created and the course overview of the existing course opens

### Requirement: Storage key
Every course SHALL have a storage key derived only from its course source: a label from the last segment of the source, followed by `-` and a short hash of the source. For a repository course a trailing `.git` SHALL NOT be part of the label. It SHALL consist only of lowercase ASCII letters, digits and `-`. The app SHALL name everything it stores on disk for a course by the storage key, never by the course source. The library SHALL store the storage key next to the course source.

#### Scenario: Local-folder course
- **WHEN** the learner adds the folder `C:\Courses\Go Basics`
- **THEN** the storage key has the form `go-basics-<hash>` and contains no `:` or `\`

#### Scenario: Repository course
- **WHEN** the course source is `https://github.com/owner/repo.git`
- **THEN** the storage key has the form `repo-<hash>` and contains no `/`

#### Scenario: Same course, same key
- **WHEN** the same repository is added as `https://github.com/Owner/Repo.git` on one machine and as `github.com/owner/repo` on another
- **THEN** both get the same storage key

#### Scenario: Stable across adds
- **WHEN** a course is added, removed and added again from the same source
- **THEN** it gets the same storage key

#### Scenario: Same last segment
- **WHEN** the learner adds `C:\a\course` and `C:\b\course`
- **THEN** the two courses have different storage keys

## ADDED Requirements

### Requirement: Repository addresses
The app SHALL accept a repository address as `https://host/path`, `ssh://[user@]host[:port]/path`, scp-like `user@host:path`, or the shorthand `github.com/owner/repo`. Any other form, including `http://`, `git://` and `file://`, SHALL be rejected with an error before any network access.

#### Scenario: GitHub shorthand
- **WHEN** the learner enters `github.com/owner/repo`
- **THEN** the app uses `https://github.com/owner/repo`

#### Scenario: scp-like ssh address
- **WHEN** the learner enters `git@gitlab.example.com:team/course.git`
- **THEN** the app uses `ssh://git@gitlab.example.com/team/course.git`

#### Scenario: Plain http
- **WHEN** the learner enters `http://example.com/course.git`
- **THEN** the app shows an invalid address error and does not contact the host

### Requirement: Repository source normalization
The course source of a repository course SHALL be the address in its full form with the scheme and host lowercased and a trailing `/` removed; the rest SHALL be stored as the learner spelled it on the first add and used to clone. Two repository sources SHALL be the same course when they are equal after also lowercasing the path and removing a trailing `.git`.

#### Scenario: Stored spelling
- **WHEN** the learner first adds `HTTPS://Example.com/Team/Course.git/`
- **THEN** the course source is `https://example.com/Team/Course.git`

#### Scenario: Different spelling of an added repository
- **WHEN** the learner adds `https://example.com/team/course` while `https://example.com/Team/Course.git` is in the library
- **THEN** the app treats it as the existing course

### Requirement: System git
Repository courses SHALL use the git installed on the learner's machine. Authentication SHALL be left to git and its configuration: credential helpers for https, ssh-agent or keys for ssh. The app SHALL NOT store credentials. Local-folder courses SHALL NOT require git.

#### Scenario: Private repository over https
- **WHEN** the learner adds a private repository over https and Git Credential Manager is installed
- **THEN** the sign-in window of Git Credential Manager may appear and, after signing in, the course is added

#### Scenario: No git on the machine
- **WHEN** git is not installed and the learner adds a repository course
- **THEN** the app shows an error saying git is required, and local-folder courses keep working

#### Scenario: Access denied
- **WHEN** the learner adds a repository they have no access to
- **THEN** the course is not added and the app shows an access error

### Requirement: Git never waits for input
The app SHALL run git so that it never waits for input in a terminal: terminal prompts SHALL be disabled and ssh SHALL run in batch mode. Graphical credential prompts of the learner's credential helper SHALL still work. A git operation that would need terminal input SHALL fail with an error instead.

#### Scenario: ssh key with a passphrase and no agent
- **WHEN** the learner adds an ssh address, the key has a passphrase and no ssh-agent holds it
- **THEN** the app shows an access error without waiting

#### Scenario: Unknown ssh host
- **WHEN** the learner adds an ssh address whose host is not in `known_hosts`
- **THEN** the app shows an error without waiting for a host key confirmation

#### Scenario: https without a credential helper
- **WHEN** a private https repository is added on a machine without a credential helper
- **THEN** the app shows an access error without waiting

### Requirement: Course copy content of a repository course
The clone SHALL contain the repository files exactly as committed: the learner's git settings for line-ending conversion and hooks SHALL NOT apply, Git LFS SHALL NOT download content, and submodules SHALL NOT be fetched.

#### Scenario: Line endings on Windows
- **WHEN** the learner's git has `core.autocrlf=true` and the course has `check.sh` committed with LF line endings
- **THEN** `check.sh` in the clone has LF line endings

#### Scenario: Git LFS file
- **WHEN** a course file is stored with Git LFS
- **THEN** the clone holds the LFS pointer file, whether or not git-lfs is installed

#### Scenario: Submodule
- **WHEN** the repository has a submodule
- **THEN** the submodule folder in the clone is empty

### Requirement: Adding a repository course
Adding by repository address SHALL first list the repository's versions without downloading it, then let the learner pick a version and confirm the script warning, then clone the repository and check out the picked version. The clone SHALL be cancellable while it runs. A cancelled or failed add SHALL leave nothing on disk and nothing in the library.

#### Scenario: Wrong address
- **WHEN** the learner enters the address of a repository that does not exist
- **THEN** the app shows an error before the version choice

#### Scenario: Already added
- **WHEN** the learner enters an address of a course already in the library
- **THEN** the course overview of the existing course opens without contacting the repository

#### Scenario: Cancelled clone
- **WHEN** the learner cancels the add while the repository is being cloned
- **THEN** git is stopped, the course is not in the library and no clone is left on disk

#### Scenario: App closed during the clone
- **WHEN** the app is closed while a repository is being cloned and started again
- **THEN** the course is not in the library and no clone is left on disk

### Requirement: Check for new versions
The course overview of a repository course SHALL have a "Check for new versions" button next to the version selector. It SHALL fetch the repository's `vX.Y.Z` tags that are not known locally and SHALL NOT change the selected version. The app SHALL NOT contact the repository of an added course in any other way. A failed check SHALL show an error and leave the course unchanged.

#### Scenario: New version released
- **WHEN** the Author has published `v1.3.0`, the learner is on `v1.2.0` and presses "Check for new versions"
- **THEN** `v1.3.0` appears in the version selector and `v1.2.0` stays selected

#### Scenario: Nothing new
- **WHEN** the learner presses "Check for new versions" and the repository has no new tags
- **THEN** the version list stays the same and the app shows that no new versions were found

#### Scenario: Offline
- **WHEN** the learner presses "Check for new versions" without network access
- **THEN** the app shows an error and the course keeps working from the clone

#### Scenario: Opening a course
- **WHEN** the learner opens the course overview or the course screen of a repository course
- **THEN** the app does not contact the repository

### Requirement: Known tags are never replaced
A tag the app already knows SHALL keep the commit it had when the app first fetched it, even if the Author later moves the tag in the repository.

#### Scenario: Author moves a tag
- **WHEN** the Author force-pushes `v1.2.0` to another commit and the learner presses "Check for new versions"
- **THEN** `v1.2.0` in the clone still points at the commit it pointed at before and the course content is unchanged

### Requirement: Interrupted version switch is repaired
On startup the app SHALL bring the clone of every repository course back to its checked-out commit, discarding local changes and untracked files. If git is not available, the repair SHALL be skipped and logged.

#### Scenario: Files from two versions
- **WHEN** a version switch was interrupted and left files of two versions in the clone
- **THEN** after the next start the clone holds exactly the files of its checked-out commit

## REMOVED Requirements

### Requirement: Public https repositories only
**Reason**: The app now uses the system git, so private repositories over https and ssh work through git's own authentication, and git on the learner's machine is required for repository courses.
**Migration**: Covered by "Repository addresses", "System git" and "Git never waits for input". No stored data is affected: no repository course could be added before.

### Requirement: Tags are fetched when a course opens
**Reason**: Fetching on every open contacts the repository, and for a private repository may show a sign-in window, without the learner asking for it.
**Migration**: New versions are found only with the "Check for new versions" button in the course overview (requirement "Check for new versions"); a course always opens from its clone.
