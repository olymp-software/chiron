## MODIFIED Requirements

### Requirement: Check gates the course
If `check` is declared, Enroll and Continue MUST NOT be available until the check has succeeded for the current `check` file. The course overview SHALL show a built-in "Check environment" button and the check status. The check's output SHALL explain to the learner what is missing and where to get it.

#### Scenario: Failed check
- **WHEN** the check exits with code 1
- **THEN** the check modal shows the script output and, after the learner closes it, Enroll or Continue stays unavailable

#### Scenario: Successful check
- **WHEN** the check exits with code 0
- **THEN** Enroll or Continue becomes available

#### Scenario: No check button on the course screen
- **WHEN** the learner is on the course screen
- **THEN** the "Check environment" button is not shown there

### Requirement: Remembered by file hash
A successful check SHALL be remembered in the course's library entry together with the SHA-256 of the `check` file taken when the run started. On opening the course overview and on entering a course the app SHALL recompute the hash; if it differs, the check SHALL be required again. The learner SHALL be able to rerun the check at any time from the course overview, also when it is satisfied.

#### Scenario: Version with a changed check
- **WHEN** the learner switches to a course version whose `check` file differs
- **THEN** the course overview requires the check again before Enroll or Continue is available

#### Scenario: Unchanged check
- **WHEN** the learner reopens the course and the `check` file is unchanged
- **THEN** the check is not required again

#### Scenario: Required check survives a restart
- **WHEN** the check became required after a version switch and the learner restarts the app without running it
- **THEN** the check is still required

#### Scenario: Rerun a satisfied check
- **WHEN** the check is satisfied and the learner presses "Check environment"
- **THEN** the check runs again

## ADDED Requirements

### Requirement: Check modal
Pressing "Check environment" SHALL open a modal that shows the output panel of the run and a Cancel button, and blocks the rest of the window while it is open. Cancel SHALL stop the check the way Stop stops a script. When the check exits, the modal SHALL show the exit code with a success or failure mark and Cancel SHALL become OK. OK SHALL close the modal; the output is not kept after that.

#### Scenario: Output while running
- **WHEN** the check prints lines over several seconds
- **THEN** the modal shows each line as it arrives

#### Scenario: Cancel a running check
- **WHEN** the check is running and the learner presses Cancel
- **THEN** the check and its child processes are terminated

#### Scenario: Finished check
- **WHEN** the check exits with code 1
- **THEN** the modal shows code 1 with a failure mark and an OK button, and OK closes the modal

#### Scenario: Window blocked while checking
- **WHEN** the check modal is open
- **THEN** the learner cannot press Sync, Enroll, Continue or the Library button until the modal is closed

### Requirement: Result of a run
Exit code 0 SHALL remember the hash taken when the run started. A non-zero exit code SHALL forget the remembered result only if it was for that same hash; a result for another hash SHALL be kept. A run stopped by Cancel or by closing the app SHALL leave the remembered result unchanged.

#### Scenario: Learner broke a working environment
- **WHEN** the check is satisfied, the learner uninstalls a required tool, reruns the check and it exits with code 1
- **THEN** the check is no longer satisfied and Enroll or Continue becomes unavailable

#### Scenario: Failure after the check file changed
- **WHEN** the check succeeded for `check` file A, Sync brought `check` file B, the check exits with code 1, and a later Sync brings file A back
- **THEN** the check is satisfied again without a rerun

#### Scenario: Cancelled run
- **WHEN** the check is satisfied, the learner reruns it and presses Cancel
- **THEN** the check is still satisfied

### Requirement: Check status
The course overview of a course that declares `check` SHALL show the check status derived from the remembered result and the current `check` file: not checked (nothing remembered), passed (remembered for the current file), or check changed (remembered for another file). The result of the last failed run SHALL NOT be shown after the check modal is closed.

#### Scenario: Never checked
- **WHEN** the learner opens the overview of a course with `check` that was never checked
- **THEN** the status is "not checked"

#### Scenario: Check file changed
- **WHEN** the check succeeded and a Sync changed the `check` file
- **THEN** the status is "check changed"

#### Scenario: Passed
- **WHEN** the check succeeded and the `check` file is unchanged
- **THEN** the status is "passed"
