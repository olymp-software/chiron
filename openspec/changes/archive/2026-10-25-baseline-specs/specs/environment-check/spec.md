# Spec Delta

## Purpose

Defines the course-level environment check: a script that verifies the learner's machine has everything the course needs, and that gates starting the course until it succeeds.

## ADDED Requirements

### Requirement: Check script declaration
A course MAY declare an environment check via `check` in `course.yaml`, a path from the course root. If `check` is absent, the course SHALL start immediately.

```sh
# scripts/check.sh
kubectl version --client || {
  echo "kubectl not found. Install it: https://kubernetes.io/docs/tasks/tools/"
  exit 1
}
```

#### Scenario: No check declared
- **WHEN** `course.yaml` has no `check`
- **THEN** the learner can open the first activity right after adding the course

### Requirement: Check run environment
The check SHALL run under the same contract as control scripts (interpreter, guaranteed commands, output, exit code, Stop), except that its working directory SHALL be the course root and `ACTIVITY_DIR` and `WORKSPACE_DIR` SHALL NOT be set. `HOST_OS`, `HOST_ARCH` and the inherited environment SHALL be present.

#### Scenario: Working directory
- **WHEN** the check runs `ls course.yaml`
- **THEN** the command succeeds

#### Scenario: No activity variables
- **WHEN** the check prints `$WORKSPACE_DIR`
- **THEN** the value is empty

### Requirement: Check gates the course
If `check` is declared, the course MUST NOT be started until the check exits with code 0. The course screen SHALL show a built-in "Check environment" button. The check's output SHALL explain to the learner what is missing and where to get it.

#### Scenario: Failed check
- **WHEN** the check exits with code 1
- **THEN** the output panel shows the script output and activities stay unavailable

#### Scenario: Successful check
- **WHEN** the check exits with code 0
- **THEN** the course can be started

### Requirement: Remembered by file hash
A successful check SHALL be remembered together with the SHA-256 of the `check` file. On opening the course the app SHALL recompute the hash; if it differs, the check SHALL be required again. The learner SHALL be able to rerun the check manually at any time.

#### Scenario: Version with a changed check
- **WHEN** the learner switches to a course version whose `check` file differs
- **THEN** the course requires the check again before activities are available

#### Scenario: Unchanged check
- **WHEN** the learner reopens the course and the `check` file is unchanged
- **THEN** the check is not required again

### Requirement: Only the check file is hashed
Only the `check` file itself SHALL be hashed. If it calls other files, the author SHALL change the `check` file when those change, e.g. by bumping a `# rev 3` comment.

#### Scenario: Helper changed, check unchanged
- **WHEN** a helper script called by `check` changes but `check` itself does not
- **THEN** the previous successful check still counts
