## MODIFIED Requirements

### Requirement: Removing a course
Removing a course SHALL by default delete its library entry, enrollment, course copy, progress, remembered environment check result and learner folders, or the workshop folder of a workshop. The confirmation dialog SHALL offer two independent checkboxes, unchecked by default: "Keep progress" (keeps the enrollment and progress) and "Keep my files". The remembered check result SHALL always be deleted, because the learner's machine may change between removing and re-adding the course. Before removal the app SHALL stop the course's running scripts.

#### Scenario: Remove everything
- **WHEN** the learner removes a course without checking either box
- **THEN** the library entry, enrollment, course copy, progress and learner folders of the course are deleted

#### Scenario: Remove a workshop
- **WHEN** the learner removes a workshop without checking either box
- **THEN** the library entry, enrollment, course copy, progress and workshop folder of the workshop are deleted

#### Scenario: Keep my files of a workshop
- **WHEN** the learner removes a workshop with "Keep my files" checked
- **THEN** the workshop folder and its content stay on disk unchanged

#### Scenario: Re-adding with kept data
- **WHEN** the learner removed a course with "Keep progress" and "Keep my files" checked and later adds the same source
- **THEN** the learner is enrolled again, the course overview shows Continue, and the previous progress and learner folders are picked up

#### Scenario: Re-adding without kept progress
- **WHEN** the learner removed a course without "Keep progress" and later adds the same source
- **THEN** the learner is not enrolled and the course overview shows Enroll

#### Scenario: Re-adding a course with a check
- **WHEN** the learner removed a course whose check was satisfied, with "Keep progress" checked, and later adds the same source
- **THEN** the check is not satisfied and Continue is unavailable until the check exits with code 0

#### Scenario: Workshop control running
- **WHEN** a workshop control started from the course overview is still running and the learner removes the workshop
- **THEN** the control is stopped before anything is deleted

## ADDED Requirements

### Requirement: Remove from the course overview
The course overview of every library item SHALL have a Remove button that opens the removal confirmation dialog. The library list SHALL NOT offer removal. After a confirmed removal the app SHALL open the library, replacing the course overview in the history.

#### Scenario: Remove button
- **WHEN** the learner opens the course overview of a course or a workshop, enrolled or not, valid or with a course error
- **THEN** the overview shows a Remove button

#### Scenario: Cancelled
- **WHEN** the learner presses Remove and then cancels the dialog
- **THEN** nothing is deleted and the course overview stays open

#### Scenario: After removal
- **WHEN** the learner confirms the removal
- **THEN** the library opens without the removed library item

#### Scenario: Back after removal
- **WHEN** the learner removed a library item and presses Back in the library
- **THEN** the course overview of the removed library item does not open

#### Scenario: Not in the library list
- **WHEN** the learner looks at a library item in the library
- **THEN** there is no control that removes it

### Requirement: Files left after removal
A library item SHALL be removed once its library entry is gone, even if some of its files cannot be deleted. The app SHALL then show a warning listing the paths left on disk.

#### Scenario: Workshop folder open in another program
- **WHEN** on Windows the learner removes a workshop while a terminal has its workshop folder as working directory
- **THEN** the workshop is no longer in the library and the app shows a warning with the path of the workshop folder

#### Scenario: Everything deleted
- **WHEN** every file of the removed library item is deleted
- **THEN** no warning is shown

### Requirement: Leftovers of a removal
On startup the app SHALL delete every course copy and `TEMP_DIR` that belongs to no library item. It SHALL NOT delete progress, learner folders or workshop folders of library items no longer in the library, since they may have been kept on purpose.

#### Scenario: Course copy left behind
- **WHEN** a removal could not delete the course copy and the learner restarts the app
- **THEN** the course copy is deleted on startup

#### Scenario: Kept progress survives restart
- **WHEN** the learner removed a course with "Keep progress" checked and restarts the app
- **THEN** the progress is still on disk, and adding the same source again enrolls the learner with it
